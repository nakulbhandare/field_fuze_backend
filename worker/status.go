package worker

import (
	"context"
	"encoding/json"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// StatusManager embeds models.StatusManager and adds AWS integration
type StatusManager struct {
	models.StatusManager
	dbClient  models.DBClient
	logger    logger.Logger
	fileMutex sync.Mutex // Protect file operations from concurrent access
}

// NewStatusManager creates a new status manager with AWS integration
func NewStatusManager(statusPath string, dbClient models.DBClient, logger logger.Logger) *StatusManager {
	return &StatusManager{
		StatusManager: models.StatusManager{
			StatusFilePath: statusPath,
		},
		dbClient: dbClient,
		logger:   logger,
	}
}

// NewStatusManagerLegacy creates a status manager without AWS integration (for backward compatibility)
func NewStatusManagerLegacy(statusPath string) *StatusManager {
	return &StatusManager{
		StatusManager: models.StatusManager{
			StatusFilePath: statusPath,
		},
		dbClient: nil,
		logger:   nil,
	}
}

// ToModelsStatusManager returns the embedded models.StatusManager
func (sm *StatusManager) ToModelsStatusManager() *models.StatusManager {
	return &sm.StatusManager
}

func (sm *StatusManager) SaveStatus(result *models.ExecutionResult) error {
	// Use mutex to prevent concurrent file operations
	sm.fileMutex.Lock()
	defer sm.fileMutex.Unlock()

	// Ensure status directory exists
	if err := os.MkdirAll(filepath.Dir(sm.StatusFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create status directory: %w", err)
	}

	// Update end time if not set
	if result.EndTime == nil && (result.Status == models.StatusCompleted || result.Status == models.StatusFailed) {
		now := time.Now()
		result.EndTime = &now
		result.Duration = now.Sub(result.StartTime)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// Create unique temporary file name to avoid race conditions
	tempFile := fmt.Sprintf("%s.tmp.%d", sm.StatusFilePath, time.Now().UnixNano())

	// Write to temporary file
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp status file: %w", err)
	}

	// Atomic rename to final file
	if err := os.Rename(tempFile, sm.StatusFilePath); err != nil {
		os.Remove(tempFile) // Clean up temp file on failure
		return fmt.Errorf("failed to rename status file: %w", err)
	}

	return nil
}

func (sm *StatusManager) LoadStatus() (*models.ExecutionResult, error) {
	// Use mutex to prevent concurrent file operations
	sm.fileMutex.Lock()
	defer sm.fileMutex.Unlock()

	data, err := os.ReadFile(sm.StatusFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read status file: %w", err)
	}

	var result models.ExecutionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	return &result, nil
}

// IsSetupCompleted checks if infrastructure setup is completed
func (sm *StatusManager) IsSetupCompleted() (bool, error) {
	status, err := sm.LoadStatus()
	if err != nil {
		return false, err
	}

	return status.Status == models.StatusCompleted && status.Success, nil
}

// GetLastExecutionTime returns the last execution time
func (sm *StatusManager) GetLastExecutionTime() (time.Time, error) {
	status, err := sm.LoadStatus()
	if err != nil {
		return time.Time{}, err
	}

	return status.StartTime, nil
}

func (sm *StatusManager) UpdateProgress(status models.WorkerStatus, message string, metadata map[string]any) error {
	currentStatus, err := sm.LoadStatus()
	if err != nil {
		// Create new status if loading fails
		currentStatus = &models.ExecutionResult{
			StartTime:     time.Now(),
			TablesCreated: make([]models.TableStatus, 0),
			Metadata:      make(map[string]any),
		}
	}

	currentStatus.Status = status
	if message != "" {
		if currentStatus.Metadata == nil {
			currentStatus.Metadata = make(map[string]any)
		}
		currentStatus.Metadata["last_message"] = message
		currentStatus.Metadata["last_update"] = time.Now()
	}

	// Merge metadata
	if metadata != nil {
		if currentStatus.Metadata == nil {
			currentStatus.Metadata = make(map[string]any)
		}
		for k, v := range metadata {
			currentStatus.Metadata[k] = v
		}
	}

	return sm.SaveStatus(currentStatus)
}

// AddTableCreated adds a table to the created list (legacy method)
func (sm *StatusManager) AddTableCreated(tableName string) error {
	// If we have AWS integration, use the enhanced method
	if sm.dbClient != nil && sm.logger != nil {
		return sm.AddTableWithAWSStatus(context.Background(), tableName)
	}

	// Fallback to legacy behavior
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	// Check if table already in list
	for _, table := range status.TablesCreated {
		if table.Name == tableName {
			return nil // Already recorded
		}
	}

	tableStatus := models.TableStatus{
		Name:      tableName,
		Status:    "CREATING",
		CreatedAt: time.Now(),
	}
	status.TablesCreated = append(status.TablesCreated, tableStatus)
	return sm.SaveStatus(status)
}

// AddTableWithAWSStatus adds or updates table status using real AWS DescribeTable data
func (sm *StatusManager) AddTableWithAWSStatus(ctx context.Context, tableName string) error {
	if sm.dbClient == nil {
		return sm.AddTableCreated(tableName) // Fallback to legacy
	}

	sm.logger.Debugf("Fetching AWS status for table: %s", tableName)

	// Fetch real table status from AWS
	fastFetcher := NewFastTableStatusFetcher(sm.dbClient, sm.logger)
	quickStatus, err := fastFetcher.GetTableStatusFast(ctx, tableName)
	if err != nil {
		sm.logger.Errorf("Failed to fetch AWS status for table %s: %v", tableName, err)
		return sm.addTableWithFallbackStatus(tableName, err)
	}

	// Load current status
	status, err := sm.LoadStatus()
	if err != nil {
		status = sm.createDefaultStatus()
	}

	// Update or add table status
	updated := false
	for i, table := range status.TablesCreated {
		if table.Name == tableName {
			// Update existing table with fresh AWS data including index details and ARN
			status.TablesCreated[i].Status = quickStatus.Status
			status.TablesCreated[i].Arn = quickStatus.Arn
			status.TablesCreated[i].IndexCount = quickStatus.IndexCount
			status.TablesCreated[i].IndexesCreated = quickStatus.IndexesCreated
			updated = true
			sm.logger.Debugf("Updated existing table status for %s: %s (ARN: %s) with %d indexes", tableName, quickStatus.Status, quickStatus.Arn, quickStatus.IndexCount)
			break
		}
	}

	if !updated {
		// Add new table status with index details and ARN
		newTable := models.TableStatus{
			Name:           tableName,
			Status:         quickStatus.Status,
			Arn:            quickStatus.Arn,
			CreatedAt:      quickStatus.LastStatusUpdate,
			IndexCount:     quickStatus.IndexCount,
			IndexesCreated: quickStatus.IndexesCreated,
		}
		status.TablesCreated = append(status.TablesCreated, newTable)
		sm.logger.Debugf("Added new table status for %s: %s (ARN: %s) with %d indexes", tableName, quickStatus.Status, quickStatus.Arn, quickStatus.IndexCount)
	}

	// Add AWS metadata to status
	if status.Metadata == nil {
		status.Metadata = make(map[string]interface{})
	}
	status.Metadata["last_aws_refresh"] = time.Now()
	status.Metadata["aws_check_latency"] = quickStatus.StatusCheckLatency

	return sm.SaveStatus(status)
}

// GetTableCurrentStatus fetches real-time status for a specific table
func (sm *StatusManager) GetTableCurrentStatus(ctx context.Context, tableName string) (*QuickTableStatus, error) {
	if sm.dbClient == nil {
		return nil, fmt.Errorf("no AWS client available for status check")
	}

	fastFetcher := NewFastTableStatusFetcher(sm.dbClient, sm.logger)
	return fastFetcher.GetTableStatusFast(ctx, tableName)
}

// RefreshAllTableStatuses refreshes status for all tracked tables using AWS DescribeTable API
func (sm *StatusManager) RefreshAllTableStatuses(ctx context.Context) error {
	if sm.dbClient == nil {
		return fmt.Errorf("no AWS client available for refresh")
	}

	status, err := sm.LoadStatus()
	if err != nil {
		return fmt.Errorf("failed to load current status: %w", err)
	}

	sm.logger.Debugf("Refreshing status for %d tracked tables", len(status.TablesCreated))

	for i, table := range status.TablesCreated {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		sm.logger.Debugf("Refreshing status for table: %s", table.Name)

		fastFetcher := NewFastTableStatusFetcher(sm.dbClient, sm.logger)
		quickStatus, err := fastFetcher.GetTableStatusFast(ctx, table.Name)
		if err != nil {
			sm.logger.Errorf("Failed to refresh status for table %s: %v", table.Name, err)
			// Mark table as having an error but don't fail the entire operation
			status.TablesCreated[i].Status = "ERROR"
			continue
		}

		// Update with fresh AWS data including index details and ARN
		status.TablesCreated[i].Status = quickStatus.Status
		status.TablesCreated[i].Arn = quickStatus.Arn
		status.TablesCreated[i].IndexCount = quickStatus.IndexCount
		status.TablesCreated[i].IndexesCreated = quickStatus.IndexesCreated
		sm.logger.Debugf("Refreshed table %s status: %s (ARN: %s) with %d indexes", table.Name, quickStatus.Status, quickStatus.Arn, quickStatus.IndexCount)
	}

	// Update metadata with refresh timestamp
	if status.Metadata == nil {
		status.Metadata = make(map[string]interface{})
	}
	status.Metadata["last_aws_refresh"] = time.Now()
	status.Metadata["refresh_method"] = "describe_table_api"

	return sm.SaveStatus(status)
}

// Helper methods

func (sm *StatusManager) addTableWithFallbackStatus(tableName string, awsError error) error {
	status, err := sm.LoadStatus()
	if err != nil {
		status = sm.createDefaultStatus()
	}

	// Check if table already exists in status
	for _, table := range status.TablesCreated {
		if table.Name == tableName {
			return nil // Already recorded
		}
	}

	// Add with fallback status
	fallbackTable := models.TableStatus{
		Name:      tableName,
		Status:    "AWS_ERROR",
		CreatedAt: time.Now(),
	}

	status.TablesCreated = append(status.TablesCreated, fallbackTable)

	// Add error info to metadata
	if status.Metadata == nil {
		status.Metadata = make(map[string]interface{})
	}
	status.Metadata[fmt.Sprintf("aws_error_%s", tableName)] = awsError.Error()

	return sm.SaveStatus(status)
}

func (sm *StatusManager) createDefaultStatus() *models.ExecutionResult {
	return &models.ExecutionResult{
		StartTime:     time.Now(),
		TablesCreated: make([]models.TableStatus, 0),
		Metadata:      make(map[string]interface{}),
	}
}

// MarkCompleted marks the setup as completed
func (sm *StatusManager) MarkCompleted() error {
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	status.Success = true
	status.Status = models.StatusCompleted
	now := time.Now()
	status.EndTime = &now
	status.Duration = now.Sub(status.StartTime)

	return sm.SaveStatus(status)
}

// MarkFailed marks the setup as failed
func (sm *StatusManager) MarkFailed(errorMsg string) error {
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	status.Success = false
	status.Status = models.StatusFailed
	status.ErrorMessage = errorMsg
	now := time.Now()
	status.EndTime = &now
	status.Duration = now.Sub(status.StartTime)

	return sm.SaveStatus(status)
}

// IncrementRetryCount increments the retry counter (stored in metadata)
func (sm *StatusManager) IncrementRetryCount() error {
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	// Store retry count in metadata instead
	if status.Metadata == nil {
		status.Metadata = make(map[string]interface{})
	}
	retryCount, ok := status.Metadata["retry_count"].(int)
	if !ok {
		retryCount = 0
	}
	status.Metadata["retry_count"] = retryCount + 1
	status.Status = models.StatusRetrying

	return sm.SaveStatus(status)
}

// GetRetryCount gets the current retry count from metadata
func (sm *StatusManager) GetRetryCount() (int, error) {
	status, err := sm.LoadStatus()
	if err != nil {
		return 0, err
	}

	if status.Metadata == nil {
		return 0, nil
	}
	retryCount, ok := status.Metadata["retry_count"].(int)
	if !ok {
		return 0, nil
	}
	return retryCount, nil
}

// ResetStatus resets the status (useful for forced re-runs)
func (sm *StatusManager) ResetStatus() error {
	return os.Remove(sm.StatusFilePath)
}

// updateTableStatusQuickly performs lightning-fast status update
func (sm *StatusManager) updateTableStatusQuickly(tableName string, quickStatus *QuickTableStatus) error {
	// Load current status
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	// Find and update the specific table
	updated := false
	for i, table := range status.TablesCreated {
		if table.Name == tableName {
			// Update essential fields including index details and ARN
			status.TablesCreated[i].Status = quickStatus.Status
			status.TablesCreated[i].Arn = quickStatus.Arn
			status.TablesCreated[i].IndexCount = quickStatus.IndexCount
			status.TablesCreated[i].IndexesCreated = quickStatus.IndexesCreated
			updated = true
			break
		}
	}

	if !updated {
		// Add as new table if not found with index details and ARN
		newTable := models.TableStatus{
			Name:           tableName,
			Status:         quickStatus.Status,
			Arn:            quickStatus.Arn,
			CreatedAt:      time.Now(),
			IndexCount:     quickStatus.IndexCount,
			IndexesCreated: quickStatus.IndexesCreated,
		}
		status.TablesCreated = append(status.TablesCreated, newTable)
	}

	// Update metadata with quick refresh info
	if status.Metadata == nil {
		status.Metadata = make(map[string]interface{})
	}
	status.Metadata["last_quick_refresh"] = time.Now()
	status.Metadata["quick_refresh_latency"] = quickStatus.StatusCheckLatency

	return sm.SaveStatus(status)
}

// =============================================================================
// Fast Table Status Fetcher
// =============================================================================

// FastTableStatusFetcher provides lightning-fast status-only operations
type FastTableStatusFetcher struct {
	dbClient models.DBClient
	logger   logger.Logger
}

// NewFastTableStatusFetcher creates a new fast status fetcher
func NewFastTableStatusFetcher(dbClient models.DBClient, logger logger.Logger) *FastTableStatusFetcher {
	return &FastTableStatusFetcher{
		dbClient: dbClient,
		logger:   logger,
	}
}

// QuickTableStatus contains essential status information with index details
type QuickTableStatus struct {
	Name               string                `json:"name"`
	Status             string                `json:"status"`          // CREATING, ACTIVE, UPDATING, etc.
	Arn                string                `json:"arn"`             // Table ARN from AWS
	IndexCount         int                   `json:"index_count"`     // Total count of indexes
	IndexesCreated     []models.IndexDetails `json:"indexes_created"` // Detailed index information with ARNs
	LastStatusUpdate   time.Time             `json:"last_status_update"`
	StatusCheckLatency time.Duration         `json:"status_check_latency"` // How long the check took
}

// GetTableStatusFast performs ultra-fast status check with minimal data
func (ftsf *FastTableStatusFetcher) GetTableStatusFast(ctx context.Context, tableName string) (*QuickTableStatus, error) {
	startTime := time.Now()

	// Use very short timeout for fast operations
	fastCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Call DescribeTable with minimal processing
	desc, err := ftsf.dbClient.DescribeTable(fastCtx, tableName)
	if err != nil {
		return nil, err
	}

	latency := time.Since(startTime)

	// Extract essential information with index details and table ARN
	indexDetails := ftsf.extractIndexDetails(desc.Table)

	quickStatus := &QuickTableStatus{
		Name:               tableName,
		Status:             string(desc.Table.TableStatus),
		Arn:                aws.ToString(desc.Table.TableArn),
		IndexCount:         len(indexDetails),
		IndexesCreated:     indexDetails,
		LastStatusUpdate:   time.Now(),
		StatusCheckLatency: latency,
	}

	ftsf.logger.Debugf("Fast status check for %s: %s (took %v)", tableName, quickStatus.Status, latency)

	return quickStatus, nil
}

// BatchGetTableStatusesFast gets status for multiple tables efficiently
func (ftsf *FastTableStatusFetcher) BatchGetTableStatusesFast(ctx context.Context, tableNames []string) map[string]*QuickTableStatus {
	results := make(map[string]*QuickTableStatus)

	// Use semaphore to limit concurrent calls (prevent overwhelming AWS)
	semaphore := make(chan struct{}, 3) // Max 3 concurrent calls
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, tableName := range tableNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Individual timeout per table
			tableCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
			defer cancel()

			status, err := ftsf.GetTableStatusFast(tableCtx, name)

			mu.Lock()
			if err != nil {
				ftsf.logger.Debugf("Failed to get fast status for %s: %v", name, err)
				// Store error info
				results[name] = &QuickTableStatus{
					Name:               name,
					Status:             "ERROR",
					LastStatusUpdate:   time.Now(),
					StatusCheckLatency: 0,
				}
			} else {
				results[name] = status
			}
			mu.Unlock()
		}(tableName)
	}

	// Wait for all with overall timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		ftsf.logger.Debugf("Batch status check completed for %d tables", len(tableNames))
	case <-ctx.Done():
		ftsf.logger.Warnf("Batch status check cancelled")
	case <-time.After(5 * time.Second):
		ftsf.logger.Warnf("Batch status check timeout")
	}

	return results
}

