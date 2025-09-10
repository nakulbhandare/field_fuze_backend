package models

// AttributeType enum for different DynamoDB attribute types
type AttributeType int

const (
	StringType AttributeType = iota
	NumberType
	BinaryType
)

// QueryConfig holds all the configuration for any DynamoDB query
type QueryConfig struct {
	TableName string
	IndexName string // nil for primary key queries
	KeyName   string
	KeyValue  string
	KeyType   AttributeType // For different data types
}
