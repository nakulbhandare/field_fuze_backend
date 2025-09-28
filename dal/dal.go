package dal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fieldfuze-backend/models"
	"fmt"

	"fieldfuze-backend/utils/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient implements DatabaseClientInterface
type DynamoDBClient struct {
	client *dynamodb.Client
	config *models.Config
	logger logger.Logger
}

// DALContainer implements DALContainerInterface
type DALContainer struct {
	databaseClient DatabaseClientInterface
}

// printPrettyJSON takes any struct or map and prints it as pretty JSON
func PrintPrettyJSON(data interface{}) string {
	prettyJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Sprintf("Failed to generate JSON: %v", err)
	}
	return string(prettyJSON)
}

// NewDALContainer creates a new DAL container with DynamoDB client
func NewDALContainer(cfg *models.Config, log logger.Logger) (DALContainerInterface, error) {
	dbClient, err := NewDynamoDBClient(cfg, log)
	if err != nil {
		return nil, err
	}
	
	return &DALContainer{
		databaseClient: dbClient,
	}, nil
}

// GetDatabaseClient returns the database client interface
func (d *DALContainer) GetDatabaseClient() DatabaseClientInterface {
	return d.databaseClient
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(cfg *models.Config, log logger.Logger) (*DynamoDBClient, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Override endpoint for local DynamoDB
	if cfg.DynamoDBEndpoint != "" {
		awsCfg.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           cfg.DynamoDBEndpoint,
				SigningRegion: cfg.AWSRegion,
			}, nil
		})
	}

	// Use static credentials if provided
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		awsCfg.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"", // session token
		))
	}

	client := dynamodb.NewFromConfig(awsCfg)

	dbClient := &DynamoDBClient{
		client: client,
		config: cfg,
		logger: log,
	}

	log.Info("✅ DynamoDB client initialized successfully")
	return dbClient, nil
}

// GetItem - Universal method for any table, any primary key or secondary index
func (db *DynamoDBClient) GetItem(ctx context.Context, config models.QueryConfig, result interface{}) error {
	// If IndexName is provided, use Query for secondary index
	if config.IndexName != "" {
		return db.getSingleItemByIndex(ctx, config, result)
	}

	// Otherwise, use GetItem for primary key
	return db.getSingleItemByPrimaryKey(ctx, config, result)
}

// getSingleItemByPrimaryKey retrieves item by primary key
func (db *DynamoDBClient) getSingleItemByPrimaryKey(ctx context.Context, config models.QueryConfig, result interface{}) error {
	key := map[string]types.AttributeValue{
		config.KeyName: db.buildAttributeValue(config.KeyValue, config.KeyType),
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.TableName),
		Key:       key,
	}

	output, err := db.client.GetItem(ctx, input)
	if err != nil {
		db.logger.Errorf("Failed to get item from %s by %s=%s: %v",
			config.TableName, config.KeyName, config.KeyValue, err)
		return err
	}

	// db.logger.Infof("DynamoDB GetItem output: %s", PrintPrettyJSON(output))

	if output.Item == nil {
		return fmt.Errorf("item not found in %s with %s=%s",
			config.TableName, config.KeyName, config.KeyValue)
	}

	if err := attributevalue.UnmarshalMap(output.Item, result); err != nil {
		db.logger.Errorf("Failed to unmarshal item: %v", err)
		return err
	}

	return nil
}

// getSingleItemByIndex retrieves item by secondary index
func (db *DynamoDBClient) getSingleItemByIndex(ctx context.Context, config models.QueryConfig, result interface{}) error {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(config.TableName),
		IndexName:              aws.String(config.IndexName),
		Limit:                  aws.Int32(1), // Only get one item
		KeyConditionExpression: aws.String("#kn0 = :kv0"),
		ExpressionAttributeNames: map[string]string{
			"#kn0": config.KeyName,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":kv0": db.buildAttributeValue(config.KeyValue, config.KeyType),
		},
	}

	output, err := db.client.Query(ctx, input)
	if err != nil {
		db.logger.Errorf("Failed to query item from %s by %s=%s using index %s: %v",
			config.TableName, config.KeyName, config.KeyValue, config.IndexName, err)
		return err
	}

	db.logger.Infof("DynamoDB Query output: %s", PrintPrettyJSON(output))

	if len(output.Items) == 0 {
		return fmt.Errorf("item not found in %s with %s=%s using index %s",
			config.TableName, config.KeyName, config.KeyValue, config.IndexName)
	}

	// Unmarshal the first item
	if err := attributevalue.UnmarshalMap(output.Items[0], result); err != nil {
		db.logger.Errorf("Failed to unmarshal item: %v", err)
		return err
	}

	return nil
}

