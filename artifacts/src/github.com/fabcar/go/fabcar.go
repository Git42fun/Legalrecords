package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"strings"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/flogging"
)

type SmartContract struct {
	contractapi.Contract
}

var logger = flogging.MustGetLogger("fabcar_cc")

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Access string `json:"access"`
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("Incorrect number of arguments. Expecting 4: ID, Name, Type, Access")
	}

	user := User{
		ID:     args[0],
		Name:   args[1],
		Type:   args[2],
		Access: args[3],
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("Failed to marshal user: %s", err.Error())
	}

	err = ctx.GetStub().PutState(user.ID, userBytes)
	if err != nil {
		return fmt.Errorf("Failed to put state: %s", err.Error())
	}

	return nil
}

func (s *SmartContract) UpdateUser(ctx contractapi.TransactionContextInterface, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("Incorrect number of arguments. Expecting 4: ID, Name, Type, Access")
	}

	user := User{
		ID:     args[0],
		Name:   args[1],
		Type:   args[2],
		Access: args[3],
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("Failed to marshal user: %s", err.Error())
	}

	err = ctx.GetStub().PutState(user.ID, userBytes)
	if err != nil {
		return fmt.Errorf("Failed to put state: %s", err.Error())
	}

	return nil
}

func (s *SmartContract) QueryUser(ctx contractapi.TransactionContextInterface, args []string) (*User, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1: ID")
	}

	userID := args[0]

	userBytes, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get state: %s", err.Error())
	}
	if userBytes == nil {
		return nil, fmt.Errorf("User does not exist: %s", userID)
	}

	var user User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal user: %s", err.Error())
	}

	return &user, nil
}

func (s *SmartContract) ListAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to get state by range: %s", err.Error())
	}
	defer resultsIterator.Close()

	var users []*User

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Failed to get next item from iterator: %s", err.Error())
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal user: %s", err.Error())
		}

		users = append(users, &user)
	}

	return users, nil
}

type LegalRecord struct {
	CaseID          string   `json:"caseID"`
	Language        string   `json:"language"`
	CaseType        string   `json:"caseType"`
	DateCreated     string   `json:"dateCreated"`
	CreatedBy       string   `json:"createdBy"`
	LastUpdated     string   `json:"lastUpdated"`
	LastUpdatedBy   string   `json:"lastUpdatedBy"`
	Judges          []string `json:"judges"`
	CourtType       string   `json:"courtType"`
	CourtCategory   string   `json:"courtCategory"`
	CourtZip        string   `json:"courtZip"`
	Confidentiality string   `json:"confidentiality"`
	UsersWithAccess []string `json:"usersWithAccess,omitempty" metadata:",optional"`
	Description     string   `json:"description"`
	Proceedings     string   `json:"proceedings"` // file path
}

func (s *SmartContract) CreateLegalRecord(ctx contractapi.TransactionContextInterface, legalRecordData string) (string, error) {
	if len(legalRecordData) == 0 {
		return "", fmt.Errorf("Please pass the correct legal record data")
	}

	var legalRecord LegalRecord
	err := json.Unmarshal([]byte(legalRecordData), &legalRecord)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshalling legal record. %s", err.Error())
	}

	legalRecordAsBytes, err := json.Marshal(legalRecord)
	if err != nil {
		return "", fmt.Errorf("Failed while marshalling legal record. %s", err.Error())
	}

	ctx.GetStub().SetEvent("CreateLegalRecord", legalRecordAsBytes)

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(legalRecord.CaseID, legalRecordAsBytes)
}

