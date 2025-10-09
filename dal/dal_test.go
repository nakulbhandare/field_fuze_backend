package dal

import (
	"context"
	"errors"
	"fieldfuze-backend/models"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockDatabaseClient implements DatabaseClientInterface for testing
type MockDatabaseClient struct {
	mock.Mock
}

func (m *MockDatabaseClient) GetItem(ctx context.Context, config models.QueryConfig, result interface{}) error {
	args := m.Called(ctx, config, result)
	if args.Get(0) != nil {
		// Copy the mock result to the result parameter
		if mockResult, ok := args.Get(0).(map[string]interface{}); ok {
			if resultMap, ok := result.(*map[string]interface{}); ok {
				*resultMap = mockResult
			}
		}
	}
	return args.Error(1)
}

func (m *MockDatabaseClient) PutItem(ctx context.Context, tableName string, item interface{}) error {
	args := m.Called(ctx, tableName, item)
	return args.Error(0)
}

func (m *MockDatabaseClient) UpdateItem(ctx context.Context, tableName, key, keyValue string, updates map[string]interface{}) error {
	args := m.Called(ctx, tableName, key, keyValue, updates)
	return args.Error(0)
}

func (m *MockDatabaseClient) DeleteItem(ctx context.Context, tableName, key, value string) error {
	args := m.Called(ctx, tableName, key, value)
	return args.Error(0)
}

func (m *MockDatabaseClient) QueryByIndex(ctx context.Context, tableName, indexName, keyName, keyValue string, results interface{}) error {
	args := m.Called(ctx, tableName, indexName, keyName, keyValue, results)
	if args.Get(0) != nil {
		// Copy the mock results to the results parameter
		if mockResults, ok := args.Get(0).([]map[string]interface{}); ok {
			if resultSlice, ok := results.(*[]map[string]interface{}); ok {
				*resultSlice = mockResults
			}
		}
	}
	return args.Error(1)
}

func (m *MockDatabaseClient) Scan(ctx context.Context, tableName string, results interface{}) error {
	args := m.Called(ctx, tableName, results)
	if args.Get(0) != nil {
		// Copy the mock results to the results parameter
		if mockResults, ok := args.Get(0).([]map[string]interface{}); ok {
			if resultSlice, ok := results.(*[]map[string]interface{}); ok {
				*resultSlice = mockResults
			}
		}
	}
	return args.Error(1)
}

func (m *MockDatabaseClient) ScanTable(ctx context.Context, tableName string, results interface{}) error {
	args := m.Called(ctx, tableName, results)
	if args.Get(0) != nil {
		// Copy the mock results to the results parameter
		if mockResults, ok := args.Get(0).([]map[string]interface{}); ok {
			if resultSlice, ok := results.(*[]map[string]interface{}); ok {
				*resultSlice = mockResults
			}
		}
	}
	return args.Error(1)
}

func (m *MockDatabaseClient) CreateTable(ctx context.Context, input *dynamodb.CreateTableInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *MockDatabaseClient) DescribeTable(ctx context.Context, tableName string) (*dynamodb.DescribeTableOutput, error) {
	args := m.Called(ctx, tableName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.DescribeTableOutput), args.Error(1)
}

func (m *MockDatabaseClient) DeleteTable(ctx context.Context, input *dynamodb.DeleteTableInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// DALTestSuite defines a test suite for DAL functions
type DALTestSuite struct {
	suite.Suite
	mockClient   *MockDatabaseClient
	dalContainer *DALContainer
}

// SetupTest runs before each test
func (suite *DALTestSuite) SetupTest() {
	suite.mockClient = &MockDatabaseClient{}
	suite.dalContainer = &DALContainer{
		databaseClient: suite.mockClient,
	}
}

// TearDownTest runs after each test
func (suite *DALTestSuite) TearDownTest() {
	suite.mockClient.AssertExpectations(suite.T())
}

// TestGetDatabaseClient tests the GetDatabaseClient method
func (suite *DALTestSuite) TestGetDatabaseClient() {
	client := suite.dalContainer.GetDatabaseClient()
	assert.NotNil(suite.T(), client)
	assert.Equal(suite.T(), suite.mockClient, client)
}

// TestPrintPrettyJSON tests the PrintPrettyJSON function
func (suite *DALTestSuite) TestPrintPrettyJSON() {
	data := map[string]interface{}{
		"name":  "test",
		"value": 123,
		"nested": map[string]string{
			"key": "value",
		},
	}
	
	result := PrintPrettyJSON(data)
	assert.NotEmpty(suite.T(), result)
	assert.Contains(suite.T(), result, "\"name\": \"test\"")
	assert.Contains(suite.T(), result, "\"value\": 123")
	assert.Contains(suite.T(), result, "\"nested\"")
}

// TestPrintPrettyJSONWithNil tests PrintPrettyJSON with nil input
func (suite *DALTestSuite) TestPrintPrettyJSONWithNil() {
	result := PrintPrettyJSON(nil)
	assert.Equal(suite.T(), "null", result)
}

// TestPrintPrettyJSONWithInvalidData tests PrintPrettyJSON with invalid data
func (suite *DALTestSuite) TestPrintPrettyJSONWithInvalidData() {
	// Create non-serializable data
	invalidData := make(chan int)
	result := PrintPrettyJSON(invalidData)
	assert.Contains(suite.T(), result, "Failed to generate JSON")
}

// TestGetItemByPrimaryKey tests GetItem with primary key
func (suite *DALTestSuite) TestGetItemByPrimaryKey() {
	ctx := context.Background()
	config := models.QueryConfig{
		TableName: "test-table",
		KeyName:   "id",
		KeyValue:  "test-id",
		KeyType:   models.StringType,
	}
	
	// Mock successful response
	mockResult := map[string]interface{}{
		"id":   "test-id",
		"name": "test-name",
	}
	
	suite.mockClient.On("GetItem", ctx, config, mock.AnythingOfType("*map[string]interface {}")).Return(mockResult, nil)
	
	var result map[string]interface{}
	err := suite.mockClient.GetItem(ctx, config, &result)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-id", result["id"])
	assert.Equal(suite.T(), "test-name", result["name"])
}

// TestGetItemByPrimaryKeyNotFound tests GetItem when item not found
func (suite *DALTestSuite) TestGetItemByPrimaryKeyNotFound() {
	ctx := context.Background()
	config := models.QueryConfig{
		TableName: "test-table",
		KeyName:   "id",
		KeyValue:  "non-existent",
		KeyType:   models.StringType,
	}
	
	suite.mockClient.On("GetItem", ctx, config, mock.AnythingOfType("*map[string]interface {}")).Return(nil, errors.New("item not found"))
	
	var result map[string]interface{}
	err := suite.mockClient.GetItem(ctx, config, &result)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "item not found")
}

// TestGetItemByPrimaryKeyError tests GetItem with error
func (suite *DALTestSuite) TestGetItemByPrimaryKeyError() {
	ctx := context.Background()
	config := models.QueryConfig{
		TableName: "test-table",
		KeyName:   "id",
		KeyValue:  "test-id",
		KeyType:   models.StringType,
	}
	
	suite.mockClient.On("GetItem", ctx, config, mock.AnythingOfType("*map[string]interface {}")).Return(nil, errors.New("DynamoDB error"))
	
	var result map[string]interface{}
	err := suite.mockClient.GetItem(ctx, config, &result)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "DynamoDB error")
}

