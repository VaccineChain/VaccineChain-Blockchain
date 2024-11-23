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
	VaccineID string `json:"vaccine_id"`
	DeviceID  string `json:"device_id"`
	Value     string `json:"value"`
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
	case "queryVaccineDataByID":
		return s.queryVaccineDataByID(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

// initLedger initializes the ledger with sample data
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	vaccines := []VaccineData{
		{VaccineID: "VAC001", DeviceID: "DEV001", Value: "-3"},
		{VaccineID: "VAC002", DeviceID: "DEV002", Value: "-4"},
		{VaccineID: "VAC003", DeviceID: "DEV003", Value: "-10"},
	}

	for _, vaccine := range vaccines {
		dataAsBytes, err := json.Marshal(vaccine)
		if err != nil {
			return shim.Error("Failed to marshal data: " + err.Error())
		}

		err = APIstub.PutState(vaccine.VaccineID, dataAsBytes)
		if err != nil {
			return shim.Error("Failed to add vaccine data: " + err.Error())
		}
	}

	return shim.Success(nil)
}

// addVaccineData adds a new vaccine data record
func (s *SmartContract) addVaccineData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Ensure client has the correct attribute/role
	// val, ok, err := cid.GetAttributeValue(APIstub, "role")
	// if err != nil || !ok || val != "admin" {
	// 	return shim.Error("Client identity does not possess the required attribute 'role' with value 'admin'")
	// }

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	var data = VaccineData{
		VaccineID: args[0],
		DeviceID:  args[1],
		Value:     args[2],
	}

	dataAsBytes, err := json.Marshal(data)
	if err != nil {
		return shim.Error("Failed to marshal data: " + err.Error())
	}

	err = APIstub.PutState(data.VaccineID, dataAsBytes)
	if err != nil {
		return shim.Error("Failed to add vaccine data: " + err.Error())
	}

	return shim.Success(dataAsBytes)
}

// queryVaccineDataByID queries vaccine data by vaccine_id
func (s *SmartContract) queryVaccineDataByID(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	dataAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get data: " + err.Error())
	} else if dataAsBytes == nil {
		return shim.Error("No data found for vaccine_id: " + args[0])
	}

	return shim.Success(dataAsBytes)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
