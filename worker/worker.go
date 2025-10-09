package worker

import (
	"context"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron"
)

// Worker
type Worker struct {
	Worker *models.Worker // Use pointer to avoid copying mutex
}

func NewWorker(ctx context.Context, cfg *models.Config, log logger.Logger) (*models.Worker, error) {
	defer ctx.Done()
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Generate unique owner ID for this instance
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

	ownerID := fmt.Sprintf("worker-%s-%s", hostname, uuid.New().String()[:8])

	// Create worker configuration
	workerConfig := &models.WorkerConfig{
		CronSchedule:      getCronScheduleForEnvironment(cfg.AppEnv),
		LockTimeout:       30 * time.Minute,
		LockRetryInterval: 5 * time.Second,
		MaxRetries:        5,
		RetryDelay:        2 * time.Second,
		BackoffMultiplier: 2.0,
		Environment:       cfg.AppEnv,
		RequiredTables:    []string{"users"},
		LockFilePath:      fmt.Sprintf("/tmp/fieldfuze-infrastructure-%s.lock", cfg.AppEnv),
		StatusFilePath:    fmt.Sprintf("/tmp/fieldfuze-status-%s.json", cfg.AppEnv),
		DryRun:            os.Getenv("INFRASTRUCTURE_DRY_RUN") == "true",
		SkipValidation:    os.Getenv("INFRASTRUCTURE_SKIP_VALIDATION") == "true",
		ForceRecreate:     os.Getenv("INFRASTRUCTURE_FORCE_RECREATE") == "true",
		RunOnce:           true,
	}

	log.Infof("Worker configuration: %+v", dal.PrintPrettyJSON(workerConfig))

	// Validate configuration
	if err := validateWorkerConfig(workerConfig); err != nil {
		return nil, fmt.Errorf("invalid worker configuration: %w", err)
	}

	// Create infrastructure setup handler first (needed for DB client)
	infrastructureSetup, err := NewInfrastructureSetup(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create infrastructure setup: %w", err)
	}

	// Initialize components with enhanced status manager
	lockManager := NewLockManager(workerConfig.LockFilePath, workerConfig.LockTimeout, workerConfig.Environment)
	// Create enhanced status manager with AWS integration for real-time status tracking
	statusManager := NewStatusManager(workerConfig.StatusFilePath, infrastructureSetup.InfrastructureSetup.DBClient, log)
	
	log.Info("Enhanced infrastructure worker initialized with lightweight AWS status tracking")

	// Create cron job with second precision
	cronJob := cron.New()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	return &models.Worker{
		Config:              cfg,
		Logger:              log,
		CronJob:             cronJob,
		LockManager:         lockManager,
		StatusManager:       statusManager.ToModelsStatusManager(),
		InfrastructureSetup: infrastructureSetup.ToModelsInfrastructureSetup(),
		WorkerConfig:        workerConfig,
		OwnerID:             ownerID,
		StopChan:            make(chan struct{}),
		Ctx:                 ctx,
		Cancel:              cancel,
	}, nil
}

// Start starts the infrastructure worker
func (w *Worker) Start() error {
	w.Worker.Logger.Info("Starting infrastructure worker...")
	w.Worker.Mu.Lock()
	defer w.Worker.Mu.Unlock()

	if w.Worker.IsRunning {
		return fmt.Errorf("worker is already running")
	}

	if w.Worker.Ctx == nil || w.Worker.Cancel == nil {
		return fmt.Errorf("worker context is nil, worker may have been improperly initialized")
	}

	// Check if context is already cancelled
	select {
	case <-w.Worker.Ctx.Done():
		return fmt.Errorf("worker context is cancelled, cannot start")
	default:
	}

	w.Worker.Logger.Infof("Starting infrastructure worker with schedule: %s", w.Worker.WorkerConfig.CronSchedule)
	w.Worker.Logger.Infof("Worker ID: %s", w.Worker.OwnerID)
	w.Worker.Logger.Infof("RunOnce mode: %v", w.Worker.WorkerConfig.RunOnce)

	// Check if setup is already completed
	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}
	if completed, err := statusManager.IsSetupCompleted(); err != nil {
		w.Worker.Logger.Errorf("Failed to check setup status: %v", err)
	} else if completed {
		if w.Worker.WorkerConfig.ForceRecreate {
			w.Worker.Logger.Info("Infrastructure setup completed but ForceRecreate is enabled")
		} else {
			w.Worker.Logger.Info("Infrastructure setup already completed, starting in monitoring mode")
			return w.startMonitoringMode()
		}
	}

	fmt.Println("hehhhehehhhehehhehh :::::::::::::::::::::::::")

	// Handle RunOnce mode
	if w.Worker.WorkerConfig.RunOnce {
		w.Worker.Logger.Info("Running in RunOnce mode - executing setup once and stopping")
		w.Worker.IsRunning = true

		// Execute setup job once with proper error handling and timeout
		go w.runOnceSetup()

		return nil
	}

	// Add cron job for normal operation
	err := w.Worker.CronJob.AddFunc(w.Worker.WorkerConfig.CronSchedule, w.executeSetupJobWithContext)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Start cron scheduler
	w.Worker.CronJob.Start()
	w.Worker.IsRunning = true

	w.Worker.Logger.Info("Infrastructure worker started successfully")

	// Try immediate execution (best effort) but not in development to avoid conflicts
	if w.Worker.WorkerConfig.Environment != "development" {
		go func() {
			w.Worker.Logger.Info("Attempting immediate infrastructure setup")
			w.executeSetupJobWithContext()
		}()
	}

	return nil

}