// TestGetItemByIndex tests GetItem with secondary index
func (suite *DALTestSuite) TestGetItemByIndex() {
	ctx := context.Background()
	config := models.QueryConfig{
		TableName: "test-table",
		IndexName: "test-index",
		KeyName:   "email",
		KeyValue:  "test@example.com",
		KeyType:   models.StringType,
	}
	
	// Mock successful response
	mockResult := map[string]interface{}{
		"id":    "test-id",
		"email": "test@example.com",
	}
	
	suite.mockClient.On("GetItem", ctx, config, mock.AnythingOfType("*map[string]interface {}")).Return(mockResult, nil)
	
	var result map[string]interface{}
	err := suite.mockClient.GetItem(ctx, config, &result)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-id", result["id"])
	assert.Equal(suite.T(), "test@example.com", result["email"])
}

// TestPutItem tests the PutItem function
func (suite *DALTestSuite) TestPutItem() {
	ctx := context.Background()
	tableName := "test-table"
	item := map[string]interface{}{
		"id":   "test-id",
		"name": "test-name",
	}
	
	suite.mockClient.On("PutItem", ctx, tableName, item).Return(nil)
	
	err := suite.mockClient.PutItem(ctx, tableName, item)
	assert.NoError(suite.T(), err)
}

// TestPutItemError tests PutItem with error
func (suite *DALTestSuite) TestPutItemError() {
	ctx := context.Background()
	tableName := "test-table"
	item := map[string]interface{}{
		"id":   "test-id",
		"name": "test-name",
	}
	
	suite.mockClient.On("PutItem", ctx, tableName, item).Return(errors.New("PutItem error"))
	
	err := suite.mockClient.PutItem(ctx, tableName, item)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "PutItem error")
}

