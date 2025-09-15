package infrastructure

import (
	_ "embed"
	"encoding/json"
	"fieldfuze-backend/dal"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/tidwall/gjson"
)

// Your existing structs
type TableSchema struct {
	TableName              string                 `json:"TableName"`
	AttributeDefinitions   []AttributeDefinition  `json:"AttributeDefinitions"`
	KeySchema              []KeySchemaElement     `json:"KeySchema"`
	ProvisionedThroughput  Throughput             `json:"ProvisionedThroughput"`
	GlobalSecondaryIndexes []GlobalSecondaryIndex `json:"GlobalSecondaryIndexes,omitempty"`
}

type AttributeDefinition struct {
	AttributeName string `json:"AttributeName"`
	AttributeType string `json:"AttributeType"`
}

type KeySchemaElement struct {
	AttributeName string `json:"AttributeName"`
	KeyType       string `json:"KeyType"`
}

type Throughput struct {
	ReadCapacityUnits  int64 `json:"ReadCapacityUnits"`
	WriteCapacityUnits int64 `json:"WriteCapacityUnits"`
}

type GlobalSecondaryIndex struct {
	IndexName             string             `json:"IndexName"`
	KeySchema             []KeySchemaElement `json:"KeySchema"`
	Projection            Projection         `json:"Projection"`
	ProvisionedThroughput Throughput         `json:"ProvisionedThroughput"`
}

type Projection struct {
	ProjectionType string `json:"ProjectionType"`
}

//go:embed table_schema.json
var tablesSchema []byte

func GetTables(tableName string) (*dynamodb.CreateTableInput, error) {
	// Map the prefixed table name back to the schema key
	// For example, "dev_users1" -> "users1"
	schemaKey := extractBaseTableName(tableName)

	// Extract table data using gjson with the schema key
	tableJson := gjson.Get(string(tablesSchema), schemaKey)
	if !tableJson.Exists() {
		return nil, fmt.Errorf("table schema not found for key: %s", schemaKey)
	}

	fmt.Println("table schema for", schemaKey, "::", dal.PrintPrettyJSON(tableJson))

	// Unmarshal directly into TableSchema struct
	var schema TableSchema
	if err := json.Unmarshal([]byte(tableJson.Raw), &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema JSON: %w", err)
	}

	// Override the table name with the actual table name (including prefix)
	schema.TableName = tableName

	fmt.Println("final table schema for", tableName, "::", dal.PrintPrettyJSON(schema))
	return schema.ToDynamoInput(), nil
}

// extractBaseTableName extracts the base table name from a prefixed table name
// For example, "dev_users1" -> "users1", "prod_orders" -> "orders"
func extractBaseTableName(tableName string) string {
	parts := strings.Split(tableName, "_")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return tableName
}

// Convert our schema to DynamoDB input
func (ts *TableSchema) ToDynamoInput() *dynamodb.CreateTableInput {
	attrDefs := []*types.AttributeDefinition{}
	for _, a := range ts.AttributeDefinitions {
		attrDefs = append(attrDefs, &types.AttributeDefinition{
			AttributeName: aws.String(a.AttributeName),
			AttributeType: types.ScalarAttributeType(a.AttributeType),
		})
	}

	keySchema := []*types.KeySchemaElement{}
	for _, k := range ts.KeySchema {
		keySchema = append(keySchema, &types.KeySchemaElement{
			AttributeName: aws.String(k.AttributeName),
			KeyType:       types.KeyType(k.KeyType),
		})
	}

	gsis := []*types.GlobalSecondaryIndex{}
	for _, g := range ts.GlobalSecondaryIndexes {
		gsikeySchema := []*types.KeySchemaElement{}
		for _, k := range g.KeySchema {
			gsikeySchema = append(gsikeySchema, &types.KeySchemaElement{
				AttributeName: aws.String(k.AttributeName),
				KeyType:       types.KeyType(k.KeyType),
			})
		}
		// Convert []*types.KeySchemaElement to []types.KeySchemaElement
		var gsikeySchemaVals []types.KeySchemaElement
		for _, k := range gsikeySchema {
			gsikeySchemaVals = append(gsikeySchemaVals, *k)
		}
		gsis = append(gsis, &types.GlobalSecondaryIndex{
			IndexName: aws.String(g.IndexName),
			KeySchema: gsikeySchemaVals,
			Projection: &types.Projection{
				ProjectionType: types.ProjectionType(g.Projection.ProjectionType),
			},
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(g.ProvisionedThroughput.ReadCapacityUnits),
				WriteCapacityUnits: aws.Int64(g.ProvisionedThroughput.WriteCapacityUnits),
			},
		})
	}

	// Convert []*types.AttributeDefinition to []types.AttributeDefinition
	var attrDefsVals []types.AttributeDefinition
	for _, a := range attrDefs {
		attrDefsVals = append(attrDefsVals, *a)
	}

	// Convert []*types.KeySchemaElement to []types.KeySchemaElement
	var keySchemaVals []types.KeySchemaElement
	for _, k := range keySchema {
		keySchemaVals = append(keySchemaVals, *k)
	}

	// Convert []*types.GlobalSecondaryIndex to []types.GlobalSecondaryIndex
	var gsisVals []types.GlobalSecondaryIndex
	for _, g := range gsis {
		gsisVals = append(gsisVals, *g)
	}

	return &dynamodb.CreateTableInput{
		TableName:            aws.String(ts.TableName),
		AttributeDefinitions: attrDefsVals,
		KeySchema:            keySchemaVals,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(ts.ProvisionedThroughput.ReadCapacityUnits),
			WriteCapacityUnits: aws.Int64(ts.ProvisionedThroughput.WriteCapacityUnits),
		},
		GlobalSecondaryIndexes: gsisVals,
	}
}