// executeSetupJobWithContext is the context-aware cron job function
func (w *Worker) executeSetupJobWithContext() {
	// Use worker's context with timeout
	ctx, cancel := context.WithTimeout(w.Worker.Ctx, 15*time.Minute)
	defer cancel()

	w.executeSetupJobInternal(ctx)
}

// runOnceSetup handles RunOnce mode execution with proper error handling and timeout
func (w *Worker) runOnceSetup() {
	defer func() {
		if r := recover(); r != nil {
			w.Worker.Logger.Errorf("RunOnce setup panicked: %v", r)
		}
		// Automatically stop worker after RunOnce execution
		w.Stop()
	}()

	// Set up timeout context for RunOnce execution
	ctx, cancel := context.WithTimeout(w.Worker.Ctx, 15*time.Minute)
	defer cancel()

	w.Worker.Logger.Info("Executing one-time infrastructure setup")

	// Execute with timeout and proper error handling
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			w.Worker.Logger.Error("RunOnce setup timed out after 15 minutes")
		} else {
			w.Worker.Logger.Error("RunOnce setup cancelled")
		}
		return
	default:
		w.executeSetupJobInternal(ctx)
	}
}

// validateWorkerConfig validates the worker configuration for conflicts and errors
func validateWorkerConfig(config *models.WorkerConfig) error {
	if config == nil {
		return fmt.Errorf("worker config cannot be nil")
	}

	// Validate required fields
	if config.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	if config.LockTimeout <= 0 {
		return fmt.Errorf("lock timeout must be positive")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if config.RetryDelay <= 0 {
		return fmt.Errorf("retry delay must be positive")
	}

	if config.BackoffMultiplier <= 1.0 {
		return fmt.Errorf("backoff multiplier must be greater than 1.0")
	}

	if len(config.RequiredTables) == 0 {
		return fmt.Errorf("at least one required table must be specified")
	}

	if config.LockFilePath == "" {
		return fmt.Errorf("lock file path is required")
	}

	if config.StatusFilePath == "" {
		return fmt.Errorf("status file path is required")
	}

	// Validate cron schedule format
	if config.CronSchedule != "" {
		cronParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		if _, err := cronParser.Parse(config.CronSchedule); err != nil {
			return fmt.Errorf("invalid cron schedule '%s': %w", config.CronSchedule, err)
		}
	}

	// Check for conflicting flags
	if config.RunOnce && config.DryRun {
		// This is actually a valid combination, just log it
	}

	if config.ForceRecreate && config.SkipValidation {
		return fmt.Errorf("ForceRecreate and SkipValidation cannot both be true")
	}

	return nil
}

// getCronScheduleForEnvironment returns environment-specific cron schedules
func getCronScheduleForEnvironment(env string) string {
	switch env {
	case "development":
		return "*/30 * * * * *" // Every 30 seconds for development
	case "testing":
		return "0 */5 * * * *" // Every 5 minutes for testing
	case "production":
		return "0 */15 * * * *" // Every 15 minutes for production
	default:
		return "0 */10 * * * *" // Every 10 minutes default
	}
}

// startMonitoringMode starts the worker in monitoring-only mode
func (w *Worker) startMonitoringMode() error {
	w.Worker.Logger.Info("Starting infrastructure worker in monitoring mode")

	// Schedule periodic health checks
	err := w.Worker.CronJob.AddFunc("0 */10 * * * *", w.healthCheckJob) // Every 10 minutes
	if err != nil {
		return fmt.Errorf("failed to add health check job: %w", err)
	}

	w.Worker.CronJob.Start()
	w.Worker.IsRunning = true

	return nil
}

// healthCheckJob performs periodic health checks
func (w *Worker) healthCheckJob() {
	w.Worker.Logger.Debug("Performing infrastructure health check")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Recreate the worker.InfrastructureSetup from models.InfrastructureSetup
	infrastructureSetup := &InfrastructureSetup{
		InfrastructureSetup: *w.Worker.InfrastructureSetup,
	}

	StatusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}

	tables := infrastructureSetup.getTableDetails()

	if err := infrastructureSetup.validateInfrastructure(ctx, tables); err != nil {
		infrastructureSetup.InfrastructureSetup.Logger.Errorf("Infrastructure health check failed: %v", err)

		// Update status to indicate health issue
		StatusManager.UpdateProgress(models.StatusFailed,
			fmt.Sprintf("Health check failed: %v", err),
			map[string]any{"health_check_failed_at": time.Now()})
	} else {
		infrastructureSetup.InfrastructureSetup.Logger.Debug("Infrastructure health check passed")
	}

}