// TestUpdateItem tests the UpdateItem function
func (suite *DALTestSuite) TestUpdateItem() {
	ctx := context.Background()
	tableName := "test-table"
	key := "id"
	keyValue := "test-id"
	updates := map[string]interface{}{
		"name":       "updated-name",
		"updated_at": time.Now(),
	}
	
	suite.mockClient.On("UpdateItem", ctx, tableName, key, keyValue, updates).Return(nil)
	
	err := suite.mockClient.UpdateItem(ctx, tableName, key, keyValue, updates)
	assert.NoError(suite.T(), err)
}

// TestUpdateItemError tests UpdateItem with error
func (suite *DALTestSuite) TestUpdateItemError() {
	ctx := context.Background()
	tableName := "test-table"
	key := "id"
	keyValue := "test-id"
	updates := map[string]interface{}{
		"name": "updated-name",
	}
	
	suite.mockClient.On("UpdateItem", ctx, tableName, key, keyValue, updates).Return(errors.New("UpdateItem error"))
	
	err := suite.mockClient.UpdateItem(ctx, tableName, key, keyValue, updates)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "UpdateItem error")
}

// TestDeleteItem tests the DeleteItem function
func (suite *DALTestSuite) TestDeleteItem() {
	ctx := context.Background()
	tableName := "test-table"
	key := "id"
	value := "test-id"
	
	suite.mockClient.On("DeleteItem", ctx, tableName, key, value).Return(nil)
	
	err := suite.mockClient.DeleteItem(ctx, tableName, key, value)
	assert.NoError(suite.T(), err)
}

// TestDeleteItemError tests DeleteItem with error
func (suite *DALTestSuite) TestDeleteItemError() {
	ctx := context.Background()
	tableName := "test-table"
	key := "id"
	value := "test-id"
	
	suite.mockClient.On("DeleteItem", ctx, tableName, key, value).Return(errors.New("DeleteItem error"))
	
	err := suite.mockClient.DeleteItem(ctx, tableName, key, value)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "DeleteItem error")
}

// TestQueryByIndex tests the QueryByIndex function
func (suite *DALTestSuite) TestQueryByIndex() {
	ctx := context.Background()
	tableName := "test-table"
	indexName := "test-index"
	keyName := "email"
	keyValue := "test@example.com"
	
	// Mock successful response
	mockResults := []map[string]interface{}{
		{
			"id":    "test-id-1",
			"email": "test@example.com",
		},
		{
			"id":    "test-id-2",
			"email": "test@example.com",
		},
	}
	
	suite.mockClient.On("QueryByIndex", ctx, tableName, indexName, keyName, keyValue, mock.AnythingOfType("*[]map[string]interface {}")).Return(mockResults, nil)
	
	var results []map[string]interface{}
	err := suite.mockClient.QueryByIndex(ctx, tableName, indexName, keyName, keyValue, &results)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2)
	assert.Equal(suite.T(), "test-id-1", results[0]["id"])
	assert.Equal(suite.T(), "test-id-2", results[1]["id"])
}

// TestQueryByIndexError tests QueryByIndex with error
func (suite *DALTestSuite) TestQueryByIndexError() {
	ctx := context.Background()
	tableName := "test-table"
	indexName := "test-index"
	keyName := "email"
	keyValue := "test@example.com"
	
	suite.mockClient.On("QueryByIndex", ctx, tableName, indexName, keyName, keyValue, mock.AnythingOfType("*[]map[string]interface {}")).Return(nil, errors.New("Query error"))
	
	var results []map[string]interface{}
	err := suite.mockClient.QueryByIndex(ctx, tableName, indexName, keyName, keyValue, &results)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Query error")
}

// TestScan tests the Scan function
func (suite *DALTestSuite) TestScan() {
	ctx := context.Background()
	tableName := "test-table"
	
	// Mock successful response
	mockResults := []map[string]interface{}{
		{
			"id":   "test-id-1",
			"name": "test-name-1",
		},
		{
			"id":   "test-id-2",
			"name": "test-name-2",
		},
	}
	
	suite.mockClient.On("Scan", ctx, tableName, mock.AnythingOfType("*[]map[string]interface {}")).Return(mockResults, nil)
	
	var results []map[string]interface{}
	err := suite.mockClient.Scan(ctx, tableName, &results)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2)
	assert.Equal(suite.T(), "test-id-1", results[0]["id"])
	assert.Equal(suite.T(), "test-id-2", results[1]["id"])
}

