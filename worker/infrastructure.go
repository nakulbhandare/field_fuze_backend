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

	// Check if all tables already exist and are active
	allTablesExist := true
	var existingTables []string
	
	for _, tableInfo := range tableDetails {
		exists, err := is.tableExists(tableInfo.Name)
		if err != nil {
			is.InfrastructureSetup.Logger.Errorf("Failed to check if table %s exists: %v", tableInfo.Name, err)
			allTablesExist = false
			break
		}
		
		if exists {
			existingTables = append(existingTables, tableInfo.Name)
			statusManager.AddTableCreated(tableInfo.Name)
		} else {
			allTablesExist = false
		}
	}

	// If all tables exist, validate and handle accordingly
	if allTablesExist {
		is.InfrastructureSetup.Logger.Info("All required tables already exist, validating infrastructure...")
		
		if err := is.validateInfrastructure(ctx, tableDetails); err != nil {
			is.InfrastructureSetup.Logger.Errorf("Infrastructure validation failed: %v", err)
			
			// Handle validation failure - attempt to fix or recreate
			if err := is.handleValidationFailure(ctx, tableDetails, statusManager, err); err != nil {
				is.InfrastructureSetup.Logger.Errorf("Failed to handle validation failure: %v", err)
				statusManager.MarkFailed(fmt.Sprintf("Infrastructure validation failed and could not be fixed: %v", err))
				return err
			}
			
			// After handling failure, re-validate to ensure it's fixed
			if err := is.validateInfrastructure(ctx, tableDetails); err != nil {
				is.InfrastructureSetup.Logger.Errorf("Infrastructure still invalid after fix attempt: %v", err)
				statusManager.MarkFailed(fmt.Sprintf("Infrastructure validation failed after fix attempt: %v", err))
				return err
			}
			
			is.InfrastructureSetup.Logger.Info("✅ Infrastructure validation issues resolved successfully")
		} else {
			is.InfrastructureSetup.Logger.Info("✅ Infrastructure already exists and is valid")
		}
		
		return statusManager.MarkCompleted()
	}

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

	// Validate infrastructure after creation
	if err := is.validateInfrastructure(ctx, tableDetails); err != nil {
		is.InfrastructureSetup.Logger.Errorf("Infrastructure validation failed: %v", err)
		statusManager.MarkFailed(fmt.Sprintf("Infrastructure validation failed: %v", err))
		return err
	}

	// Mark as completed
	is.InfrastructureSetup.Logger.Info("🎉 Infrastructure setup completed successfully!")
	return statusManager.MarkCompleted()
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
			IndexCount:  is.getGSIIndexCount(tableName), // Only count Global Secondary Indexes
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

// getGSIIndexCount returns the expected count of Global Secondary Indexes for a table
func (is *InfrastructureSetup) getGSIIndexCount(tableName string) int {
	switch tableName {
	case "users1":
		return 2 // email-index, username-index
	case "role":
		return 2 // name-index, status-index
	default:
		return 0
	}
}

