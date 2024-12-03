package ingestions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubix/dripper/manager/ingestions"
	"github.com/ubix/dripper/manager/ingestions/fixtures"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

func TestParseUpdateSourceStatusFromKafka(t *testing.T) {
	sourceDefinitionId := fixtures.GetNewUUID()
	sourceId := fixtures.GetNewUUID()
	name := fixtures.GetName()

	input := map[string]interface{}{
		"source": map[string]interface{}{
			"sourceDefinitionId": sourceDefinitionId.String(),
			"sourceId":           sourceId.String(),
			"name":               name,
			"documentationUrl":   fixtures.DocumentationUrl,
			"icon":               fixtures.IconUrl,
			"configuration":      fixtures.Configuration,
		},
		"status": fixtures.SourceStatus,
	}

	expectedSource := &dataextractor.Source{
		SourceDefinitionId: sourceDefinitionId,
		SourceId:           sourceId,
		Name:               name,
		DocumentationUrl:   fixtures.DocumentationUrl,
		Icon:               fixtures.IconUrl,
		Configuration:      &fixtures.SourceConfiguration,
	}
	expectedStatus := fixtures.SourceStatus
	expected := &ingestions.UpdateSourceStatus{
		Source: expectedSource,
		Status: ingestions.Status(expectedStatus),
	}

	testCases := []struct {
		name        string
		input       interface{}
		expected    *ingestions.UpdateSourceStatus
		expectError bool
	}{
		{
			name:        "Parse UpdateSourceStatusFromKafka: Valid Input",
			input:       input,
			expected:    expected,
			expectError: false,
		},
		{
			name:  "Parse UpdateSourceStatusFromKafka: Missing Source",
			input: map[string]interface{}{"status": fixtures.SourceStatus},
			expected: &ingestions.UpdateSourceStatus{
				Source: nil,
				Status: ingestions.Status(fixtures.SourceStatus),
			},
			expectError: false,
		},
		{
			name: "Parse UpdateSourceStatusFromKafka: Invalid UUID Format",
			input: map[string]interface{}{
				"source": map[string]interface{}{
					"sourceDefinitionId": "invalid-uuid",
					"sourceId":           sourceId.String(),
					"name":               name,
					"documentationUrl":   fixtures.DocumentationUrl,
					"icon":               fixtures.IconUrl,
				},
				"status": fixtures.SourceStatus,
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()

			output, err := ingestions.ParseUpdateSourceStatusFromKafka(tc.input)

			if tc.expectError {
				assert.Error(subT, err)
				return
			}

			assert.NoError(subT, err)
			assert.Equal(subT, tc.expected, output)
		})
	}
}

func TestParseGetSourceTablesStatusMessage(t *testing.T) {
	name := fixtures.GetName()
	input := fixtures.GetTablesFinishedMessageInterface(name)
	output := fixtures.GetTablesFinishedMessage(name)

	testCases := []struct {
		name     string
		input    interface{}
		expected *ingestions.GetSourceTablesStatus
	}{
		{
			name:     "Parse GetSourceTablesStatus message: default input",
			input:    input,
			expected: output,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()
			output, err := ingestions.ParseGetSourceTablesStatusMessage(tc.input)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			assert.Equal(t, tc.expected, output)
		})
	}
}