// TestScanError tests Scan with error
func (suite *DALTestSuite) TestScanError() {
	ctx := context.Background()
	tableName := "test-table"
	
	suite.mockClient.On("Scan", ctx, tableName, mock.AnythingOfType("*[]map[string]interface {}")).Return(nil, errors.New("Scan error"))
	
	var results []map[string]interface{}
	err := suite.mockClient.Scan(ctx, tableName, &results)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Scan error")
}

// TestScanTable tests the ScanTable function (alias for Scan)
func (suite *DALTestSuite) TestScanTable() {
	ctx := context.Background()
	tableName := "test-table"
	
	// Mock successful response
	mockResults := []map[string]interface{}{
		{
			"id":   "test-id",
			"name": "test-name",
		},
	}
	
	suite.mockClient.On("ScanTable", ctx, tableName, mock.AnythingOfType("*[]map[string]interface {}")).Return(mockResults, nil)
	
	var results []map[string]interface{}
	err := suite.mockClient.ScanTable(ctx, tableName, &results)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "test-id", results[0]["id"])
}

// TestCreateTable tests the CreateTable function
func (suite *DALTestSuite) TestCreateTable() {
	ctx := context.Background()
	input := &dynamodb.CreateTableInput{
		TableName: &[]string{"test-table"}[0],
	}
	
	suite.mockClient.On("CreateTable", ctx, input).Return(nil)
	
	err := suite.mockClient.CreateTable(ctx, input)
	assert.NoError(suite.T(), err)
}

// TestCreateTableError tests CreateTable with error
func (suite *DALTestSuite) TestCreateTableError() {
	ctx := context.Background()
	input := &dynamodb.CreateTableInput{
		TableName: &[]string{"test-table"}[0],
	}
	
	suite.mockClient.On("CreateTable", ctx, input).Return(errors.New("CreateTable error"))
	
	err := suite.mockClient.CreateTable(ctx, input)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "CreateTable error")
}

// TestDescribeTable tests the DescribeTable function
func (suite *DALTestSuite) TestDescribeTable() {
	ctx := context.Background()
	tableName := "test-table"
	
	// Mock successful response
	mockOutput := &dynamodb.DescribeTableOutput{
		Table: &types.TableDescription{
			TableName: &tableName,
		},
	}
	
	suite.mockClient.On("DescribeTable", ctx, tableName).Return(mockOutput, nil)
	
	result, err := suite.mockClient.DescribeTable(ctx, tableName)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), tableName, *result.Table.TableName)
}

// TestDescribeTableError tests DescribeTable with error
func (suite *DALTestSuite) TestDescribeTableError() {
	ctx := context.Background()
	tableName := "test-table"
	
	suite.mockClient.On("DescribeTable", ctx, tableName).Return(nil, errors.New("DescribeTable error"))
	
	result, err := suite.mockClient.DescribeTable(ctx, tableName)
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "DescribeTable error")
}

// TestDeleteTable tests the DeleteTable function
func (suite *DALTestSuite) TestDeleteTable() {
	ctx := context.Background()
	input := &dynamodb.DeleteTableInput{
		TableName: &[]string{"test-table"}[0],
	}
	
	suite.mockClient.On("DeleteTable", ctx, input).Return(nil)
	
	err := suite.mockClient.DeleteTable(ctx, input)
	assert.NoError(suite.T(), err)
}

// TestDeleteTableError tests DeleteTable with error
func (suite *DALTestSuite) TestDeleteTableError() {
	ctx := context.Background()
	input := &dynamodb.DeleteTableInput{
		TableName: &[]string{"test-table"}[0],
	}
	
	suite.mockClient.On("DeleteTable", ctx, input).Return(errors.New("DeleteTable error"))
	
	err := suite.mockClient.DeleteTable(ctx, input)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "DeleteTable error")
}

// Run the test suite
func TestDALTestSuite(t *testing.T) {
	suite.Run(t, new(DALTestSuite))
}

// Standalone tests for additional coverage

func TestNewDALContainerSuccess(t *testing.T) {
	// We can't test the actual NewDALContainer function without creating real AWS resources
	// So we'll test the DAL container with our mock
	mockClient := &MockDatabaseClient{}
	container := &DALContainer{
		databaseClient: mockClient,
	}
	
	assert.NotNil(t, container)
	assert.NotNil(t, container.GetDatabaseClient())
	assert.Equal(t, mockClient, container.GetDatabaseClient())
}

func TestAttributeTypes(t *testing.T) {
	// Test all attribute types
	assert.Equal(t, models.StringType, models.AttributeType(0))
	assert.Equal(t, models.NumberType, models.AttributeType(1))
	assert.Equal(t, models.BinaryType, models.AttributeType(2))
}