// extractIndexDetails extracts detailed index information including ARNs from AWS table description
func (ftsf *FastTableStatusFetcher) extractIndexDetails(table *types.TableDescription) []models.IndexDetails {
	var indexDetails []models.IndexDetails

	// Extract Global Secondary Indexes
	for _, gsi := range table.GlobalSecondaryIndexes {
		index := models.IndexDetails{
			Name:   aws.ToString(gsi.IndexName),
			Status: string(gsi.IndexStatus),
			Arn:    aws.ToString(gsi.IndexArn),
			Type:   "GSI",
		}

		// Set creation time (GSI doesn't have creation time directly, use current time)
		index.CreatedAt = time.Now()

		indexDetails = append(indexDetails, index)
	}

	// Extract Local Secondary Indexes
	for _, lsi := range table.LocalSecondaryIndexes {
		index := models.IndexDetails{
			Name:      aws.ToString(lsi.IndexName),
			Status:    "ACTIVE", // LSIs are always ACTIVE when table is ACTIVE
			Arn:       aws.ToString(lsi.IndexArn),
			Type:      "LSI",
			CreatedAt: time.Now(), // LSI doesn't provide creation time
		}

		indexDetails = append(indexDetails, index)
	}

	ftsf.logger.Debugf("Extracted %d index details for table %s", len(indexDetails), aws.ToString(table.TableName))
	return indexDetails
}

