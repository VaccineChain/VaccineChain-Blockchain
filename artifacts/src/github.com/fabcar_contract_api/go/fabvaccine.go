package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Log represents a single log entry
type Log struct {
	Value     string `json:"value"`
	Timestamp string `json:"timestamp"`
	IDVacxin  string `json:"idvacxin"`
}

// LogChaincode provides functions for managing a log
type LogChaincode struct {
	contractapi.Contract
}

// InitLedger adds a base set of logs to the ledger
func (c *LogChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	logs := []Log{
		{Value: "Sample Log 1", Timestamp: time.Now().Format(time.RFC3339), IDVacxin: "VAC123456"},
		{Value: "Sample Log 2", Timestamp: time.Now().Format(time.RFC3339), IDVacxin: "VAC123457"},
	}

	for i, log := range logs {
		logAsBytes, _ := json.Marshal(log)
		err := ctx.GetStub().PutState(fmt.Sprintf("LOG%d", i), logAsBytes)

		if err != nil {
			return fmt.Errorf("failed to put log to world state: %v", err)
		}
	}

	return nil
}

// CreateLog adds a new log to the ledger
func (c *LogChaincode) CreateLog(ctx contractapi.TransactionContextInterface, logID string, value string, idvacxin string) error {
	timestamp := time.Now().Format(time.RFC3339)
	log := Log{
		Value:     value,
		Timestamp: timestamp,
		IDVacxin:  idvacxin,
	}

	logAsBytes, _ := json.Marshal(log)

	return ctx.GetStub().PutState(logID, logAsBytes)
}

// ReadLog retrieves a log from the ledger
func (c *LogChaincode) ReadLog(ctx contractapi.TransactionContextInterface, logID string) (*Log, error) {
	logAsBytes, err := ctx.GetStub().GetState(logID)

	if err != nil {
		return nil, fmt.Errorf("failed to read log from world state: %v", err)
	}

	if logAsBytes == nil {
		return nil, fmt.Errorf("log does not exist: %s", logID)
	}

	log := new(Log)
	_ = json.Unmarshal(logAsBytes, log)

	return log, nil
}

// QueryLogsByVacxinID retrieves all logs with the given vacxin ID from the ledger
func (c *LogChaincode) QueryLogsByVacxinID(ctx contractapi.TransactionContextInterface, idvacxin string) ([]*Log, error) {
	queryString := fmt.Sprintf(`{"selector":{"idvacxin":"%s"}}`, idvacxin)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var logs []*Log
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var log Log
		err = json.Unmarshal(queryResponse.Value, &log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// UpdateLog updates an existing log in the ledger
func (c *LogChaincode) UpdateLog(ctx contractapi.TransactionContextInterface, logID string, value string, idvacxin string) error {
	timestamp := time.Now().Format(time.RFC3339)
	log := Log{
		Value:     value,
		Timestamp: timestamp,
		IDVacxin:  idvacxin,
	}

	logAsBytes, _ := json.Marshal(log)

	return ctx.GetStub().PutState(logID, logAsBytes)
}

// DeleteLog deletes a log from the ledger
func (c *LogChaincode) DeleteLog(ctx contractapi.TransactionContextInterface, logID string) error {
	return ctx.GetStub().DelState(logID)
}

func main() {
	chaincode, err := contractapi.NewChaincode(&LogChaincode{})
	if err != nil {
		fmt.Printf("Error creating log chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting log chaincode: %s", err.Error())
	}
}