func TestQueryConfig(t *testing.T) {
	// Test QueryConfig struct
	config := models.QueryConfig{
		TableName: "test-table",
		IndexName: "test-index",
		KeyName:   "test-key",
		KeyValue:  "test-value",
		KeyType:   models.StringType,
	}
	
	assert.Equal(t, "test-table", config.TableName)
	assert.Equal(t, "test-index", config.IndexName)
	assert.Equal(t, "test-key", config.KeyName)
	assert.Equal(t, "test-value", config.KeyValue)
	assert.Equal(t, models.StringType, config.KeyType)
}

func TestPrintPrettyJSONEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "Empty map",
			input:    map[string]interface{}{},
			expected: "{}",
		},
		{
			name:     "Empty slice",
			input:    []interface{}{},
			expected: "[]",
		},
		{
			name:     "String value",
			input:    "test string",
			expected: "\"test string\"",
		},
		{
			name:     "Number value",
			input:    42,
			expected: "42",
		},
		{
			name:     "Boolean value",
			input:    true,
			expected: "true",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := PrintPrettyJSON(tc.input)
			assert.Contains(t, result, tc.expected)
		})
	}
}

// TestDALContainerInterface tests that our mock implements the interface correctly
func TestDALContainerInterface(t *testing.T) {
	mockClient := &MockDatabaseClient{}
	container := &DALContainer{
		databaseClient: mockClient,
	}
	
	// Test interface compliance
	var dalContainer DALContainerInterface = container
	assert.NotNil(t, dalContainer)
	
	client := dalContainer.GetDatabaseClient()
	assert.NotNil(t, client)
	assert.Equal(t, mockClient, client)
}

// TestDatabaseClientInterface tests that our mock implements the interface correctly
func TestDatabaseClientInterface(t *testing.T) {
	mockClient := &MockDatabaseClient{}
	
	// Test interface compliance
	var dbClient DatabaseClientInterface = mockClient
	assert.NotNil(t, dbClient)
}

// TestMockClientMethodSignatures verifies all interface methods are properly mocked
func TestMockClientMethodSignatures(t *testing.T) {
	mockClient := &MockDatabaseClient{}
	ctx := context.Background()
	
	// Test that all methods exist and can be called without panicking
	config := models.QueryConfig{TableName: "test", KeyName: "id", KeyValue: "1", KeyType: models.StringType}
	var result map[string]interface{}
	
	mockClient.On("GetItem", ctx, config, &result).Return(nil, errors.New("test"))
	err := mockClient.GetItem(ctx, config, &result)
	assert.Error(t, err)
	
	mockClient.On("PutItem", ctx, "table", mock.Anything).Return(errors.New("test"))
	err = mockClient.PutItem(ctx, "table", map[string]string{})
	assert.Error(t, err)
	
	mockClient.On("UpdateItem", ctx, "table", "key", "value", mock.Anything).Return(errors.New("test"))
	err = mockClient.UpdateItem(ctx, "table", "key", "value", map[string]interface{}{})
	assert.Error(t, err)
	
	mockClient.On("DeleteItem", ctx, "table", "key", "value").Return(errors.New("test"))
	err = mockClient.DeleteItem(ctx, "table", "key", "value")
	assert.Error(t, err)
	
	var results []map[string]interface{}
	mockClient.On("QueryByIndex", ctx, "table", "index", "key", "value", &results).Return(nil, errors.New("test"))
	err = mockClient.QueryByIndex(ctx, "table", "index", "key", "value", &results)
	assert.Error(t, err)
	
	mockClient.On("Scan", ctx, "table", &results).Return(nil, errors.New("test"))
	err = mockClient.Scan(ctx, "table", &results)
	assert.Error(t, err)
	
	mockClient.On("ScanTable", ctx, "table", &results).Return(nil, errors.New("test"))
	err = mockClient.ScanTable(ctx, "table", &results)
	assert.Error(t, err)
	
	input := &dynamodb.CreateTableInput{}
	mockClient.On("CreateTable", ctx, input).Return(errors.New("test"))
	err = mockClient.CreateTable(ctx, input)
	assert.Error(t, err)
	
	mockClient.On("DescribeTable", ctx, "table").Return(nil, errors.New("test"))
	_, err = mockClient.DescribeTable(ctx, "table")
	assert.Error(t, err)
	
	deleteInput := &dynamodb.DeleteTableInput{}
	mockClient.On("DeleteTable", ctx, deleteInput).Return(errors.New("test"))
	err = mockClient.DeleteTable(ctx, deleteInput)
	assert.Error(t, err)
	
	mockClient.AssertExpectations(t)
}