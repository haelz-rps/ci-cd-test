package dataextractor

import "github.com/google/uuid"

type SourceDefinition struct {
	SourceDefinitionId uuid.UUID `json:"sourceDefinitionId"`
	Name               string    `json:"name"`
	DocumentationUrl   string    `json:"documentationUrl"`
	Icon               string    `json:"icon"`
}

type SourceConfiguration map[string]interface{}

type SourceDefinitionConfiguration struct {
	SourceDefinitionId  *uuid.UUID           `json:"sourceDefinitionId"`
	DocumentationUrl    *string              `json:"documentationUrl,omitempty"`
	SourceConfiguration *SourceConfiguration `json:"sourceConfiguration"`
}

type Source struct {
	SourceDefinitionId uuid.UUID            `json:"sourceDefinitionId"`
	SourceId           uuid.UUID            `json:"sourceId"`
	Name               string               `json:"name"`
	DocumentationUrl   string               `json:"documentationUrl"`
	Icon               string               `json:"icon"`
	Configuration      *SourceConfiguration `json:"configuration,omitempty"`
}
