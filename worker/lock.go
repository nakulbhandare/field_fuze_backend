package worker

import (
	"encoding/json"
	"fieldfuze-backend/models"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LockManager handles distributed locking for infrastructure setup
type LockManager struct {
	models.LockManager
}

// NewLockManager creates a new lock manager
func NewLockManager(lockPath string, timeout time.Duration, env string) *models.LockManager {
	return &models.LockManager{
		LockFilePath: lockPath,
		LockTimeout:  timeout,
		Environment:  env,
	}
}

func (lm *LockManager) AcquireLock(ownerID string) (*models.LockInfo, error) {

	if err := os.MkdirAll(filepath.Dir(lm.LockFilePath), 0755); err != nil {
		return nil, err
	}
	if existingLock, err := lm.readLockFile(); err == nil {
		if time.Now().Before(existingLock.ExpiresAt) {
			if existingLock.Owner == ownerID && existingLock.Environment == lm.Environment {
				// Lock is already held by this owner, extend it
				return lm.extendLock(existingLock, ownerID)
			}

		}
	}

	// Create new lock
	lockInfo := &models.LockInfo{
		ID:          fmt.Sprintf("infra-lock-%d", time.Now().UnixNano()),
		Owner:       ownerID,
		AcquiredAt:  time.Now(),
		ExpiresAt:   time.Now().Add(lm.LockTimeout),
		Environment: lm.Environment,
	}

	if err := lm.writeLockFile(lockInfo); err != nil {
		return nil, fmt.Errorf("failed to create initial lock file: %w", err)
	}
	return lockInfo, nil
}

func (lm *LockManager) readLockFile() (*models.LockInfo, error) {
	data, err := os.ReadFile(lm.LockFilePath)
	if err != nil {
		return nil, err
	}

	var lockInfo models.LockInfo
	if err := json.Unmarshal(data, &lockInfo); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	return &lockInfo, nil
}

func (lm *LockManager) extendLock(existingLock *models.LockInfo, ownerID string) (*models.LockInfo, error) {
	if existingLock.Owner != ownerID {
		return nil, fmt.Errorf("cannot extend lock owned by %s", existingLock.Owner)
	}

	extendedLock := &models.LockInfo{
		ID:          existingLock.ID,
		Owner:       existingLock.Owner,
		AcquiredAt:  existingLock.AcquiredAt,
		ExpiresAt:   time.Now().Add(lm.LockTimeout),
		Environment: existingLock.Environment,
	}

	if err := lm.writeLockFile(extendedLock); err != nil {
		return nil, fmt.Errorf("failed to extend lock: %w", err)
	}
	return extendedLock, nil
}

func (lm *LockManager) writeLockFile(lockInfo *models.LockInfo) error {
	data, err := json.MarshalIndent(lockInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize lock info: %w", err)
	}
	tempFile := lm.LockFilePath + ".tmp"

	fmt.Printf("Debug: lockFilePath=%s, tempFile=%s\n", lm.LockFilePath, tempFile) // Debug line

	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp lock file: %w", err)
	}
	if err := os.Rename(tempFile, lm.LockFilePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename temp lock file: %w", err)
	}
	return nil
}

// CleanupExpiredLocks removes expired lock files
func (lm *LockManager) CleanupExpiredLocks() error {
	lockInfo, err := lm.readLockFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No lock file to clean up
		}
		return err
	}

	if time.Now().After(lockInfo.ExpiresAt) {
		return os.Remove(lm.LockFilePath)
	}

	return nil
}

// ReleaseLock releases the distributed lock
func (lm *LockManager) ReleaseLock(lockInfo *models.LockInfo) error {
	// Verify we own the lock
	currentLock, err := lm.readLockFile()
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist, nothing to release
			return nil
		}
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	if currentLock.Owner != lockInfo.Owner {
		return fmt.Errorf("cannot release lock owned by %s", currentLock.Owner)
	}

	// Remove lock file
	if err := os.Remove(lm.LockFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	return nil
}