// =============================================================================
// Lightweight Status Refresher
// =============================================================================

// LightweightStatusRefresher provides fast, non-blocking status updates
type LightweightStatusRefresher struct {
	statusManager *StatusManager
	dbClient      models.DBClient
	logger        logger.Logger

	// Non-blocking configuration
	refreshInterval time.Duration
	maxRefreshTime  time.Duration // Maximum time allowed for a single refresh cycle

	// Async operation management
	refreshChan chan string    // Channel for refresh requests
	stopChan    chan struct{}  // Channel for stopping
	workerWg    sync.WaitGroup // Wait group for graceful shutdown
	isRunning   bool
	mu          sync.RWMutex
}

// NewLightweightStatusRefresher creates a fast, non-blocking status refresher
func NewLightweightStatusRefresher(statusManager *StatusManager, dbClient models.DBClient, logger logger.Logger) *LightweightStatusRefresher {
	return &LightweightStatusRefresher{
		statusManager:   statusManager,
		dbClient:        dbClient,
		logger:          logger,
		refreshInterval: 5 * time.Minute,        // Refresh every 5 minutes
		maxRefreshTime:  10 * time.Second,       // Never spend more than 10 seconds on refresh
		refreshChan:     make(chan string, 100), // Buffered channel for non-blocking requests
		stopChan:        make(chan struct{}),
	}
}

