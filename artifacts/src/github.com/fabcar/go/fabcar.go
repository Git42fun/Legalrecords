package main

import (
	
	"encoding/json"
	"fmt"
	
	
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
	Password   string `json:"password"`
	Type   string `json:"type"`
	Access string `json:"access"`
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, userData string) (string, error) {
    if len(userData) == 0 {
        return "", fmt.Errorf("Please pass the correct user data")
    }

    var user User
    err := json.Unmarshal([]byte(userData), &user)
    if err != nil {
        return "", fmt.Errorf("Failed while unmarshalling user. %s", err.Error())
    }

    // Validate user data (e.g., check if ID is unique, password complexity, etc.)
    // This is a placeholder for additional validation logic

    userAsBytes, err := json.Marshal(user)
    if err != nil {
        return "", fmt.Errorf("Failed while marshalling user. %s", err.Error())
    }

    // Set an event for the creation of a new user
    ctx.GetStub().SetEvent("CreateUser", userAsBytes)

    // Store the user in the ledger
    return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(user.ID, userAsBytes)
}

// UpdateUser updates an existing user in the ledger
func (s *SmartContract) UpdateUser(ctx contractapi.TransactionContextInterface, userID string, updateFieldsJSON string) error {
    // Retrieve the existing user
    userAsBytes, err := ctx.GetStub().GetState(userID)
    if err != nil {
        return fmt.Errorf("Failed to get user: %s", err.Error())
    }
    if userAsBytes == nil {
        return fmt.Errorf("User does not exist")
    }

    var user User
    err = json.Unmarshal(userAsBytes, &user)
    if err != nil {
        return fmt.Errorf("Failed to unmarshal user: %s", err.Error())
    }

    // Unmarshal the update fields JSON into a map
    var updateFields map[string]interface{}
    err = json.Unmarshal([]byte(updateFieldsJSON), &updateFields)
    if err != nil {
        return fmt.Errorf("Failed to unmarshal update fields: %s", err.Error())
    }

    // Update the user fields
    for field, value := range updateFields {
        switch field {
        case "name":
            user.Name = value.(string)
        case "password":
            user.Password = value.(string)
        case "type":
            user.Type = value.(string)
        case "access":
            user.Access = value.(string)
        default:
            return fmt.Errorf("Invalid field name: %s", field)
        }
    }

    // Marshal the updated user back to JSON
    updatedUserAsBytes, err := json.Marshal(user)
    if err != nil {
        return fmt.Errorf("Failed to marshal user: %s", err.Error())
    }

    // Update the user in the ledger
    err = ctx.GetStub().PutState(userID, updatedUserAsBytes)
    if err != nil {
        return fmt.Errorf("Failed to update user: %s", err.Error())
    }

    return nil
}

func (s *SmartContract) QueryUser(ctx contractapi.TransactionContextInterface, userID string) (*User, error) {
    userAsBytes, err := ctx.GetStub().GetState(userID)

    if err != nil {
        return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
    }

    if userAsBytes == nil {
        return nil, fmt.Errorf("%s does not exist", userID)
    }

    user := new(User)
    err = json.Unmarshal(userAsBytes, user)
    if err != nil {
        return nil, fmt.Errorf("Failed to unmarshal user. %s", err.Error())
    }

    // Optionally, you can add access control logic here if needed

    return user, nil
}

// QueryAllUsers queries all users in the system
func (s *SmartContract) QueryAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
    // Start the query with an empty string to get all users
    queryIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("", []string{})
    if err != nil {
        return nil, fmt.Errorf("Failed to get state iterator. %s", err.Error())
    }
    defer queryIterator.Close()

    var users []*User
    for queryIterator.HasNext() {
        queryResponse, err := queryIterator.Next()
        if err != nil {
            return nil, fmt.Errorf("Failed to get next item from iterator. %s", err.Error())
        }

        var user User
        err = json.Unmarshal(queryResponse.Value, &user)
        if err != nil {
            return nil, fmt.Errorf("Failed to unmarshal user. %s", err.Error())
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
	UsersWithAccess []string `json:"usersWithAccess"`
	Description     string   `json:"description"`
	Proceedings     string   `json:"proceedings"` // file path
}

func (s *SmartContract) CreateLegalRecord(ctx contractapi.TransactionContextInterface, legalRecordData string) (string, error) {
	value, ok, err := cid.GetAttributeValue(ctx.GetStub(), "role")
    if err != nil {
        return "", fmt.Errorf("failed while getting attribute. %s", err.Error())
    }
        if !ok {
        return "", fmt.Errorf("No role attribute found in client identity")
    }
    if value != "approver" {
        return "", fmt.Errorf("You are not authorized to perform this action")
    }

	if len(legalRecordData) == 0 {
		return "", fmt.Errorf("Please pass the correct legal record data")
	}

	var legalRecord LegalRecord
	err2 := json.Unmarshal([]byte(legalRecordData), &legalRecord)
	if err2 != nil {
		return "", fmt.Errorf("Failed while unmarshalling legal record. %s", err2.Error())
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
	value, ok, err := cid.GetAttributeValue(ctx.GetStub(), "role")
    if err != nil {
        return fmt.Errorf("failed while getting attribute. %s", err.Error())
    }
        if !ok {
        return fmt.Errorf("No role attribute found in client identity")
    }
    if value != "approver" {
        return fmt.Errorf("You are not authorized to perform this action")
    }

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


func (s *SmartContract) QueryLegalRecord(ctx contractapi.TransactionContextInterface, caseID string, username string) (*LegalRecord, error) {
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

	hasAccess := false 
    if strings.EqualFold(legalRecord.Confidentiality, "PUBLIC") {
        // iterate through the users with access to check if the user has access
		hasAccess = true

    } else{
        for _, user := range legalRecord.UsersWithAccess {
            if strings.EqualFold(user, username) {
                hasAccess = true
                break
            }
        }
	}

    if !hasAccess {
        return nil, fmt.Errorf("Access Denied: You do not have access to this legal record.")
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
