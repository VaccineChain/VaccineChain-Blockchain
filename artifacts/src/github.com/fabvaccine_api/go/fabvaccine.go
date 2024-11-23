package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type VaccineData struct {
	VaccineID string `json:"vaccine_id"`
	DeviceID  string `json:"device_id"`
	Value     string `json:"value"`
}

type QueryResult struct {
	Key    string       `json:"Key"`
	Record *VaccineData `json:"Record"`
}

// InitLedger thêm một số dữ liệu mẫu vào ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	vaccineData := []VaccineData{
		{VaccineID: "VAC1", DeviceID: "DEV1", Value: "75.5"},
		{VaccineID: "VAC2", DeviceID: "DEV2", Value: "80.2"},
		{VaccineID: "VAC3", DeviceID: "DEV3", Value: "65.7"},
	}

	for i, vaccine := range vaccineData {
		vaccineAsBytes, _ := json.Marshal(vaccine)
		err := ctx.GetStub().PutState("VAC"+strconv.Itoa(i+1), vaccineAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}
	return nil
}

// AddVaccineData thêm dữ liệu vaccine vào ledger
func (s *SmartContract) AddVaccineData(ctx contractapi.TransactionContextInterface, vaccineID string, deviceID string, value string) error {
	vaccine := VaccineData{
		VaccineID: vaccineID,
		DeviceID:  deviceID,
		Value:     value,
	}

	vaccineAsBytes, _ := json.Marshal(vaccine)
	return ctx.GetStub().PutState(vaccineID, vaccineAsBytes)
}

// QueryVaccineData lấy dữ liệu vaccine theo VaccineID
func (s *SmartContract) QueryVaccineData(ctx contractapi.TransactionContextInterface, vaccineID string) (*VaccineData, error) {
	vaccineAsBytes, err := ctx.GetStub().GetState(vaccineID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if vaccineAsBytes == nil {
		return nil, fmt.Errorf("VaccineData with ID %s does not exist", vaccineID)
	}

	vaccine := new(VaccineData)
	_ = json.Unmarshal(vaccineAsBytes, vaccine)
	return vaccine, nil
}

// QueryAllVaccineData lấy tất cả dữ liệu vaccine
func (s *SmartContract) QueryAllVaccineData(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := "VAC0"
	endKey := "VAC99"

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		vaccine := new(VaccineData)
		_ = json.Unmarshal(queryResponse.Value, vaccine)

		queryResult := QueryResult{
			Key:    queryResponse.Key,
			Record: vaccine,
		}
		results = append(results, queryResult)
	}

	return results, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create vaccine chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting vaccine chaincode: %s", err.Error())
	}
}