// Helper method to build attribute values based on type
func (db *DynamoDBClient) buildAttributeValue(value string, attrType models.AttributeType) types.AttributeValue {
	switch attrType {
	case models.NumberType:
		return &types.AttributeValueMemberN{Value: value}
	case models.BinaryType:
		// Assuming value is base64 encoded
		data, _ := base64.StdEncoding.DecodeString(value)
		return &types.AttributeValueMemberB{Value: data}
	default: // StringType
		return &types.AttributeValueMemberS{Value: value}
	}
}

// PutItem stores an item in DynamoDB
func (db *DynamoDBClient) PutItem(ctx context.Context, tableName string, item interface{}) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	_, err = db.client.PutItem(ctx, input)
	return err
}

// UpdateItem updates an item in DynamoDB
func (db *DynamoDBClient) UpdateItem(ctx context.Context, tableName, key, keyValue string, updates map[string]interface{}) error {
	updateExpression := "SET "
	expressionAttributeNames := make(map[string]string)
	expressionAttributeValues := make(map[string]types.AttributeValue)

	i := 0
	for field, value := range updates {
		if i > 0 {
			updateExpression += ", "
		}

		attrName := "#" + field
		attrValue := ":" + field

		updateExpression += attrName + " = " + attrValue
		expressionAttributeNames[attrName] = field

		av, err := attributevalue.Marshal(value)
		if err != nil {
			return err
		}
		expressionAttributeValues[attrValue] = av
		i++
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			key: &types.AttributeValueMemberS{Value: keyValue},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              types.ReturnValueAllNew,
	}

	_, err := db.client.UpdateItem(ctx, input)
	return err
}

// DeleteItem deletes an item from DynamoDB
func (db *DynamoDBClient) DeleteItem(ctx context.Context, tableName, key, value string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			key: &types.AttributeValueMemberS{Value: value},
		},
	}

	_, err := db.client.DeleteItem(ctx, input)
	return err
}

// QueryByIndex queries items using a global secondary index
func (db *DynamoDBClient) QueryByIndex(ctx context.Context, tableName, indexName, keyName, keyValue string, results interface{}) error {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String(indexName),
		Limit:                  aws.Int32(50),
		KeyConditionExpression: aws.String("#kn0 = :kv0"),
		ExpressionAttributeNames: map[string]string{
			"#kn0": keyName,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":kv0": &types.AttributeValueMemberS{Value: keyValue},
		},
	}

	output, err := db.client.Query(ctx, input)
	if err != nil {
		return err
	}

	return attributevalue.UnmarshalListOfMaps(output.Items, results)
}

// Scan scans the entire table
func (db *DynamoDBClient) Scan(ctx context.Context, tableName string, results interface{}) error {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	output, err := db.client.Scan(ctx, input)
	if err != nil {
		return err
	}

	return attributevalue.UnmarshalListOfMaps(output.Items, results)
}

// CreateTable creates a table
func (db *DynamoDBClient) CreateTable(ctx context.Context, input *dynamodb.CreateTableInput) error {
	_, err := db.client.CreateTable(ctx, input)
	return err
}

// DescribeTable describes a table
func (db *DynamoDBClient) DescribeTable(ctx context.Context, tableName string) (*dynamodb.DescribeTableOutput, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}
	return db.client.DescribeTable(ctx, input)
}

// DeleteTable deletes a table
func (db *DynamoDBClient) DeleteTable(ctx context.Context, input *dynamodb.DeleteTableInput) error {
	_, err := db.client.DeleteTable(ctx, input)
	return err
}

// ScanTable scans a table (alias for Scan)
func (db *DynamoDBClient) ScanTable(ctx context.Context, tableName string, results interface{}) error {
	return db.Scan(ctx, tableName, results)
}