// GetStatus returns the current worker status
func (w *Worker) GetStatus() (*models.ExecutionResult, error) {
	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}
	return statusManager.LoadStatus()
}

// IsRunning returns whether the worker is currently running
func (w *Worker) IsRunning() bool {
	return w.Worker.IsRunning
}

// ForceSetup forces a setup execution (admin use)
func (w *Worker) ForceSetup() error {
	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}
	if w.Worker.WorkerConfig.ForceRecreate {
		// Reset status to allow re-run
		if err := statusManager.ResetStatus(); err != nil {
			w.Worker.Logger.Errorf("Failed to reset status: %v", err)
		}
	}

	// Trigger immediate execution
	go w.executeSetupJob()

	return nil
}

// executeSetupJob is the main cron job function (legacy compatibility)
func (w *Worker) executeSetupJob() {
	ctx := context.Background()
	w.executeSetupJobInternal(ctx)
}

// executeSetupJobInternal is the core setup execution logic
func (w *Worker) executeSetupJobInternal(ctx context.Context) {
	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}
	// Check if worker is stopping
	select {
	case <-w.Worker.Ctx.Done():
		w.Worker.Logger.Info("Worker is stopping, skipping execution")
		return
	case <-ctx.Done():
		w.Worker.Logger.Info("Context cancelled, skipping execution")
		return
	default:
	}

	// Check if deletion is scheduled - prioritize deletion over setup
	if w.Worker.WorkerConfig.DeletionScheduled && w.Worker.WorkerConfig.DeletionRequested {
		w.Worker.Logger.Info("Infrastructure deletion job triggered")
		w.executeDeletionJob(ctx)
		return
	}

	w.Worker.Logger.Info("Infrastructure setup job triggered")

	// Step 1: Check if already completed
	if completed, err := statusManager.IsSetupCompleted(); err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			w.Worker.Logger.Debug("Status file not found, proceeding with setup")
		} else {
			w.Worker.Logger.Errorf("Failed to check completion status: %v", err)
		}
	} else if completed && !w.Worker.WorkerConfig.ForceRecreate {
		w.Worker.Logger.Info("Infrastructure setup already completed successfully, skipping execution")
		if !w.Worker.WorkerConfig.RunOnce {
			w.Stop()
		}
		return
	}

	// Step 2: Try to acquire lock with context
	lockInfo, err := w.acquireLockWithContext(ctx)
	if err != nil {
		w.Worker.Logger.Warnf("Failed to acquire lock: %v", err)
		return
	}

	// Ensure lock is released
	defer func() {
		lockManager := &LockManager{LockManager: *w.Worker.LockManager}
		if err := lockManager.ReleaseLock(lockInfo); err != nil {
			w.Worker.Logger.Errorf("Failed to release lock: %v", err)
		}
	}()

	w.Worker.Logger.Info("Lock acquired, starting infrastructure setup")

	// Step 3: Execute setup with comprehensive error handling
	if err := w.executeSetupWithErrorHandling(ctx); err != nil {
		w.Worker.Logger.Errorf("Infrastructure setup failed: %v", err)

		// Handle retry logic only if not in RunOnce mode
		if !w.Worker.WorkerConfig.RunOnce {
			if err := w.handleSetupFailure(err); err != nil {
				w.Worker.Logger.Errorf("Failed to handle setup failure: %v", err)
			}
		}
		return
	}

	// Step 4: Setup completed successfully
	w.Worker.Logger.Info("🎉 Infrastructure setup completed successfully! All resources are ready.")

	// Stop the cron job since we're done (except in RunOnce mode where it auto-stops)
	if !w.Worker.WorkerConfig.RunOnce {
		w.Stop()
	}
}

