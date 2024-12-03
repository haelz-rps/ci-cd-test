package dataextractor

import (
	"regexp"
	"strings"
)

type Table struct {
	Name                    string                 `json:"name"`
	JsonSchema              map[string]interface{} `json:"jsonSchema"`
	SupportedSyncModes      []string               `json:"supportedSyncModes"`
	SourceDefinedCursor     bool                   `json:"sourceDefinedCursor"`
	DefaultCursorField      []string               `json:"defaultCursorField"`
	SourceDefinedPrimaryKey [][]string             `json:"sourceDefinedPrimaryKey"`
	SyncEnabled             bool                   `json:"syncEnabled"`
	Namespace               *string                `json:"namespace,omitempty"`
	SelectedCursorField     *[]string              `json:"selectedCursorField,omitempty"`
}

const (
	cleanTableNameRegex  = `[^a-z|A-Z|0-9]`
	tableNameCharReplace = "_"
)

func (t Table) GetHiveName() string {
	cleanTableNameRegexCompiled := regexp.MustCompile(cleanTableNameRegex)
	return strings.ToLower(cleanTableNameRegexCompiled.ReplaceAllString(t.Name, tableNameCharReplace))
}
