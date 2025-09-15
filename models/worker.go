package models

import (
	"context"
	"fieldfuze-backend/utils/logger"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/robfig/cron"
)

// DBClient interface to avoid circular dependency
type DBClient interface {
	CreateTable(ctx context.Context, input *dynamodb.CreateTableInput) error
	DescribeTable(ctx context.Context, tableName string) (*dynamodb.DescribeTableOutput, error)
	DeleteTable(ctx context.Context, input *dynamodb.DeleteTableInput) error
}

// StatusManager handles infrastructure setup status tracking
type StatusManager struct {
	StatusFilePath string
}

// LockManager handles distributed locking for infrastructure setup
type LockManager struct {
	LockFilePath string
	LockTimeout  time.Duration
	Environment  string
}

// Worker manages the infrastructure setup cron job
type Worker struct {
	Config              *Config
	Logger              logger.Logger
	CronJob             *cron.Cron
	LockManager         *LockManager
	StatusManager       *StatusManager
	InfrastructureSetup *InfrastructureSetup

	// Worker configuration
	WorkerConfig *WorkerConfig
	OwnerID      string
	IsRunning    bool
	StopChan     chan struct{}

	// Synchronization and state management
	Mu        sync.RWMutex
	Ctx       context.Context
	Cancel    context.CancelFunc
	SetupOnce sync.Once
	StopOnce  sync.Once
}

// InfrastructureSetup handles the actual infrastructure creation
type InfrastructureSetup struct {
	Config   *Config
	Logger   logger.Logger
	DBClient DBClient
}

// WorkerConfig holds configuration for the infrastructure worker
type WorkerConfig struct {
	// Cron schedule
	CronSchedule string `json:"cron_schedule"`

	// Lock settings
	LockTimeout       time.Duration `json:"lock_timeout"`
	LockRetryInterval time.Duration `json:"lock_retry_interval"`

	// Retry settings
	MaxRetries        int           `json:"max_retries"`
	RetryDelay        time.Duration `json:"retry_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`

	// Environment settings
	Environment    string   `json:"environment"`
	RequiredTables []string `json:"required_tables"`

	// Paths
	LockFilePath   string `json:"lock_file_path"`
	StatusFilePath string `json:"status_file_path"`

	// Feature flags
	DryRun         bool `json:"dry_run"`
	SkipValidation bool `json:"skip_validation"`
	ForceRecreate  bool `json:"force_recreate"`
	RunOnce        bool `json:"run_once"`

	// Deletion flags
	DeletionScheduled bool `json:"deletion_scheduled"`
	DeletionRequested bool `json:"deletion_requested"`
}

// LockInfo represents distributed lock information
type LockInfo struct {
	ID          string    `json:"id"`
	Owner       string    `json:"owner"`
	AcquiredAt  time.Time `json:"acquired_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Environment string    `json:"environment"`
}

// WorkerStatus represents the current status of the infrastructure worker
type WorkerStatus string

const (
	StatusIdle              WorkerStatus = "idle"
	StatusRunning           WorkerStatus = "running"
	StatusCompleted         WorkerStatus = "completed"
	StatusFailed            WorkerStatus = "failed"
	StatusRetrying          WorkerStatus = "retrying"
	StatusDeletionScheduled WorkerStatus = "deletion_scheduled"
	StatusDeleting          WorkerStatus = "deleting"
	StatusDeleted           WorkerStatus = "deleted"
	StatusDeletionFailed    WorkerStatus = "deletion_failed"
)

// ExecutionResult holds the result of infrastructure setup execution
type ExecutionResult struct {
	Success        bool                   `json:"success"`
	Status         WorkerStatus           `json:"status"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	Duration       time.Duration          `json:"duration"`
	TablesCreated  []string               `json:"tables_created"`
	IndexesCreated []string               `json:"indexes_created"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	RetryCount     int                    `json:"retry_count"`
	Environment    string                 `json:"environment"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// TableInfo represents table creation information
type TableInfo struct {
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	IndexCount  int               `json:"index_count"`
	BillingMode string            `json:"billing_mode"`
	Tags        map[string]string `json:"tags"`
	Indexes     []string          `json:"indexes,omitempty"`
	ParseName   string            `json:"parse_name"`
}
