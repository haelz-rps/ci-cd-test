package airbyte_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
	airbyte "github.com/ubix/dripper/manager/repositories/data-extractor/airbyte"
	airbyteclient "github.com/ubix/dripper/manager/repositories/data-extractor/airbyte/airbyte-client"
	fixtures "github.com/ubix/dripper/manager/repositories/data-extractor/airbyte/fixtures"
)

func TestCleanJsonSchema(t *testing.T) {
	input, expected := fixtures.GetJsonSchemaFixtures()

	output := airbyte.CleanJsonSchema(input)

	assert.Equal(t, output, expected)
}

func TestParseSourceReadList(t *testing.T) {
	sourceDefinitionId := fixtures.GetNewUUID()
	sourceId := fixtures.GetNewUUID()
	name := fixtures.GetName()

	input := &airbyteclient.SourceReadList{
		Sources: []airbyteclient.SourceRead{
			{
				SourceDefinitionId: sourceDefinitionId,
				SourceId:           sourceId,
				Name:               name,
			},
		},
	}

	expected := []*dataextractor.Source{
		{
			SourceDefinitionId: sourceDefinitionId,
			SourceId:           sourceId,
			Name:               name,
		},
	}

	testCases := []struct {
		name     string
		input    *airbyteclient.SourceReadList
		expected []*dataextractor.Source
	}{
		{
			name:     "Parse SourceReadList: Default Input",
			input:    input,
			expected: expected,
		},
		{
			name: "Parse SourceReadList: empty Input",
			input: &airbyteclient.SourceReadList{
				Sources: []airbyteclient.SourceRead{},
			},
			expected: []*dataextractor.Source{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()
			output := airbyte.ParseSourceReadList(tc.input)

			assert.Equal(t, output, tc.expected)
		})
	}
}

func TestParseSourceRead(t *testing.T) {
	sourceDefinitionId := fixtures.GetNewUUID()
	sourceId := fixtures.GetNewUUID()
	name := fixtures.GetName()

	configuration := map[string]interface{}{}
	expectedSourceConfiguration := dataextractor.SourceConfiguration{}

	input := airbyteclient.SourceRead{
		SourceDefinitionId:      sourceDefinitionId,
		SourceId:                sourceId,
		Name:                    name,
		ConnectionConfiguration: configuration,
	}

	expected := &dataextractor.Source{
		SourceDefinitionId: sourceDefinitionId,
		SourceId:           sourceId,
		Name:               name,
		Configuration:      &expectedSourceConfiguration,
	}

	testCases := []struct {
		name     string
		input    airbyteclient.SourceRead
		expected *dataextractor.Source
	}{
		{
			name:     "Parse SourceRead: Default Input",
			input:    input,
			expected: expected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()
			output := airbyte.ParseSourceRead(tc.input)

			assert.Equal(t, output, tc.expected)
		})
	}
}

func TestListConnectors(t *testing.T) {
	listConnectorsInput, listConnectorsExpected := fixtures.GetListConnectorsFixtures()
	listConnectorsWithoutScheduleInput, listConnectorsWithoutScheduleExpected := fixtures.GetListConnectorsWithoutSchedule()

	testCases := []struct {
		name     string
		input    *airbyteclient.ConnectionReadList
		expected []*dataextractor.Connector
	}{
		{
			name:     "List Connectors Test: Default Input",
			input:    listConnectorsInput,
			expected: listConnectorsExpected,
		},
		{
			name:     "List Connectors Test: Without Schedule",
			input:    listConnectorsWithoutScheduleInput,
			expected: listConnectorsWithoutScheduleExpected,
		},
		{
			name: "List Connectors Test: When The Input Is Empty",
			input: &airbyteclient.ConnectionReadList{
				Connections: []airbyteclient.ConnectionRead{},
			},

			expected: []*dataextractor.Connector{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()
			t.Name()
			output := airbyte.ParseListConnectionsForWorkspaceWithResponse(tc.input)

			assert.Equal(t, output, tc.expected)
		})
	}
}

func TestParseConnectionRead(t *testing.T) {
	connectionId := fixtures.GetNewUUID()
	destinationId := fixtures.GetNewUUID()
	name := fixtures.GetName()
	operationsIds := fixtures.GetOperationsIds()
	sourceId := fixtures.GetNewUUID()
	connectionScheduleUnits := fixtures.GetConnectionScheduleUnits()

	input := fixtures.GetDataExtractorConnectionReadWithBasicSchedule(
		connectionId,
		destinationId,
		name,
		operationsIds,
		sourceId,
		connectionScheduleUnits,
	)

	expected := fixtures.GetConnectorBasicSchedule(
		connectionId,
		destinationId,
		name,
		operationsIds,
		sourceId,
		connectionScheduleUnits,
	)

	testCases := []struct {
		name     string
		input    *airbyteclient.ConnectionRead
		expected *dataextractor.Connector
	}{
		{
			name:     "Parse ConnectionRead: Default Input",
			input:    &input,
			expected: expected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()
			output := airbyte.ParseConnectionRead(tc.input)

			assert.Equal(t, output, tc.expected)
		})
	}
}

func TestGenerateCreateConnectionRequestBody(t *testing.T) {
	connectionId := fixtures.GetNewUUID()
	destinationId := fixtures.GetNewUUID()
	name := fixtures.GetName()
	operationsIds := fixtures.GetOperationsIds()
	sourceId := fixtures.GetNewUUID()
	connectionScheduleUnits := fixtures.GetConnectionScheduleUnits()

	input := fixtures.GetConnectorBasicSchedule(
		connectionId,
		destinationId,
		name,
		operationsIds,
		sourceId,
		connectionScheduleUnits,
	)

	expected := fixtures.GetCreateConnectionWithBasicSchedule(
		connectionId,
		destinationId,
		name,
		operationsIds,
		sourceId,
		connectionScheduleUnits,
	)

	testCases := []struct {
		name     string
		input    *dataextractor.Connector
		expected airbyteclient.WebBackendConnectionCreate
	}{
		{
			name:     "Generate Create Connection",
			input:    input,
			expected: expected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()
			output := airbyte.GenerateCreateConnectionRequestBody(tc.input, &destinationId)

			assert.Equal(t, output, tc.expected)
		})
	}
}
