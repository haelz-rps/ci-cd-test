package dataextractor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

func TestTableGetHiveName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Parse simple table name",
			input:    "some_table",
			expected: "some_table",
		},
		{
			name:     "Parse buggy table name",
			input:    "Some Buggy Table&123",
			expected: "some_buggy_table_123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()

			table := dataextractor.Table{Name: tc.input}

			output := table.GetHiveName()

			assert.Equal(t, output, tc.expected)
		})
	}

}
