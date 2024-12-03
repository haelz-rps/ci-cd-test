package hivemetastore_test

import (
	"sort"
	"testing"

	"github.com/beltran/gohive/hive_metastore"
	"github.com/stretchr/testify/assert"

	hivemetastore "github.com/ubix/dripper/manager/repositories/data-catalog/hive-metastore"
	"github.com/ubix/dripper/manager/repositories/data-catalog/hive-metastore/fixtures"
)

func TestMockJsonSchemaParse(t *testing.T) {
	simpleSchemaInput, simpleSchemaExpected := fixtures.GetSimpleJsonSchemaFixtures()
	nestedSchemaInput, nestedSchemaExpected := fixtures.GetNestedJsonSchemaFixtures()

	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected []*hive_metastore.FieldSchema
	}{
		{
			name:     "Simple Json Schema",
			input:    simpleSchemaInput,
			expected: simpleSchemaExpected,
		},
		{
			name:     "Complex Nested Json Schema",
			input:    nestedSchemaInput,
			expected: nestedSchemaExpected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()
			t.Name()
			output := hivemetastore.GetColumnsFromJsonSchema(tc.input)

			// sort the output to avoid random json keys parser error
			sort.Slice(output, func(i, j int) bool {
				return output[i].Name < output[j].Name
			})

			assert.Equal(t, len(tc.expected), len(output))
			for i := range tc.expected {
				assert.Equal(t, *tc.expected[i], *output[i])
			}
		})
	}
}

func TestRealJsonSchemaParse(t *testing.T) {
	hubspotDealsSchemaInput, hubspotDealsLengthExpected := fixtures.GetHubspotDealsJsonSchemaFixtures()
	shopifyOrdersSchemaInput, shopifyOrdersLengthExpected := fixtures.GetShopifyOrdersJsonSchemaFixtures()

	testCases := []struct {
		name           string
		input          map[string]interface{}
		expectedLength int
	}{
		{
			name:           "Hubspot Deals Json Schema",
			input:          hubspotDealsSchemaInput,
			expectedLength: hubspotDealsLengthExpected,
		},
		{
			name:           "Shopify Orders Json Schema",
			input:          shopifyOrdersSchemaInput,
			expectedLength: shopifyOrdersLengthExpected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()
			t.Name()
			output := hivemetastore.GetColumnsFromJsonSchema(tc.input)

			assert.Equal(t, tc.expectedLength, len(output))
		})
	}
}
