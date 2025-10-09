package dal

import (
	"context"
	"fieldfuze-backend/models"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DatabaseClientInterface defines the contract for database operations
type DatabaseClientInterface interface {
	// Core CRUD operations
	GetItem(ctx context.Context, config models.QueryConfig, result interface{}) error
	PutItem(ctx context.Context, tableName string, item interface{}) error
	UpdateItem(ctx context.Context, tableName, key, keyValue string, updates map[string]interface{}) error
	DeleteItem(ctx context.Context, tableName, key, value string) error
	
	// Query and Scan operations
	QueryByIndex(ctx context.Context, tableName, indexName, keyName, keyValue string, results interface{}) error
	Scan(ctx context.Context, tableName string, results interface{}) error
	ScanTable(ctx context.Context, tableName string, results interface{}) error
	
	// Table management operations
	CreateTable(ctx context.Context, input *dynamodb.CreateTableInput) error
	DescribeTable(ctx context.Context, tableName string) (*dynamodb.DescribeTableOutput, error)
	DeleteTable(ctx context.Context, input *dynamodb.DeleteTableInput) error
}

// DALContainerInterface defines the contract for the DAL container
type DALContainerInterface interface {
	GetDatabaseClient() DatabaseClientInterface
}