package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct{}

// VaccineData represents a structure for storing vaccine information
type VaccineData struct {
	VaccineID   string `json:"vaccine_id"`
	DeviceID    string `json:"device_id"`
	Value       string `json:"value"`
	CreatedDate string `json:"created_date"` // Thêm trường thời gian tạo
}

var logger = flogging.MustGetLogger("vaccine_cc")

// Init initializes the smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke routes to the appropriate function based on the function name
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is: %s", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "addVaccineData":
		return s.addVaccineData(APIstub, args)
	case "queryVaccineDataByVaccineID":
		return s.queryVaccineDataByVaccineID(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

// initLedger initializes the ledger with sample data
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	vaccines := []VaccineData{
		{VaccineID: "VAC001", DeviceID: "DEV002", Value: "-4"},
		{VaccineID: "VAC002", DeviceID: "DEV003", Value: "-10"},
	}

	for _, vaccine := range vaccines {
		// Lấy timestamp giả định (hoặc thực tế)
		txTime, err := APIstub.GetTxTimestamp()
		if err != nil {
			return shim.Error("Failed to get transaction timestamp: " + err.Error())
		}
		timestamp := fmt.Sprintf("%d", txTime.Seconds)

		// Thêm trường CreatedDate vào dữ liệu mẫu
		vaccine.CreatedDate = timestamp

		// Tạo khóa bằng VaccineID và timestamp
		compositeKey := vaccine.VaccineID + "_" + timestamp

		dataAsBytes, err := json.Marshal(vaccine)
		if err != nil {
			return shim.Error("Failed to marshal data: " + err.Error())
		}

		err = APIstub.PutState(compositeKey, dataAsBytes)
		if err != nil {
			return shim.Error("Failed to add vaccine data: " + err.Error())
		}
	}

	return shim.Success(nil)
}

// addVaccineData adds a new vaccine data record
func (s *SmartContract) addVaccineData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// Lấy timestamp từ giao dịch
	txTime, err := APIstub.GetTxTimestamp()
	if err != nil {
		return shim.Error("Failed to get transaction timestamp: " + err.Error())
	}
	timestamp := fmt.Sprintf("%d", txTime.Seconds)

	// Tạo Composite Key chỉ với VaccineID và timestamp
	compositeKey := args[0] + "_" + timestamp

	var data = VaccineData{
		VaccineID:   args[0],
		DeviceID:    args[1],
		Value:       args[2],
		CreatedDate: timestamp, // Thêm thời gian tạo
	}

	dataAsBytes, err := json.Marshal(data)
	if err != nil {
		return shim.Error("Failed to marshal data: " + err.Error())
	}

	// Lưu dữ liệu với Composite Key
	err = APIstub.PutState(compositeKey, dataAsBytes)
	if err != nil {
		return shim.Error("Failed to add vaccine data: " + err.Error())
	}

	return shim.Success([]byte("Data added successfully with Key: " + compositeKey))
}

// queryVaccineDataByID queries vaccine data by vaccine_id
func (s *SmartContract) queryVaccineDataByVaccineID(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// Tạo khóa tìm kiếm với prefix là VaccineID
	prefix := args[0] + "_"

	// Truy vấn tất cả bản ghi bắt đầu bằng prefix
	resultsIterator, err := APIstub.GetStateByRange(prefix, prefix+"~")
	if err != nil {
		return shim.Error("Failed to query data by VaccineID: " + err.Error())
	}
	defer resultsIterator.Close()

	var results []VaccineData

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Failed to iterate results: " + err.Error())
		}

		var data VaccineData
		err = json.Unmarshal(queryResponse.Value, &data)
		if err != nil {
			return shim.Error("Failed to unmarshal data: " + err.Error())
		}

		results = append(results, data)
	}

	resultsAsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error("Failed to marshal query results: " + err.Error())
	}

	return shim.Success(resultsAsBytes)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
