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
	logger.Info("client-run", logger.InProgress, messageArgs...)
	defer client.conn.Close()

	inputFile, outputFile, err := client.openFiles()
	if err != nil {
		logger.Error("create-output-file", logger.Fail, messageArgs...)
		return err
	}
	defer inputFile.Close()
	defer outputFile.Close()

	// enviar mensaje de hello
	if err := client.sendHello(); err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Error("send-hello", logger.Fail, messageArgs...)
		return err
	}
	logger.Info("client-run:sent-hello", logger.InProgress, messageArgs...)

	// enviar apuestas
	if err = client.sendBets(inputFile); err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Error("send-bets", logger.Fail, messageArgs...)
		return err
	}
	logger.Info("client-run:sent-bets", logger.InProgress, messageArgs...)

	// recibir ganadores
	if err = client.receiveWinners(outputFile); err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Error("receive-winners", logger.Fail, messageArgs...)
		return err
	}
	logger.Info("client-run:received-winners", logger.InProgress, messageArgs...)

	return nil
}

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

func (client *Client) sendHello() error {
	// enviar paquete de hello: inicio de conexion
	helloPacket := protocol.CreateHelloPacket(client.config.AgencyId, client.config.BatchSize)
	if err := safe_socket.SendAll(client.conn, helloPacket.Serialize()); err != nil {
		return err
	}
	ack, err := readPacket(client.conn)
	if err != nil {
		return err
	}
	// esperar por ack
	_, err = protocol.AckFromBytes(ack)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) sendBets(inputFile *os.File) error {
	scanner := bufio.NewScanner(inputFile)
	batchSize := int(client.config.BatchSize)
	logger.Info("client-run:sending-bets", logger.InProgress, "batch-size", batchSize)
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
		packet := protocol.CreateBetInfoPacket(bets)
		serializedPacket := packet.Serialize()
		if !keepScanning {
			protocol.SetLastPacketFlag(serializedPacket, protocol.LENGTH_BYTES)
		}
		if err := safe_socket.SendAll(client.conn, serializedPacket); err != nil {
			return err
		}
		// esperar ack del server
		ack, err := readPacket(client.conn)
		if err != nil {
			return err
		}
		if _, err = protocol.AckFromBytes(ack); err != nil {
			return err
		}

	}
	return nil
}

func (client *Client) receiveWinners(outputFile *os.File) error {
	keepReceiving := true
	for keepReceiving {
		responsePacket, err := readPacket(client.conn)
		if err != nil {
			return err
		}

		isLast := protocol.GetLastPacketFlag(responsePacket, 0)
		keepReceiving = !isLast
		switch responsePacket[0] {
		case protocol.TYPE_BET:
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
		case protocol.TYPE_ACK:
			// llego un ack en vez de winners: no hay winners
			// enviar ack y terminar
			ack := protocol.CreateAckPacket().Serialize()
			if err := safe_socket.SendAll(client.conn, ack); err != nil {
				return err
			}
		default:
			return errors.New("Unknown packet type")
		}
	}
	return nil
}