func (s *SmartContract) UpdateLegalRecord(ctx contractapi.TransactionContextInterface, caseID string, updateFieldsJSON string) error {
	// if len(args) != 2 {
	// 	return fmt.Errorf("Incorrect number of arguments. Expecting 2: caseID and JSON string of fields to update")
	// }

	// caseID := args[0]
	// updateFieldsJSON := args[1]

	// Retrieve the existing legal record
	legalRecordAsBytes, err := ctx.GetStub().GetState(caseID)
	if err != nil {
		return fmt.Errorf("Failed to get legal record: %s", err.Error())
	}
	if legalRecordAsBytes == nil {
		return fmt.Errorf("Legal record does not exist")
	}

	var legalRecord LegalRecord
	err = json.Unmarshal(legalRecordAsBytes, &legalRecord)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal legal record: %s", err.Error())
	}

	// Unmarshal the update fields JSON into a map
	var updateFields map[string]interface{}
	err = json.Unmarshal([]byte(updateFieldsJSON), &updateFields)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal update fields: %s", err.Error())
	}

	// Update the legal record fields
	for field, value := range updateFields {
		switch field {
		case "lastUpdated":
			legalRecord.LastUpdated = value.(string)
		case "lastUpdatedBy":
			legalRecord.LastUpdatedBy = value.(string)
		case "judges":
			// append to existing judges
			// judges := value.([]interface{})
			// var judgesStrings []string
			// for _, judge := range judges {
			// 	judgesStrings = append(judgesStrings, judge.(string))
			// }
			// legalRecord.Judges = judgesStrings
			newJudges := value.([]interface{})
			// Append new judges to the existing list
			for _, newJudge := range newJudges {
				legalRecord.Judges = append(legalRecord.Judges, newJudge.(string))
			}
		case "courtType":
			legalRecord.CourtType = value.(string)
		case "courtCategory":
			legalRecord.CourtCategory = value.(string)
		case "courtZip":
			legalRecord.CourtZip = value.(string)
		case "description":
			legalRecord.Description = value.(string)
		case "proceedings":
			legalRecord.Proceedings = value.(string)
		default:
			return fmt.Errorf("Invalid field name: %s", field)
		}
	}

	// Marshal the updated legal record back to JSON
	updatedLegalRecordAsBytes, err := json.Marshal(legalRecord)
	if err != nil {
		return fmt.Errorf("Failed to marshal legal record: %s", err.Error())
	}

	// Update the legal record in the ledger
	err = ctx.GetStub().PutState(caseID, updatedLegalRecordAsBytes)
	if err != nil {
		return fmt.Errorf("Failed to update legal record: %s", err.Error())
	}

	return nil
}


func (s *SmartContract) QueryLegalRecord(ctx contractapi.TransactionContextInterface, caseID string) (*LegalRecord, error) {
	legalRecordAsBytes, err := ctx.GetStub().GetState(caseID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if legalRecordAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", caseID)
	}

	legalRecord := new(LegalRecord)
	err = json.Unmarshal(legalRecordAsBytes, legalRecord)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal legal record. %s", err.Error())
	}

	return legalRecord, nil
}

func (s *SmartContract) QueryAllLegalRecords(ctx contractapi.TransactionContextInterface) ([]*LegalRecord, error) {
	// Start the query with an empty string to get all keys
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to get state by range: %s", err.Error())
	}
	defer resultsIterator.Close()

	var publicLegalRecords []*LegalRecord

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Failed to get next item from iterator: %s", err.Error())
		}

		var legalRecord LegalRecord
		err = json.Unmarshal(queryResponse.Value, &legalRecord)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal legal record: %s", err.Error())
		}

		// Check if the legal record is public
		if strings.EqualFold(legalRecord.Confidentiality, "PUBLIC") {
			publicLegalRecords = append(publicLegalRecords, &legalRecord)
		}
	}

	return publicLegalRecords, nil
}



type Car struct {
	ID      string `json:"id"`
	Make    string `json:"make"`
	Model   string `json:"model"`
	Color   string `json:"color"`
	Owner   string `json:"owner"`
	AddedAt uint64 `json:"addedAt"`
}

func (s *SmartContract) CreateCar(ctx contractapi.TransactionContextInterface, carData string) (string, error) {

	if len(carData) == 0 {
		return "", fmt.Errorf("Please pass the correct car data")
	}

	var car Car
	err := json.Unmarshal([]byte(carData), &car)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling car. %s", err.Error())
	}

	carAsBytes, err := json.Marshal(car)
	if err != nil {
		return "", fmt.Errorf("failed while marshling car. %s", err.Error())
	}

	ctx.GetStub().SetEvent("CreateAsset", carAsBytes)

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(car.ID, carAsBytes)
}

func (s *SmartContract) Bid(ctx contractapi.TransactionContextInterface, orderID string) (string, error) {
	//verify that submitting client has the role of courier
	// err := ctx.GetClientIdentity().AssertAttributeValue("role", "Courier")
	// if err != nil {
	// 	return "", fmt.Errorf("submitting client not authorized to create a bid, does not have courier role")
	// }
	// get courier bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}
	BidJSON, ok := transientMap["bid"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}
	// get the implicit collection name using the courier's organization ID and verify that courier is targeting their peer to store the bid
	// collection, err := getClientImplicitCollectionNameAndVerifyClientOrg(ctx)
	// if err != nil {
	// 	return "", err
	// }
	// the transaction ID is used as a unique index for the bid
	bidTxID := ctx.GetStub().GetTxID()

	// create a composite key using the transaction ID
	bidKey, err := ctx.GetStub().CreateCompositeKey("bid", []string{orderID, bidTxID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}
	// put the bid into the organization's implicit data collection

	// err = ctx.GetStub().PutPrivateData(collection, bidKey, BidJSON)
	err = ctx.GetStub().PutPrivateData("_implicit_org_Org3MSP", bidKey, []byte(BidJSON))
	if err != nil {
		return "", fmt.Errorf("failed to input bid price into collection: %v", err)
	}
	// return the trannsaction ID so couriers can identify their bid
	return bidTxID, nil
}

func (s *SmartContract) ABACTest(ctx contractapi.TransactionContextInterface, carData string) (string, error) {

	mspId, err := cid.GetMSPID(ctx.GetStub())
	if err != nil {
		return "", fmt.Errorf("failed while getting identity. %s", err.Error())
	}
	if mspId != "Org2MSP" {
		return "", fmt.Errorf("You are not authorized to create Car Data")
	}

	if len(carData) == 0 {
		return "", fmt.Errorf("Please pass the correct car data")
	}

	var car Car
	err = json.Unmarshal([]byte(carData), &car)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling car. %s", err.Error())
	}

	carAsBytes, err := json.Marshal(car)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling car. %s", err.Error())
	}

	ctx.GetStub().SetEvent("CreateAsset", carAsBytes)

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(car.ID, carAsBytes)
}