// Start begins the lightweight status refresher
func (lsr *LightweightStatusRefresher) Start(ctx context.Context) {
	lsr.mu.Lock()
	defer lsr.mu.Unlock()

	if lsr.isRunning {
		lsr.logger.Warn("Lightweight status refresher is already running")
		return
	}

	lsr.isRunning = true
	lsr.logger.Info("Starting lightweight non-blocking status refresher")

	// Start the refresh worker goroutine
	lsr.workerWg.Add(1)
	go lsr.refreshWorker(ctx)

	// Start the periodic scheduler goroutine
	lsr.workerWg.Add(1)
	go lsr.periodicScheduler(ctx)
}

// Stop gracefully stops the refresher
func (lsr *LightweightStatusRefresher) Stop() {
	lsr.mu.Lock()
	defer lsr.mu.Unlock()

	if !lsr.isRunning {
		return
	}

	lsr.logger.Info("Stopping lightweight status refresher")
	lsr.isRunning = false
	close(lsr.stopChan)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		lsr.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		lsr.logger.Info("Lightweight status refresher stopped gracefully")
	case <-time.After(5 * time.Second):
		lsr.logger.Warn("Lightweight status refresher stop timeout - forcing shutdown")
	}
}

// RequestRefresh asynchronously requests a status refresh for a specific table
func (lsr *LightweightStatusRefresher) RequestRefresh(tableName string) {
	select {
	case lsr.refreshChan <- tableName:
		lsr.logger.Debugf("Queued status refresh for table: %s", tableName)
	default:
		lsr.logger.Warnf("Refresh queue full, skipping refresh for table: %s", tableName)
	}
}