func (is *InfrastructureSetup) getTableIndexes(tableName string) []string {
	switch tableName {
	case "users1":
		return []string{"email-index", "username-index"} // Only GSI indexes
	case "role":
		return []string{"name-index", "status-index"} // Only GSI indexes
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
func (is *InfrastructureSetup) deleteTableWithRetry(ctx context.Context, tableName string, _ *StatusManager) error {
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

// validateInfrastructure validates the created infrastructure with intelligent waiting
func (is *InfrastructureSetup) validateInfrastructure(ctx context.Context, tables []*models.TableInfo) error {
	is.InfrastructureSetup.Logger.Info("Starting intelligent infrastructure validation...")
	
	// Phase 1: Wait for tables to become active
	if err := is.waitForTablesActive(ctx, tables); err != nil {
		return fmt.Errorf("tables did not become active: %w", err)
	}
	
	// Phase 2: Validate table configuration  
	return is.validateTableConfiguration(ctx, tables)
}

// waitForTablesActive waits for all tables to reach ACTIVE status
func (is *InfrastructureSetup) waitForTablesActive(ctx context.Context, tables []*models.TableInfo) error {
	maxWait := 10 * time.Minute
	checkInterval := 15 * time.Second
	
	timeoutCtx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()
	
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	
	is.InfrastructureSetup.Logger.Info("Waiting for DynamoDB tables to become active...")
	
	for {
		allActive := true
		var pendingTables []string
		
		for _, table := range tables {
			desc, err := is.InfrastructureSetup.DBClient.DescribeTable(ctx, table.Name)
			if err != nil {
				return fmt.Errorf("failed to describe table %s: %w", table.Name, err)
			}
			
			if desc.Table.TableStatus != "ACTIVE" {
				allActive = false
				pendingTables = append(pendingTables, fmt.Sprintf("%s(%s)", table.Name, desc.Table.TableStatus))
			}
		}
		
		if allActive {
			is.InfrastructureSetup.Logger.Info("All tables are now ACTIVE")
			return nil
		}
		
		is.InfrastructureSetup.Logger.Infof("Waiting for tables to become active: %v", pendingTables)
		
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for tables to become active after %v: %v", maxWait, pendingTables)
		case <-ticker.C:
			continue
		}
	}
}

// validateTableConfiguration validates table configuration after they're active
func (is *InfrastructureSetup) validateTableConfiguration(ctx context.Context, tables []*models.TableInfo) error {
	is.InfrastructureSetup.Logger.Info("Validating table configuration...")

	for _, table := range tables {
		// Check table status (should be ACTIVE at this point)
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

// handleValidationFailure handles infrastructure validation failures
// by attempting to fix issues or recreating problematic tables
func (is *InfrastructureSetup) handleValidationFailure(ctx context.Context, tableDetails []*models.TableInfo, statusManager *StatusManager, validationErr error) error {
	is.InfrastructureSetup.Logger.Warnf("Handling infrastructure validation failure: %v", validationErr)
	
	// Update status to indicate we're fixing validation issues
	if err := statusManager.UpdateProgress(models.StatusRunning, "Fixing infrastructure validation issues", map[string]any{
		"validation_error": validationErr.Error(),
		"fix_started_at":   time.Now(),
	}); err != nil {
		is.InfrastructureSetup.Logger.Errorf("Failed to update status: %v", err)
	}
	
	// Try to identify and fix specific validation issues
	for _, tableInfo := range tableDetails {
		if err := is.validateAndFixTable(ctx, tableInfo, nil); err != nil {
			is.InfrastructureSetup.Logger.Errorf("Failed to validate/fix table %s: %v", tableInfo.Name, err)
			
			// If we can't fix the table, try to recreate it
			is.InfrastructureSetup.Logger.Warnf("Attempting to recreate table %s due to validation failure", tableInfo.Name)
			
			// Delete the problematic table
			if err := is.deleteTableWithRetry(ctx, tableInfo.Name, nil); err != nil {
				is.InfrastructureSetup.Logger.Errorf("Failed to delete problematic table %s: %v", tableInfo.Name, err)
				return fmt.Errorf("failed to delete problematic table %s: %w", tableInfo.Name, err)
			}
			
			// Wait for table to be fully deleted
			if err := is.waitForTableDeleted(ctx, tableInfo.Name); err != nil {
				is.InfrastructureSetup.Logger.Errorf("Table %s failed to delete completely: %v", tableInfo.Name, err)
				return fmt.Errorf("table %s failed to delete: %w", tableInfo.Name, err)
			}
			
			// Recreate the table
			if err := is.createTableWithRetry(ctx, tableInfo); err != nil {
				is.InfrastructureSetup.Logger.Errorf("Failed to recreate table %s: %v", tableInfo.Name, err)
				return fmt.Errorf("failed to recreate table %s: %w", tableInfo.Name, err)
			}
			
			// Record successful table recreation
			statusManager.AddTableCreated(tableInfo.Name)
			is.InfrastructureSetup.Logger.Infof("✅ Successfully recreated table: %s", tableInfo.Name)
		}
	}
	
	is.InfrastructureSetup.Logger.Info("Infrastructure validation failure handling completed")
	return nil
}

// validateAndFixTable validates a specific table and attempts to fix issues
func (is *InfrastructureSetup) validateAndFixTable(ctx context.Context, tableInfo *models.TableInfo, _ *StatusManager) error {
	is.InfrastructureSetup.Logger.Debugf("Validating table: %s", tableInfo.Name)
	
	// Check table status
	desc, err := is.InfrastructureSetup.DBClient.DescribeTable(ctx, tableInfo.Name)
	if err != nil {
		return fmt.Errorf("table %s describe failed: %w", tableInfo.Name, err)
	}
	
	// Check if table is active
	if desc.Table.TableStatus != "ACTIVE" {
		is.InfrastructureSetup.Logger.Warnf("Table %s is not active: %s", tableInfo.Name, desc.Table.TableStatus)
		
		// Wait for table to become active (up to 5 minutes)
		timeout := 5 * time.Minute
		checkInterval := 10 * time.Second
		
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-timeoutCtx.Done():
				return fmt.Errorf("table %s did not become active within %v", tableInfo.Name, timeout)
			case <-ticker.C:
				desc, err := is.InfrastructureSetup.DBClient.DescribeTable(ctx, tableInfo.Name)
				if err != nil {
					return fmt.Errorf("failed to check table %s status: %w", tableInfo.Name, err)
				}
				
				if desc.Table.TableStatus == "ACTIVE" {
					is.InfrastructureSetup.Logger.Infof("Table %s is now active", tableInfo.Name)
					goto checkIndexes
				}
				
				is.InfrastructureSetup.Logger.Debugf("Table %s status: %s (waiting for ACTIVE)", tableInfo.Name, desc.Table.TableStatus)
			}
		}
	}
	
checkIndexes:
	// Verify indexes match expected count
	expectedIndexes := tableInfo.IndexCount
	actualIndexes := len(desc.Table.GlobalSecondaryIndexes)
	
	if actualIndexes != expectedIndexes {
		return fmt.Errorf("table %s has %d indexes, expected %d", tableInfo.Name, actualIndexes, expectedIndexes)
	}
	
	is.InfrastructureSetup.Logger.Debugf("Table %s validation passed", tableInfo.Name)
	return nil
}

// waitForTableDeleted waits for a single table to be deleted
func (is *InfrastructureSetup) waitForTableDeleted(ctx context.Context, tableName string) error {
	timeout := 10 * time.Minute
	checkInterval := 10 * time.Second
	
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	
	// First, do an immediate check
	exists, err := is.tableExists(tableName)
	if err != nil {
		is.InfrastructureSetup.Logger.Errorf("Failed to check table %s existence: %v", tableName, err)
	} else if !exists {
		is.InfrastructureSetup.Logger.Infof("Table %s has been deleted", tableName)
		return nil
	}
	
	// If still exists, continue with polling
	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for table %s to be deleted", tableName)
		case <-ticker.C:
			exists, err := is.tableExists(tableName)
			if err != nil {
				is.InfrastructureSetup.Logger.Errorf("Failed to check table %s existence: %v", tableName, err)
				continue
			}
			
			if !exists {
				is.InfrastructureSetup.Logger.Infof("Table %s has been deleted", tableName)
				return nil
			}
			
			is.InfrastructureSetup.Logger.Debugf("Table %s still exists, waiting for deletion", tableName)
		}
	}
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
