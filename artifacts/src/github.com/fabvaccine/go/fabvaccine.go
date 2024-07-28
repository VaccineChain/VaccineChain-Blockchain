package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

// Vaccine :  Define the Vaccine structure, with 4 properties.  Structure tags are used by encoding/json library
type Vaccine struct {
	DeviceId  string `json:"deviceId"`
	VaccineId string `json:"vaccineId"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Unit      string `json:"unit"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("fabvaccine_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryVaccine":
		return s.queryVaccine(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createVaccine":
		return s.createVaccine(APIstub, args)
	case "createPrivateVaccine":
		return s.createPrivateVaccine(APIstub, args)
	case "queryAllVaccines":
		return s.queryAllVaccines(APIstub)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func (s *SmartContract) queryVaccine(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	vaccineAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(vaccineAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	vaccines := []Vaccine{
		{DeviceId: "DEV001", VaccineId: "VAC001", Type: "Temperature", Value: "2.5", Unit: "°C", Timestamp: "2024-07-01T10:00:00", Status: "Normal"},
		{DeviceId: "DEV002", VaccineId: "VAC002", Type: "Temperature", Value: "3.0", Unit: "°C", Timestamp: "2024-07-02T11:00:00", Status: "Normal"},
		{DeviceId: "DEV003", VaccineId: "VAC003", Type: "Temperature", Value: "4.0", Unit: "°C", Timestamp: "2024-07-03T12:00:00", Status: "Normal"},
		{DeviceId: "DEV004", VaccineId: "VAC004", Type: "Temperature", Value: "2.0", Unit: "°C", Timestamp: "2024-07-04T13:00:00", Status: "Normal"},
		{DeviceId: "DEV005", VaccineId: "VAC005", Type: "Temperature", Value: "5.0", Unit: "°C", Timestamp: "2024-07-05T14:00:00", Status: "Normal"},
		{DeviceId: "DEV006", VaccineId: "VAC006", Type: "Temperature", Value: "1.5", Unit: "°C", Timestamp: "2024-07-06T15:00:00", Status: "Normal"},
		{DeviceId: "DEV007", VaccineId: "VAC007", Type: "Temperature", Value: "3.5", Unit: "°C", Timestamp: "2024-07-07T16:00:00", Status: "Normal"},
		{DeviceId: "DEV008", VaccineId: "VAC008", Type: "Temperature", Value: "2.8", Unit: "°C", Timestamp: "2024-07-08T17:00:00", Status: "Normal"},
		{DeviceId: "DEV009", VaccineId: "VAC009", Type: "Temperature", Value: "3.2", Unit: "°C", Timestamp: "2024-07-09T18:00:00", Status: "Normal"},
		{DeviceId: "DEV010", VaccineId: "VAC010", Type: "Temperature", Value: "4.5", Unit: "°C", Timestamp: "2024-07-10T19:00:00", Status: "Normal"},
	}

	i := 0
	for i < len(vaccines) {
		vaccineAsBytes, _ := json.Marshal(vaccines[i])
		APIstub.PutState("VACCINE"+strconv.Itoa(i), vaccineAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createVaccine(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var vaccine = Vaccine{
		DeviceId:  args[0],
		VaccineId: args[1],
		Type:      args[2],
		Value:     args[3],
		Unit:      args[4],
		Timestamp: args[5],
		Status:    args[6],
	}

	vaccineAsBytes, _ := json.Marshal(vaccine)
	APIstub.PutState(args[0], vaccineAsBytes)

	indexName := "owner~key"
	statusIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{vaccine.Status, vaccine.VaccineId})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(statusIndexKey, value)

	return shim.Success(vaccineAsBytes)
}

func (s *SmartContract) createPrivateVaccine(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type vaccineTransientInput struct {
		DeviceId  string `json:"deviceId"` // the field tags are needed to keep case from bouncing around
		VaccineId string `json:"vaccineId"`
		Type      string `json:"type"`
		Value     string `json:"value"`
		Unit      string `json:"unit"`
		Timestamp string `json:"timestamp"`
		Status    string `json:"status"`
		Key       string `json:"key"`
	}
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private vaccine data must be passed in transient map.")
	}

	logger.Infof("Starting createPrivateVaccine")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	vaccineDataAsBytes, ok := transMap["vaccine"]
	if !ok {
		return shim.Error("vaccine must be a key in the transient map")
	}
	logger.Infof("Transient data: " + string(vaccineDataAsBytes))

	if len(vaccineDataAsBytes) == 0 {
		return shim.Error("vaccine value in the transient map must be a non-empty JSON string")
	}

	var vaccineInput vaccineTransientInput
	err = json.Unmarshal(vaccineDataAsBytes, &vaccineInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(vaccineDataAsBytes) + "Error is: " + err.Error())
	}

	if len(vaccineInput.Key) == 0 {
		return shim.Error("key field must be a non-empty string")
	}
	if len(vaccineInput.DeviceId) == 0 {
		return shim.Error("deviceId field must be a non-empty string")
	}
	if len(vaccineInput.VaccineId) == 0 {
		return shim.Error("vaccineId field must be a non-empty string")
	}
	if len(vaccineInput.Type) == 0 {
		return shim.Error("type field must be a non-empty string")
	}
	if len(vaccineInput.Value) == 0 {
		return shim.Error("value field must be a non-empty string")
	}
	if len(vaccineInput.Unit) == 0 {
		return shim.Error("unit field must be a non-empty string")
	}
	if len(vaccineInput.Timestamp) == 0 {
		return shim.Error("timestamp field must be a non-empty string")
	}
	if len(vaccineInput.Status) == 0 {
		return shim.Error("status field must be a non-empty string")
	}

	logger.Infof("Validating vaccine")

	// ==== Check if vaccine already exists ====
	vaccineAsBytes, err := APIstub.GetPrivateData("collectionVaccines", vaccineInput.Key)
	if err != nil {
		return shim.Error("Failed to get vaccine: " + err.Error())
	} else if vaccineAsBytes != nil {
		return shim.Error("This vaccine already exists: " + vaccineInput.Key)
	}

	logger.Infof("Creating new vaccine")

	var vaccine = Vaccine{
		DeviceId:  vaccineInput.DeviceId,
		VaccineId: vaccineInput.VaccineId,
		Type:      vaccineInput.Type,
		Value:     vaccineInput.Value,
		Unit:      vaccineInput.Unit,
		Timestamp: vaccineInput.Timestamp,
		Status:    vaccineInput.Status,
	}

	vaccineAsBytes, err = json.Marshal(vaccine)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutPrivateData("collectionVaccines", vaccineInput.Key, vaccineAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(vaccineAsBytes)
}

func (s *SmartContract) queryAllVaccines(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "DEV001"
	endKey := "DEV010"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllVaccines:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) restictedMethod(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// get an ID for the client which is guaranteed to be unique within the MSP
	//id, err := cid.GetID(APIstub) -

	// get the MSP ID of the client's identity
	//mspid, err := cid.GetMSPID(APIstub) -

	// get the value of the attribute
	//val, ok, err := cid.GetAttributeValue(APIstub, "attr1") -

	// get the X509 certificate of the client, or nil if the client's identity was not based on an X509 certificate
	//cert, err := cid.GetX509Certificate(APIstub) -

	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as APPROVER have access this method!")
	} else {
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1")
		}

		vaccineAsBytes, _ := APIstub.GetState(args[0])
		return shim.Success(vaccineAsBytes)
	}

}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	vaccineName := args[0]

	resultsIterator, err := stub.GetHistoryForKey(vaccineName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForAsset returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
