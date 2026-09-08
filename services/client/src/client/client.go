package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   uint32
	BatchSize  uint8
	InputFile  string
	OutputFile string
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func (client *Client) Run() error {
	messageArgs := []any{"agency-id", client.config.AgencyId}
	action := "client-run"
	logger.Info(action, logger.InProgress, messageArgs...)
	defer client.conn.Close()

	inputFile, outputFile, err := client.openFiles()
	if err != nil {
		errArgs := []any{"err", err}
		logger.Error(action+":open-files", logger.Fail, append(messageArgs, errArgs...))
		return err
	}
	defer inputFile.Close()
	defer outputFile.Close()

	// enviar mensaje de hello
	if err := client.sendHello(); err != nil {
		errArgs := []any{"err", err}
		logger.Error(action+":send-hello", logger.Fail, append(messageArgs, errArgs...))
		return err
	}
	logger.Info(action+":hello-sent", logger.InProgress, messageArgs...)

	// enviar apuestas
	if err = client.sendBets(inputFile); err != nil {
		errArgs := []any{"err", err}
		logger.Error("send-bets", logger.Fail, append(messageArgs, errArgs...))
		return err
	}
	logger.Info(action+":sent-bets", logger.InProgress, messageArgs...)

	// recibir ganadores
	if err = client.receiveWinners(outputFile); err != nil {
		errArgs := []any{"err", err}
		logger.Error("receive-winners", logger.Fail, append(messageArgs, errArgs...))
		return err
	}
	logger.Info(action+":winners-received", logger.InProgress, messageArgs...)

	return nil
}

// Permite cerrar y limpiar correctamente los recursos del cliente
func (client *Client) GracefulShutdown() {
	client.conn.Close()
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

// Abre los archivos de entrada y salida, y los retorna en ese orden
func (client *Client) openFiles() (*os.File, *os.File, error) {
	// abrir archivo de entrada
	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		return nil, nil, err
	}

	// abrir archivo de salida
	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		return nil, nil, err
	}
	return inputFile, outputFile, nil
}

// Envia un paquete de tipo HELLO al servidor con la informacion del cliente
func (client *Client) sendHello() error {
	// enviar paquete de hello: inicio de conexion
	helloPacket := protocol.CreateHelloPacket(client.config.AgencyId, client.config.BatchSize)
	if err := safe_socket.SendAll(client.conn, helloPacket.Serialize()); err != nil {
		return err
	}
	// esperar por ack
	return client.waitForAck()
}

// Envia todas las apuestas presentes en el archivo de entrada. Si no todas pudieron enviarse
// correctamente retorna error
func (client *Client) sendBets(inputFile *os.File) error {
	batchSize := int(client.config.BatchSize)
	logger.Info("send-bets", logger.InProgress, "batch-size", batchSize)
	scanner := bufio.NewScanner(inputFile)
	keepScanning := scanner.Scan()
	csvBet := scanner.Text()
	for keepScanning {
		// recorrer bets para formar un batch
		bets := make([]lottery.Bet, 0, batchSize)
		for len(bets) < batchSize {
			bet, err := lottery.FromCsv(csvBet, client.config.AgencyId)
			if err != nil {
				return err
			}
			bets = append(bets, bet)
			keepScanning = scanner.Scan()
			if !keepScanning {
				break
			}
			csvBet = scanner.Text()

		}
		if err := scanner.Err(); err != nil {
			return err
		}
		if len(bets) == 0 {
			break
		}
		// ya se tiene batch, enviarlo
		if err := client.sendBatch(bets, !keepScanning); err != nil {
			return err
		}
		// esperar ack del server
		if err := client.waitForAck(); err != nil {
			return err
		}

	}
	return nil
}

// Espera por un mensaje de tipo ACK del servidor
func (client *Client) waitForAck() error {
	ack, err := readPacket(client.conn)
	if err != nil {
		return err
	}
	if _, err = protocol.AckFromBytes(ack); err != nil {
		return err
	}
	return nil
}

// Envia un batch de apuestas al servidor
func (client *Client) sendBatch(bets []lottery.Bet, isLast bool) error {
	packet := protocol.CreateBetInfoPacket(bets)
	serializedPacket := packet.Serialize()
	if isLast { // si es ultimo batch, setear flag de ultimo paquete
		protocol.SetLastPacketFlag(serializedPacket, protocol.LENGTH_BYTES)
	}
	return safe_socket.SendAll(client.conn, serializedPacket)
}

// Recibe los ganadores de la loteria. Si no todos fueron procesados correctamente,
// retorna error
func (client *Client) receiveWinners(outputFile *os.File) error {
	logger.Info("receive-winners", logger.InProgress)
	keepReceiving := true
	for keepReceiving {
		// leer paquete del server con winners
		responsePacket, err := readPacket(client.conn)
		if err != nil {
			return err
		}
		// ver tipo de paquete y parsear
		isLast := protocol.GetLastPacketFlag(responsePacket, 0)
		keepReceiving = !isLast
		switch responsePacket[0] {
		case protocol.TYPE_BET:
			// parsear y guardar bets ganadoras
			if err := client.betReceptionAction(responsePacket, outputFile); err != nil {
				return err
			}
		case protocol.TYPE_ACK:
			// llego un ack en vez de winners: no hay winners
			// enviar ack y terminar
			if err := client.sendAck(); err != nil {
				return err
			}
		default:
			return errors.New("Unknown packet type")
		}
	}
	return nil
}

// Recibe un batch del servidor con ganadores de la loteria
func (client *Client) betReceptionAction(responsePacket []byte, outputFile *os.File) error {
	// parsear y guardar bets ganadoras
	betInfo, err := protocol.BetInfoFromBytes(responsePacket, client.config.AgencyId, int(client.config.BatchSize))
	if err != nil {
		return err
	}
	for _, bet := range betInfo.Bets {
		if _, err = fmt.Fprintln(outputFile, bet.ToCsv()); err != nil {
			return err
		}
	}
	ack := protocol.CreateAckPacket().Serialize()
	if err := safe_socket.SendAll(client.conn, ack); err != nil {
		return err
	}
	return nil
}

// Envia un mensaje de tipo ACK al servidor
func (client *Client) sendAck() error {
	ack := protocol.CreateAckPacket().Serialize()
	if err := safe_socket.SendAll(client.conn, ack); err != nil {
		return err
	}
	return nil
}