func (s *SmartContract) CreatePrivateDataImplicitForOrg1(ctx contractapi.TransactionContextInterface, carData string) (string, error) {

	if len(carData) == 0 {
		return "", fmt.Errorf("please pass the correct document data")
	}

	var car Car
	err := json.Unmarshal([]byte(carData), &car)
	if err != nil {
		return "", fmt.Errorf("failed while un-marshalling document. %s", err.Error())
	}

	carAsBytes, err := json.Marshal(car)
	if err != nil {
		return "", fmt.Errorf("failed while marshalling car. %s", err.Error())
	}

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutPrivateData("_implicit_org_Org1MSP", car.ID, carAsBytes)
}

//
func (s *SmartContract) UpdateCarOwner(ctx contractapi.TransactionContextInterface, carID string, newOwner string) (string, error) {

	if len(carID) == 0 {
		return "", fmt.Errorf("Please pass the correct car id")
	}

	carAsBytes, err := ctx.GetStub().GetState(carID)

	if err != nil {
		return "", fmt.Errorf("Failed to get car data. %s", err.Error())
	}

	if carAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", carID)
	}

	car := new(Car)
	_ = json.Unmarshal(carAsBytes, car)

	car.Owner = newOwner

	carAsBytes, err = json.Marshal(car)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling car. %s", err.Error())
	}

	//  txId := ctx.GetStub().GetTxID()

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(car.ID, carAsBytes)

}

func (s *SmartContract) GetHistoryForAsset(ctx contractapi.TransactionContextInterface, carID string) (string, error) {

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(carID)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
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

	return string(buffer.Bytes()), nil
}

func (s *SmartContract) GetCarById(ctx contractapi.TransactionContextInterface, carID string) (*Car, error) {
	if len(carID) == 0 {
		return nil, fmt.Errorf("Please provide correct contract Id")
		// return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, err := ctx.GetStub().GetState(carID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if carAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", carID)
	}

	car := new(Car)
	_ = json.Unmarshal(carAsBytes, car)

	return car, nil

}

func (s *SmartContract) DeleteCarById(ctx contractapi.TransactionContextInterface, carID string) (string, error) {
	if len(carID) == 0 {
		return "", fmt.Errorf("Please provide correct contract Id")
	}

	return ctx.GetStub().GetTxID(), ctx.GetStub().DelState(carID)
}

func (s *SmartContract) GetContractsForQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]Car, error) {

	queryResults, err := s.getQueryResultForQueryString(ctx, queryString)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from ----world state. %s", err.Error())
	}

	return queryResults, nil

}

func (s *SmartContract) getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]Car, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []Car{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		newCar := new(Car)

		err = json.Unmarshal(response.Value, newCar)
		if err != nil {
			return nil, err
		}

		results = append(results, *newCar)
	}
	return results, nil
}

func (s *SmartContract) GetDocumentUsingCarContract(ctx contractapi.TransactionContextInterface, documentID string) (string, error) {
	if len(documentID) == 0 {
		return "", fmt.Errorf("Please provide correct contract Id")
	}

	params := []string{"GetDocumentById", documentID}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := ctx.GetStub().InvokeChaincode("document_cc", queryArgs, "mychannel")

	return string(response.Payload), nil

}

func (s *SmartContract) CreateDocumentUsingCarContract(ctx contractapi.TransactionContextInterface, functionName string, documentData string) (string, error) {
	if len(documentData) == 0 {
		return "", fmt.Errorf("Please provide correct document data")
	}

	params := []string{functionName, documentData}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := ctx.GetStub().InvokeChaincode("document_cc", queryArgs, "mychannel")

	return string(response.Payload), nil

}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincodes: %s", err.Error())
	}

}
