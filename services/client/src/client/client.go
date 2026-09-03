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
	if err := client.sendHello(); err != nil {
		logger.Error("send-hello", logger.Fail, messageArgs...)
		return err
	}
	logger.Info("client-run:sent-hello", logger.InProgress, messageArgs...)

	// enviar apuestas
	if err = client.sendBets(inputFile); err != nil {
		logger.Error("send-bets", logger.Fail, messageArgs...)
		return err
	}
	logger.Info("client-run:sent-bets", logger.InProgress, messageArgs...)

	// recibir ganadores
	if err = client.receiveWinners(outputFile); err != nil {
		logger.Error("receive-winners", logger.Fail, messageArgs...)
		return err
	}
	logger.Info("client-run:received-winners", logger.InProgress, messageArgs...)

	return nil
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
	helloPacket := protocol.CreateHelloPacket(client.config.AgencyId, client.config.BatchSize)
	if err := safe_socket.SendAll(client.conn, helloPacket.Serialize()); err != nil {
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
		// recorrer batch
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
		packet := protocol.CreateBetInfoPacket(bets)
		serializedPacket := packet.Serialize()
		if !keepScanning {
			protocol.SetLastPacketFlag(serializedPacket, protocol.LENGTH_BYTES)
		}
		if err := safe_socket.SendAll(client.conn, serializedPacket); err != nil {
			return err
		}
		//logger.Info("client-run:sending-bets", logger.InProgress, "sending-batch:last", !keepScanning, "bets", bets)
	}
	return nil
}

func (client *Client) receiveWinners(outputFile *os.File) error {
	keepReceiving := true
	for keepReceiving {
		messageLength, err := safe_socket.RecvAll(client.conn, protocol.LENGTH_BYTES)
		if err != nil {
			return err
		}
		logger.Info("receiving-winners", logger.InProgress, "message-len", binary.BigEndian.Uint16(messageLength))
		length := binary.BigEndian.Uint16(messageLength)
		responsePacket, err := safe_socket.RecvAll(client.conn, int(length))
		if err != nil {
			return err
		}

		logger.Info("response-packet", logger.InProgress, responsePacket[0])
		isLast := protocol.GetLastPacketFlag(responsePacket, 0)
		keepReceiving = !isLast
		switch responsePacket[0] {
		case protocol.TYPE_BET:
			betInfo, err := protocol.BetInfoFromBytes(responsePacket, client.config.AgencyId, int(client.config.BatchSize))
			logger.Info("winners-recvd", logger.InProgress, "keep-receiving", keepReceiving)
			if err != nil {
				return err
			}
			for _, bet := range betInfo.Bets {
				if _, err = fmt.Fprintln(outputFile, bet.ToCsv()); err != nil {
					return err
				}
			}
			logger.Info("bets-received", logger.InProgress, "bets", betInfo.Bets)
		default:
			return errors.New("Unknown packet type")
		}
	}
	return nil
}
