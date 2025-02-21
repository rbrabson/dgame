package discjson

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// FieldPresent checks to see if the field is present in the JSON object.
func FieldPresent(inputJSON []byte, fieldName string) bool {
	log.Trace("--> FieldPresent")
	defer log.Trace("<-- FieldPresent")

	var data map[string]interface{}
	json.Unmarshal(inputJSON, &data)

	// Add the new field if it doesn't exist
	if data[fieldName] != nil {
		return true
	}
	return false
}

// getFilename adds an arbritary field to a JSON object.
func AddFieldToJSON(inputJSON []byte, fieldName string, fieldValue interface{}) ([]byte, error) {
	log.Trace("--> AddFieldToJSON")
	defer log.Trace("<-- AddFieldToJSON")

	// Unmarshal the JSON into a map
	var data map[string]interface{}
	err := json.Unmarshal(inputJSON, &data)
	if err != nil {
		return nil, err
	}

	// Add the new field if it doesn't exist
	if data[fieldName] != nil {
		return inputJSON, nil
	}
	data[fieldName] = fieldValue

	// Marshal the map back to JSON
	updatedJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	fmt.Println((string)(updatedJSON))

	return updatedJSON, nil
}
