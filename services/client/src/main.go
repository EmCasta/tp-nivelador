package main

import (
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/client"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
)

func parseNumber(str string) (uint64, error) {
	number, err := strconv.ParseUint(str, lottery.BASE_10, lottery.BIT_SIZE)
	if err != nil {
		return 0, errors.New("AGENCY_ID must be numeric")
	}
	return number, nil
}

func loadConfig() (client.ClientConfig, error) {
	agencyId := os.Getenv("AGENCY_ID")
	if agencyId == "" {
		return client.ClientConfig{}, errors.New("AGENCY_ID environment variable is required")
	}
	numericId, err := parseNumber(agencyId)
	if err != nil {
		return client.ClientConfig{}, errors.New("AGENCY_ID must be numeric")
	}

	batchSize := os.Getenv("BATCH_SIZE")
	if batchSize == "" {
		return client.ClientConfig{}, errors.New("BATCH_SIZE environment variable is required")
	}
	numericBatchSize, err := parseNumber(batchSize)
	if err != nil {
		return client.ClientConfig{}, errors.New("BATCH_SIZE must be numeric")
	}
	if numericBatchSize == 0 {
		return client.ClientConfig{}, errors.New("BATCH_SIZE cannot be 0")
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		return client.ClientConfig{}, errors.New("SERVER_HOST environment variable is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return client.ClientConfig{}, errors.New("SERVER_PORT environment variable is required")
	}

	inputFile := os.Getenv("INPUT_FILE")
	if inputFile == "" {
		return client.ClientConfig{}, errors.New("INPUT_FILE environment variable is required")
	}

	outputFile := os.Getenv("OUTPUT_FILE")
	if outputFile == "" {
		return client.ClientConfig{}, errors.New("OUTPUT_FILE environment variable is required")
	}

	return client.ClientConfig{
		ServerHost: serverHost,
		ServerPort: serverPort,
		AgencyId:   uint32(numericId),
		BatchSize:  uint8(numericBatchSize),
		InputFile:  inputFile,
		OutputFile: outputFile,
	}, nil
}

func handleSigtermSignal(client *client.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c // bloquear hasta que llegue la señal
		client.GracefulShutdown()
	}()
}

func run() int {
	config, err := loadConfig()
	if err != nil {
		logger.Error("load-config", logger.Fail, "err", err)
		return 1
	}

	client, err := client.NewClient(config)
	if err != nil {
		logger.Error("client-new", logger.Fail, "err", err)
		return 1
	}

	handleSigtermSignal(client)

	if err := client.Run(); err != nil {
		logger.Error("client-run", logger.Fail, "err", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
