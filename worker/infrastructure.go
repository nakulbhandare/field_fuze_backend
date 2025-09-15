package worker

import (
	"context"
	"errors"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/infrastructure"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go"
)

type InfrastructureSetup struct {
	InfrastructureSetup models.InfrastructureSetup
}

// ToModelsInfrastructureSetup returns the embedded models.InfrastructureSetup
func (sm *InfrastructureSetup) ToModelsInfrastructureSetup() *models.InfrastructureSetup {
	return &sm.InfrastructureSetup
}

// NewInfrastructureSetup creates a new infrastructure setup handler
func NewInfrastructureSetup(cfg *models.Config, log logger.Logger) (*InfrastructureSetup, error) {
	// Create DynamoDB client
	dbClient, err := dal.NewDynamoDBClient(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create DynamoDB client: %w", err)
	}

	return &InfrastructureSetup{
		InfrastructureSetup: models.InfrastructureSetup{
			Config:   cfg,
			Logger:   log,
			DBClient: dbClient,
		},
	}, nil
}

func (is *InfrastructureSetup) Execute(ctx context.Context, statusManager *StatusManager) error {
	is.InfrastructureSetup.Logger.Info("Starting infrastructure setup...")

	if err := statusManager.UpdateProgress(models.StatusRunning, "Starting infrastructure setup", nil); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	tableDetails := is.getTableDetails()

	fmt.Println("Tables to process :: ", dal.PrintPrettyJSON(tableDetails))

	// Create tables sequentially to avoid throttling
	for _, tableInfo := range tableDetails {
		if err := is.createTableWithRetry(ctx, tableInfo); err != nil {
			is.InfrastructureSetup.Logger.Errorf("Failed to create table %s: %v", tableInfo.Name, err)
			statusManager.MarkFailed(fmt.Sprintf("Failed to create table %s: %v", tableInfo.Name, err))
			return err
		}

		// Record successful table creation
		statusManager.AddTableCreated(tableInfo.Name)
		is.InfrastructureSetup.Logger.Infof("✅ Successfully created table: %s", tableInfo.Name)
	}

	return nil
}

func (is *InfrastructureSetup) getTableDetails() []*models.TableInfo {
	var tableDetails []*models.TableInfo

	for _, tableName := range is.InfrastructureSetup.Config.Tables {

		fmt.Println("Setting Up the Table :::", tableName)

		tableDetails = append(tableDetails, &models.TableInfo{
			Name:   is.InfrastructureSetup.Config.DynamoDBTablePrefix + "_" + tableName,
			Status: "CREATING",
			Tags: map[string]string{
				"Environment": is.InfrastructureSetup.Config.AppEnv,
				"Application": is.InfrastructureSetup.Config.AppName,
				"TableType":   tableName,
				"CreatedBy":   "infrastructure-worker",
				"Version":     is.InfrastructureSetup.Config.AppVersion,
				"Service":     "FieldFuze",
			},
			CreatedAt:   time.Now(),
			IndexCount:  3, // email, username, verification-token indexes
			Indexes:     is.getTableIndexes(tableName),
			BillingMode: is.getBillingMode(),
			ParseName:   tableName,
		})

	}

	return tableDetails
}

// Helper methods for environment-specific configuration
func (is *InfrastructureSetup) getBillingMode() string {
	if is.InfrastructureSetup.Config.AppEnv == "prod" {
		return "PROVISIONED"
	}
	return "PAY_PER_REQUEST"
}

func (is *InfrastructureSetup) getTableIndexes(tableName string) []string {
	switch tableName {
	case "users":
		return []string{"id", "email", "username"}
	default:
		return []string{}
	}
}

