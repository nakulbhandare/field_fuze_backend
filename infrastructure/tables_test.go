package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTables(t *testing.T) {
	table, err := GetTables("test_users1")

	assert.NoError(t, err)
	assert.NotNil(t, table)
	assert.NotNil(t, table.TableName)
	assert.NotNil(t, table.KeySchema)
	assert.NotNil(t, table.AttributeDefinitions)
}

func TestExtractBaseTableName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"prefix_table_name", "name"},
		{"test_users", "users"},
		{"dev_organizations", "organizations"},
		{"prod_jobs", "jobs"},
		{"table_without_prefix", "prefix"},
		{"", ""},
	}

	for _, test := range tests {
		result := extractBaseTableName(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestGetTablesWithDifferentTableNames(t *testing.T) {
	tests := []string{"users1", "role", "organization"}

	for _, tableName := range tests {
		table, err := GetTables(tableName)
		if err == nil {
			assert.NotNil(t, table)
			assert.NotNil(t, table.TableName)
		}
	}
}

func TestExtractBaseTableNameEdgeCases(t *testing.T) {
	// Additional test cases for extractBaseTableName
	tests := []struct {
		input    string
		expected string
	}{
		{"users1", "users1"},
		{"prefix_users1", "users1"},
		{"long_prefix_table_name", "name"},
		{"single", "single"},
	}

	for _, test := range tests {
		result := extractBaseTableName(test.input)
		assert.Equal(t, test.expected, result)
	}
}