// RequestRefreshAll asynchronously requests refresh for all tracked tables
func (lsr *LightweightStatusRefresher) RequestRefreshAll() {
	lsr.RequestRefresh("__ALL_TABLES__") // Special marker for all tables
}

// periodicScheduler schedules regular refresh cycles
func (lsr *LightweightStatusRefresher) periodicScheduler(ctx context.Context) {
	defer lsr.workerWg.Done()

	ticker := time.NewTicker(lsr.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lsr.logger.Debug("Triggering periodic status refresh")
			lsr.RequestRefreshAll()
		case <-lsr.stopChan:
			lsr.logger.Debug("Periodic scheduler stopped")
			return
		case <-ctx.Done():
			lsr.logger.Debug("Periodic scheduler stopped due to context cancellation")
			return
		}
	}
}

// refreshWorker processes refresh requests asynchronously
func (lsr *LightweightStatusRefresher) refreshWorker(ctx context.Context) {
	defer lsr.workerWg.Done()

	for {
		select {
		case tableName := <-lsr.refreshChan:
			lsr.processRefreshRequest(ctx, tableName)
		case <-lsr.stopChan:
			lsr.logger.Debug("Refresh worker stopped")
			return
		case <-ctx.Done():
			lsr.logger.Debug("Refresh worker stopped due to context cancellation")
			return
		}
	}
}