// ScheduleDelete schedules infrastructure deletion
func (w *Worker) ScheduleDelete() error {
	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}
	w.Worker.Mu.Lock()
	defer w.Worker.Mu.Unlock()

	// Check if deletion is already scheduled
	if w.Worker.WorkerConfig.DeletionScheduled {
		return fmt.Errorf("deletion already scheduled")
	}

	// Mark deletion as requested
	w.Worker.WorkerConfig.DeletionRequested = true
	w.Worker.WorkerConfig.DeletionScheduled = true

	// Update status to indicate deletion is scheduled
	if err := statusManager.UpdateProgress(models.StatusDeletionScheduled, "Infrastructure deletion scheduled", map[string]any{
		"scheduled_at":       time.Now(),
		"deletion_requested": true,
	}); err != nil {
		w.Worker.Logger.Errorf("Failed to update status for deletion scheduling: %v", err)
	}

	w.Worker.Logger.Warn("Infrastructure deletion has been scheduled")
	return nil
}

// executeDeletionJob executes the infrastructure deletion
func (w *Worker) executeDeletionJob(ctx context.Context) {
	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}

	w.Worker.Logger.Warn("Starting infrastructure deletion process")

	// Step 1: Try to acquire lock with context
	lockInfo, err := w.acquireLockWithContext(ctx)
	if err != nil {
		w.Worker.Logger.Warnf("Failed to acquire lock for deletion: %v", err)
		return
	}

	// Ensure lock is released
	defer func() {
		lockManager := &LockManager{LockManager: *w.Worker.LockManager}
		if err := lockManager.ReleaseLock(lockInfo); err != nil {
			w.Worker.Logger.Errorf("Failed to release lock after deletion: %v", err)
		}
	}()

	w.Worker.Logger.Warn("Lock acquired, starting infrastructure deletion")

	// Step 2: Update status to deleting
	if err := statusManager.UpdateProgress(models.StatusDeleting, "Deleting infrastructure", map[string]any{
		"deletion_started_at": time.Now(),
	}); err != nil {
		w.Worker.Logger.Errorf("Failed to update status to deleting: %v", err)
	}

	// Step 3: Execute deletion with timeout
	deletionCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	if err := w.executeInfrastructureDeletion(deletionCtx); err != nil {
		w.Worker.Logger.Errorf("Infrastructure deletion failed: %v", err)

		// Mark as failed
		statusManager.UpdateProgress(models.StatusDeletionFailed, fmt.Sprintf("Deletion failed: %v", err), map[string]any{
			"deletion_failed_at": time.Now(),
			"error":              err.Error(),
		})
		return
	}

	// Step 4: Mark as successfully deleted
	if err := statusManager.UpdateProgress(models.StatusDeleted, "Infrastructure successfully deleted", map[string]any{
		"deletion_completed_at": time.Now(),
	}); err != nil {
		w.Worker.Logger.Errorf("Failed to update status to deleted: %v", err)
	}

	// Reset deletion flags
	w.Worker.Mu.Lock()
	w.Worker.WorkerConfig.DeletionScheduled = false
	w.Worker.WorkerConfig.DeletionRequested = false
	w.Worker.Mu.Unlock()

	w.Worker.Logger.Warn("🗑️ Infrastructure deletion completed successfully!")

	// Stop the worker since infrastructure is deleted
	w.Stop()
}

// executeInfrastructureDeletion executes the infrastructure deletion
func (w *Worker) executeInfrastructureDeletion(ctx context.Context) error {
	// Recreate the worker.InfrastructureSetup from models.InfrastructureSetup
	infrastructureSetup := &InfrastructureSetup{
		InfrastructureSetup: *w.Worker.InfrastructureSetup,
	}
	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}
	return infrastructureSetup.ExecuteDelete(ctx, statusManager)
}

// Stop stops the infrastructure worker
func (w *Worker) Stop() error {
	var err error
	w.Worker.StopOnce.Do(func() {
		w.Worker.Mu.Lock()
		defer w.Worker.Mu.Unlock()

		if !w.Worker.IsRunning {
			return
		}

		w.Worker.Logger.Info("Stopping infrastructure worker service")

		// Cancel context first to signal all operations to stop
		if w.Worker.Cancel != nil {
			w.Worker.Cancel()
		}

		// Stop cron job if running
		if w.Worker.CronJob != nil {
			w.Worker.CronJob.Stop()
			// Cron jobs stopped (robfig/cron does not provide a way to wait for jobs to finish)
			w.Worker.Logger.Info("Cron jobs stopped")
		}

		w.Worker.IsRunning = false

		// Signal stop to any waiting goroutines
		select {
		case <-w.Worker.StopChan:
			// Already closed
		default:
			close(w.Worker.StopChan)
		}

		w.Worker.Logger.Info("Infrastructure worker stopped")
	})

	return err
}

