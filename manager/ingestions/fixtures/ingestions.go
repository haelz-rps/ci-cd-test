package fixtures

import (
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/ubix/dripper/manager/ingestions"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

func GetNewUUID() uuid.UUID {
	return uuid.New()
}

func GetName() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.IntN(len(letterBytes))]
	}
	return string(b)
}

var (
	cursorField []string               = []string{"sample_cursor_field"}
	jsonSchema  map[string]interface{} = map[string]interface{}{
		"properties": map[string]interface{}{
			"column1": map[string]interface{}{
				"type": "string",
			},
		},
		"type": "object",
	}
	primaryKey            [][]string                        = [][]string{{"key1", "key2"}, {"key3", "key4"}}
	sourceDefinedCursor   bool                              = true
	supportedSyncModesStr []string                          = []string{"full_refresh", "incremental"}
	DocumentationUrl      string                            = "https://example.com/doc"
	IconUrl               string                            = "https://example.com/icon.png"
	Configuration         map[string]interface{}            = map[string]interface{}{"key": "value"}
	SourceConfiguration   dataextractor.SourceConfiguration = map[string]interface{}{"key": "value"}
	SourceStatus          string                            = "active"
)

func GetTable(name string) *dataextractor.Table {
	return &dataextractor.Table{
		Name:                    name,
		JsonSchema:              jsonSchema,
		SupportedSyncModes:      supportedSyncModesStr,
		SourceDefinedCursor:     sourceDefinedCursor,
		DefaultCursorField:      cursorField,
		SourceDefinedPrimaryKey: primaryKey,
	}
}

func GetTablesFinishedMessage(name string) *ingestions.GetSourceTablesStatus {
	return &ingestions.GetSourceTablesStatus{
		Status: ingestions.Finished,
		Tables: []*dataextractor.Table{
			GetTable(name),
		},
	}
}

func GetTablesFinishedMessageInterface(name string) interface{} {
	return struct {
		Status interface{}
		Tables interface{}
	}{
		Status: string(ingestions.Finished),
		Tables: []*dataextractor.Table{
			GetTable(name),
		},
	}
}
