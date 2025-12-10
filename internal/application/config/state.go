package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ApplicationState represents the persistent state of the application
type ApplicationState struct {
	LastNotificationId int       `json:"last_notification_id"`
	UpdatedAt          time.Time `json:"updated_at"`
	Version            string    `json:"version"`
}

// StateManager handles persistent application state
type StateManager struct {
	filePath string
	state    *ApplicationState
	mutex    sync.RWMutex
}

const defaultStateFile = "conf/state.json"

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		filePath: defaultStateFile,
		state: &ApplicationState{
			LastNotificationId: 0,
			UpdatedAt:          time.Now(),
			Version:            "1.0",
		},
	}
}

// LoadState loads the state from disk
func (sm *StateManager) LoadState() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(sm.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(sm.filePath); os.IsNotExist(err) {
		// File doesn't exist, use default state
		return sm.saveStateUnsafe()
	}

	// Read and parse the file
	data, err := os.ReadFile(sm.filePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, sm.state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	return nil
}

// SaveState saves the current state to disk
func (sm *StateManager) SaveState() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	return sm.saveStateUnsafe()
}

// saveStateUnsafe saves state without locking (internal use)
func (sm *StateManager) saveStateUnsafe() error {
	sm.state.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file first, then rename (atomic operation)
	tempFile := sm.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	if err := os.Rename(tempFile, sm.filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// GetLastNotificationId returns the last processed notification ID
func (sm *StateManager) GetLastNotificationId() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.state.LastNotificationId
}

// SetLastNotificationId sets the last processed notification ID
func (sm *StateManager) SetLastNotificationId(id int) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.state.LastNotificationId = id
}

// UpdateLastNotificationId updates and saves the last notification ID
func (sm *StateManager) UpdateLastNotificationId(id int) error {
	sm.SetLastNotificationId(id)
	return sm.SaveState()
}
