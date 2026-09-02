package client

import (
	"bufio"
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

func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close()

	// abrir archivo de entrada
	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
		return err
	}
	defer inputFile.Close()

	// abrir archivo de salida
	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error("create-output-file", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
		return err
	}
	defer outputFile.Close()

	// enviar mensaje de hello
	messageArgs := []any{"agency-id", client.config.AgencyId}
	helloPacket := protocol.CreateHelloPacket(client.config.AgencyId, client.config.BatchSize)
	if err := safe_socket.SendAll(client.conn, helloPacket.Serialize()); err != nil {
		logger.Error("send-hello", logger.Fail, messageArgs...)
		return err
	}

	// leer archivo linea a linea
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		logger.Info(mainAction, logger.InProgress, messageArgs...)

		csvBet := scanner.Text()
		bet, err := lottery.FromCsv(csvBet, client.config.AgencyId)
		if err != nil {
			logger.Error("parse-csv-bet", logger.InProgress, "agency-id", client.config.AgencyId, "err", err)
			continue // skippear linea e intentar con la siguiente (TODO: revisar!)
		}

		// enviar info de apuesta
		packet := protocol.CreateBetInfoPacket(bet)
		if err := safe_socket.SendAll(client.conn, packet.Serialize()); err != nil {
			logger.Error("send-message", logger.Fail, messageArgs...)
			return err
		}
		logger.Info("packet-sent", logger.InProgress, "agency-id", client.config.AgencyId, "len", len(packet.Serialize()))
	}
	if err = scanner.Err(); err != nil {
		logger.Error("read-input-file", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
		return err
	}

	// enviar mensaje de fin
	endPacket := protocol.CreateEndPacket()
	if err := safe_socket.SendAll(client.conn, endPacket.Serialize()); err != nil {
		logger.Error("send-end-bets", logger.Fail, messageArgs...)
		return err
	}
	logger.Info("end-packet-sent", logger.InProgress, "agency-id", client.config.AgencyId)

	// TODO: mejorar esto!
	for true {
		// leer primer byte de longitud
		messageLength, err := safe_socket.RecvAll(client.conn, 1)
		if err != nil {
			logger.Error("recv-length", logger.Fail, messageArgs...)
			return err
		}
		logger.Info("msg-len-received", logger.InProgress, "agency-id", client.config.AgencyId, "recv-len", len(messageLength))
		length := uint8(messageLength[0])
		logger.Info("msg-len-received", logger.InProgress, "agency-id", client.config.AgencyId, "actual-len", length)

		// leer paquete per se
		responsePacket, err := safe_socket.RecvAll(client.conn, int(length))
		if err != nil {
			logger.Error("recv-response", logger.Fail, messageArgs...)
			return err
		}

		// actuar segun recepcion
		switch responsePacket[0] {
		case protocol.TYPE_BET:
			betInfo, err := protocol.BetInfoFromBytes(responsePacket, client.config.AgencyId)
			if err != nil {
				logger.Error("deserialize-bet-info", logger.Fail, messageArgs...)
				return err
			}
			if _, err = fmt.Fprintln(outputFile, betInfo.Bet.ToCsv()); err != nil {
				logger.Error("write-output-file", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
				return err
			}
		case protocol.TYPE_END:
			_, err := protocol.EndFromBytes(responsePacket)
			if err != nil {
				logger.Error("deserialize-end", logger.Fail, messageArgs...)
				return err
			}
			logger.Info("end-packet-received", logger.Success, messageArgs...)
			return nil
		default:
			continue // paquete invalido, por ahora lo ignoro (TODO: revisar!)
		}
	}
	return nil
}