// acquireLockWithContext tries to acquire lock with context cancellation support
func (w *Worker) acquireLockWithContext(ctx context.Context) (*models.LockInfo, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Try to acquire lock with timeout
	type result struct {
		lockInfo *models.LockInfo
		err      error
	}

	resultChan := make(chan result, 1)

	go func() {
		lockManager := &LockManager{LockManager: *w.Worker.LockManager}
		lockInfo, err := lockManager.AcquireLock(w.Worker.OwnerID)
		resultChan <- result{lockInfo, err}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("lock acquisition cancelled: %w", ctx.Err())
	case res := <-resultChan:
		return res.lockInfo, res.err
	}
}

// executeSetupWithErrorHandling executes setup with comprehensive error handling
func (w *Worker) executeSetupWithErrorHandling(ctx context.Context) error {
	// Create execution context with timeout
	setupCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	// Initialize execution result
	result := &models.ExecutionResult{
		StartTime:     time.Now(),
		Status:        models.StatusRunning,
		Environment:   w.Worker.Config.AppEnv,
		TablesCreated: make([]models.TableStatus, 0),
		Metadata:      make(map[string]any),
	}

	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}
	// Save initial status
	if err := statusManager.SaveStatus(result); err != nil {
		w.Worker.Logger.Errorf("Failed to save initial status: %v", err)
	}

	// Handle dry run mode
	if w.Worker.WorkerConfig.DryRun {
		w.Worker.Logger.Info("Running in DRY RUN mode - no actual changes will be made")
		result.Success = true
		result.Status = models.StatusCompleted
		result.Metadata["dry_run"] = true
		return statusManager.SaveStatus(result)
	}

	// Execute infrastructure setup
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("infrastructure setup panicked: %v", r)
			w.Worker.Logger.Errorf("Setup panic: %v", err)
			statusManager.MarkFailed(err.Error())
		}
	}()
	// Recreate the worker.InfrastructureSetup from models.InfrastructureSetup
	infrastructureSetup := &InfrastructureSetup{
		InfrastructureSetup: *w.Worker.InfrastructureSetup,
	}
	return infrastructureSetup.Execute(setupCtx, statusManager)
}

// handleSetupFailure handles setup failures with retry logic
func (w *Worker) handleSetupFailure(setupErr error) error {
	statusManager := &StatusManager{StatusManager: *w.Worker.StatusManager}

	// Check if we should retry using metadata
	retryCount, err := statusManager.GetRetryCount()
	if err != nil {
		w.Worker.Logger.Warnf("Failed to get retry count, assuming 0: %v", err)
		retryCount = 0
	}

	if retryCount >= w.Worker.WorkerConfig.MaxRetries {
		w.Worker.Logger.Errorf("Maximum retries (%d) exceeded, giving up", w.Worker.WorkerConfig.MaxRetries)
		return statusManager.MarkFailed(fmt.Sprintf("Max retries exceeded: %v", setupErr))
	}

	// Increment retry count
	if err := statusManager.IncrementRetryCount(); err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	// Calculate next retry delay with exponential backoff
	retryDelay := w.calculateRetryDelay(retryCount)

	w.Worker.Logger.Warnf("Setup failed (attempt %d/%d), will retry in %v: %v",
		retryCount+1, w.Worker.WorkerConfig.MaxRetries+1, retryDelay, setupErr)

	// Update status with retry information
	return statusManager.UpdateProgress(models.StatusRetrying,
		fmt.Sprintf("Retrying after failure: %v", setupErr),
		map[string]any{
			"next_retry_at": time.Now().Add(retryDelay),
			"last_error":    setupErr.Error(),
		})
}

// calculateRetryDelay calculates the delay for the next retry using exponential backoff
func (w *Worker) calculateRetryDelay(retryCount int) time.Duration {
	delay := float64(w.Worker.WorkerConfig.RetryDelay.Nanoseconds())

	// Apply exponential backoff
	for range retryCount {
		delay *= w.Worker.WorkerConfig.BackoffMultiplier
	}

	// Cap at maximum delay (1 hour)
	maxDelay := float64((1 * time.Hour).Nanoseconds())
	if delay > maxDelay {
		delay = maxDelay
	}

	return time.Duration(int64(delay))
}
