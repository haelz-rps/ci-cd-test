package ingestions

import (
	"encoding/json"
)

func ParseGetSourceTablesStatusMessage(tablesMessage interface{}) (*GetSourceTablesStatus, error) {
	var message *GetSourceTablesStatus
	jsonValue, err := json.Marshal(tablesMessage)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonValue, &message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func ParseUpdateSourceStatusFromKafka(updateSourceMessage interface{}) (*UpdateSourceStatus, error) {
	var message *UpdateSourceStatus
	jsonValue, err := json.Marshal(updateSourceMessage)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonValue, &message)
	if err != nil {
		return nil, err
	}
	return message, nil
}
