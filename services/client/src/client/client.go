package client

import (
	"bufio"
	"encoding/binary"
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

	scanner := bufio.NewScanner(inputFile)
	keepScanning := true
	for keepScanning {
		// recorrer batch
		batchSize := int(client.config.BatchSize)
		bets := make([]lottery.Bet, 0, batchSize)
		for len(bets) < batchSize {
			keepScanning = scanner.Scan()
			if !keepScanning {
				break
			}
			csvBet := scanner.Text()
			bet, err := lottery.FromCsv(csvBet, client.config.AgencyId)
			if err != nil {
				logger.Error("parse-csv-bet", logger.InProgress, "agency-id", client.config.AgencyId, "err", err)
				return err
			}
			bets = append(bets, bet)
		}
		if err = scanner.Err(); err != nil {
			logger.Error("read-input-file", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
			return err
		}

		// enviar info de apuesta
		packet := protocol.CreateBetInfoPacket(bets, batchSize)
		serializedPacket := packet.Serialize()
		if !keepScanning {
			protocol.SetLastPacketFlag(serializedPacket)
		}
		if err := safe_socket.SendAll(client.conn, serializedPacket); err != nil {
			logger.Error("send-message", logger.Fail, messageArgs...)
			return err
		}
		logger.Info("packet-sent", logger.InProgress, "agency-id", client.config.AgencyId, "len", len(packet.Serialize()))
	}

	keepReceiving := true
	for keepReceiving {
		// leer primer byte de longitud
		messageLength, err := safe_socket.RecvAll(client.conn, protocol.LENGTH_BYTES)
		if err != nil {
			logger.Error("recv-length", logger.Fail, messageArgs...)
			return err
		}
		logger.Info("msg-len-received", logger.InProgress, "agency-id", client.config.AgencyId, "recv-len", len(messageLength))
		length := binary.BigEndian.Uint16(messageLength)
		logger.Info("msg-len-received", logger.InProgress, "agency-id", client.config.AgencyId, "actual-len", length)

		// leer paquete per se
		responsePacket, err := safe_socket.RecvAll(client.conn, int(length))
		if err != nil {
			logger.Error("recv-response", logger.Fail, messageArgs...)
			return err
		}

		isLast := protocol.GetLastPacketFlag(responsePacket)
		keepReceiving = !isLast
		// actuar segun recepcion
		switch responsePacket[0] {
		case protocol.TYPE_BET:
			betInfo, err := protocol.BetInfoFromBytes(responsePacket, client.config.AgencyId, int(client.config.BatchSize))
			if err != nil {
				logger.Error("deserialize-bet-info", logger.Fail, messageArgs...)
				return err
			}
			for _, bet := range betInfo.Bets {
				if _, err = fmt.Fprintln(outputFile, bet.ToCsv()); err != nil {
					logger.Error("write-output-file", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
					return err
				}
			}
		// SACAR ESTO (o refactor)!!
		case protocol.TYPE_END:
			_, err := protocol.EndFromBytes(responsePacket)
			if err != nil {
				logger.Error("deserialize-end", logger.Fail, messageArgs...)
				return err
			}
			logger.Info("end-packet-received", logger.Success, messageArgs...)
			return nil
		default:
			logger.Error("deserialize-packet", logger.Fail, messageArgs...)
			return errors.New("Unknown packet type")
		}
	}
	return nil
}