// processRefreshRequest handles a single refresh request with timeout
func (lsr *LightweightStatusRefresher) processRefreshRequest(ctx context.Context, tableName string) {
	// Create a timeout context for this refresh operation
	refreshCtx, cancel := context.WithTimeout(ctx, lsr.maxRefreshTime)
	defer cancel()

	startTime := time.Now()

	if tableName == "__ALL_TABLES__" {
		lsr.refreshAllTablesLightweight(refreshCtx)
	} else {
		lsr.refreshSingleTableLightweight(refreshCtx, tableName)
	}

	duration := time.Since(startTime)
	lsr.logger.Debugf("Status refresh completed in %v for: %s", duration, tableName)

	// Warn if refresh took too long
	if duration > lsr.maxRefreshTime/2 {
		lsr.logger.Warnf("Status refresh took %v (approaching limit of %v) for table: %s",
			duration, lsr.maxRefreshTime, tableName)
	}
}

// refreshSingleTableLightweight performs fast status refresh for a single table
func (lsr *LightweightStatusRefresher) refreshSingleTableLightweight(ctx context.Context, tableName string) {
	// Use a very short timeout for individual table refresh
	tableCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Fetch only essential status information
	tableStatus, err := lsr.getTableStatusFast(tableCtx, tableName)
	if err != nil {
		lsr.logger.Debugf("Failed to refresh status for table %s: %v", tableName, err)
		return
	}

	// Update only the essential status fields without full processing
	if err := lsr.statusManager.updateTableStatusQuickly(tableName, tableStatus); err != nil {
		lsr.logger.Debugf("Failed to update status for table %s: %v", tableName, err)
	}
}

// refreshAllTablesLightweight performs fast refresh for all tracked tables
func (lsr *LightweightStatusRefresher) refreshAllTablesLightweight(ctx context.Context) {
	// Get list of tracked tables
	status, err := lsr.statusManager.LoadStatus()
	if err != nil {
		lsr.logger.Debugf("Failed to load status for refresh: %v", err)
		return
	}

	lsr.logger.Debugf("Refreshing status for %d tracked tables", len(status.TablesCreated))

	// Refresh each table with individual timeouts
	for _, table := range status.TablesCreated {
		// Check if we should stop
		select {
		case <-ctx.Done():
			lsr.logger.Debug("Refresh cancelled")
			return
		default:
		}

		// Refresh with individual timeout
		lsr.refreshSingleTableLightweight(ctx, table.Name)
	}
}

// getTableStatusFast performs ultra-fast status retrieval
func (lsr *LightweightStatusRefresher) getTableStatusFast(ctx context.Context, tableName string) (*QuickTableStatus, error) {
	fastFetcher := NewFastTableStatusFetcher(lsr.dbClient, lsr.logger)
	return fastFetcher.GetTableStatusFast(ctx, tableName)
}