// createTableWithRetry creates a table with retry logic
func (is *InfrastructureSetup) createTableWithRetry(ctx context.Context, tableInfo *models.TableInfo) error {
	maxRetries := 3
	baseDelay := 5 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := time.Duration(attempt) * baseDelay
			is.InfrastructureSetup.Logger.Infof("Retrying table creation for %s in %v (attempt %d/%d)", tableInfo.Name, delay, attempt+1, maxRetries+1)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Check if table already exists
		if exists, err := is.tableExists(tableInfo.Name); err != nil {
			is.InfrastructureSetup.Logger.Errorf("Failed to check if table exists: %v", err)
			continue
		} else if exists {
			is.InfrastructureSetup.Logger.Infof("✅ Table %s already exists, skipping creation", tableInfo.Name)
			return nil
		}

		// Create the table
		if err := is.createTableFromEmbeddedJSON(tableInfo.Name); err != nil {
			is.InfrastructureSetup.Logger.Errorf("Attempt %d failed to create table %s: %v", attempt+1, tableInfo.Name, err)

			if attempt == maxRetries {
				return fmt.Errorf("failed to create table %s after %d attempts: %w", tableInfo.Name, maxRetries+1, err)
			}
			continue
		}

		// Success
		return nil
	}

	return fmt.Errorf("exhausted all retry attempts for table %s", tableInfo.Name)
}

func (is *InfrastructureSetup) createTableFromEmbeddedJSON(tableName string) error {
	input, err := infrastructure.GetTables(tableName)
	if err != nil {
		return fmt.Errorf("failed to get table input: %w", err)
	}
	err = is.InfrastructureSetup.DBClient.CreateTable(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	fmt.Printf("✅ Table %s created successfully :: %v\n", tableName, input)
	return nil
}

// isTableNotFoundError checks if error indicates table not found
func isTableNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for AWS service error using smithy-go
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == "ResourceNotFoundException"
	}

	// Fallback to string matching for other error types
	errorStr := err.Error()
	return strings.Contains(errorStr, "ResourceNotFoundException") ||
		strings.Contains(errorStr, "Table not found") ||
		strings.Contains(errorStr, "Requested resource not found")
}

// tableExists checks if a table already exists
func (is *InfrastructureSetup) tableExists(tableName string) (bool, error) {
	_, err := is.InfrastructureSetup.DBClient.DescribeTable(context.Background(), tableName)
	if err != nil {
		// Check if error is "table not found"
		if isTableNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// deleteTableWithRetry deletes a table with retry logic
func (is *InfrastructureSetup) deleteTableWithRetry(ctx context.Context, tableName string, statusManager *StatusManager) error {
	maxRetries := 3
	baseDelay := 5 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := time.Duration(attempt) * baseDelay
			is.InfrastructureSetup.Logger.Infof("Retrying table deletion for %s in %v (attempt %d/%d)", tableName, delay, attempt+1, maxRetries+1)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Check if table exists
		exists, err := is.tableExists(tableName)
		if err != nil {
			is.InfrastructureSetup.Logger.Errorf("Failed to check if table exists: %v", err)
			continue
		} else if !exists {
			is.InfrastructureSetup.Logger.Infof("✅ Table %s does not exist, skipping deletion", tableName)
			return nil
		}

		// Delete the table
		if err := is.deleteTable(tableName); err != nil {
			is.InfrastructureSetup.Logger.Errorf("Attempt %d failed to delete table %s: %v", attempt+1, tableName, err)

			if attempt == maxRetries {
				return fmt.Errorf("failed to delete table %s after %d attempts: %w", tableName, maxRetries+1, err)
			}
			continue
		}

		// Success
		return nil
	}

	return fmt.Errorf("exhausted all retry attempts for table %s deletion", tableName)
}

// deleteTable deletes a specific table
func (is *InfrastructureSetup) deleteTable(tableName string) error {
	is.InfrastructureSetup.Logger.Warnf("Deleting table: %s", tableName)

	input := &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}

	return is.InfrastructureSetup.DBClient.DeleteTable(context.Background(), input)
}

// waitForTablesDeleted waits for all tables to be deleted
func (is *InfrastructureSetup) waitForTablesDeleted(ctx context.Context, tables []*models.TableInfo, statusManager *StatusManager) error {
	timeout := 10 * time.Minute
	checkInterval := 10 * time.Second

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// First, do an immediate check to see if tables are already deleted
	allDeleted := true
	for _, table := range tables {
		exists, err := is.tableExists(table.Name)
		if err != nil {
			// If we can't check, assume it still exists
			is.InfrastructureSetup.Logger.Errorf("Failed to check table %s existence: %v", table.Name, err)
			allDeleted = false
			break
		}

		if exists {
			is.InfrastructureSetup.Logger.Debugf("Table %s still exists", table.Name)
			allDeleted = false
			break
		}
	}

	if allDeleted {
		is.InfrastructureSetup.Logger.Info("All tables have been deleted")
		return nil
	}

	// If not all deleted, continue with polling
	for {
		select {
		case <-timeoutCtx.Done():
			// Log current status of all tables for debugging
			for _, table := range tables {
				if exists, err := is.tableExists(table.Name); err == nil && exists {
					is.InfrastructureSetup.Logger.Errorf("Table %s still exists after timeout", table.Name)
				}
			}
			return fmt.Errorf("timeout waiting for tables to be deleted")
		case <-ticker.C:
			allDeleted := true

			for _, table := range tables {
				exists, err := is.tableExists(table.Name)
				if err != nil {
					is.InfrastructureSetup.Logger.Errorf("Failed to check table %s existence: %v", table.Name, err)
					allDeleted = false
					continue
				}

				if exists {
					is.InfrastructureSetup.Logger.Debugf("Table %s still exists", table.Name)
					allDeleted = false
					break
				}
			}

			if allDeleted {
				is.InfrastructureSetup.Logger.Info("All tables have been deleted")
				return nil
			}

			// Update progress
			statusManager.UpdateProgress(models.StatusRunning, "Waiting for tables to be deleted", nil)
		}
	}
}

// validateInfrastructure validates the created infrastructure
func (is *InfrastructureSetup) validateInfrastructure(ctx context.Context, tables []*models.TableInfo) error {
	is.InfrastructureSetup.Logger.Info("Validating infrastructure setup")

	for _, table := range tables {
		// Check table status
		desc, err := is.InfrastructureSetup.DBClient.DescribeTable(ctx, table.Name)
		if err != nil {
			return fmt.Errorf("table %s validation failed: %w", table.Name, err)
		}

		if desc.Table.TableStatus != "ACTIVE" {
			return fmt.Errorf("table %s is not active: %s", table.Name, desc.Table.TableStatus)
		}

		// Verify indexes
		expectedIndexes := table.IndexCount
		actualIndexes := len(desc.Table.GlobalSecondaryIndexes)

		if actualIndexes != expectedIndexes {
			return fmt.Errorf("table %s has %d indexes, expected %d", table.Name, actualIndexes, expectedIndexes)
		}

		is.InfrastructureSetup.Logger.Infof("Table %s validation passed", table.Name)
	}

	is.InfrastructureSetup.Logger.Info("Infrastructure validation completed successfully")
	return nil
}

// ExecuteDelete runs the infrastructure deletion
func (is *InfrastructureSetup) ExecuteDelete(ctx context.Context, statusManager *StatusManager) error {
	is.InfrastructureSetup.Logger.Warn("Starting infrastructure deletion execution")

	// Update status
	if err := statusManager.UpdateProgress(models.StatusDeleting, "Starting infrastructure deletion", nil); err != nil {
		is.InfrastructureSetup.Logger.Errorf("Failed to update status: %v", err)
	}

	// Get required tables for environment
	tables := is.getTableDetails()

	// Delete tables sequentially
	for _, tableInfo := range tables {
		if err := is.deleteTableWithRetry(ctx, tableInfo.Name, statusManager); err != nil {
			is.InfrastructureSetup.Logger.Errorf("Failed to delete table %s: %v", tableInfo.Name, err)
			statusManager.UpdateProgress(models.StatusDeletionFailed, fmt.Sprintf("Failed to delete table %s: %v", tableInfo.Name, err), nil)
			return err
		}

		is.InfrastructureSetup.Logger.Warnf("🗑️ Successfully deleted table: %s", tableInfo.Name)
	}

	// Wait for all tables to be deleted
	if err := is.waitForTablesDeleted(ctx, tables, statusManager); err != nil {
		statusManager.UpdateProgress(models.StatusDeletionFailed, fmt.Sprintf("Tables failed to be deleted: %v", err), nil)
		return err
	}

	// Mark as completed
	if err := statusManager.UpdateProgress(models.StatusDeleted, "Infrastructure deletion completed", map[string]any{
		"deleted_tables": len(tables),
		"completed_at":   time.Now(),
	}); err != nil {
		is.InfrastructureSetup.Logger.Errorf("Failed to mark deletion as completed: %v", err)
	}

	is.InfrastructureSetup.Logger.Warn("🗑️ Infrastructure deletion completed successfully! All tables deleted.")
	return nil
}
