package config

import (
	"context"
	"sync/atomic"
	"time"
)

// PeriodicSaver handles periodic saving of the last notification ID
type PeriodicSaver struct {
	config       *Config
	lastId       int64
	saveInterval time.Duration
	lastSavedId  int64
	stopChan     chan struct{}
}

// NewPeriodicSaver creates a new periodic saver
func NewPeriodicSaver(config *Config, saveInterval time.Duration) *PeriodicSaver {
	if saveInterval == 0 {
		saveInterval = 30 * time.Second // Default to 30 seconds
	}

	return &PeriodicSaver{
		config:       config,
		saveInterval: saveInterval,
		stopChan:     make(chan struct{}),
	}
}

// UpdateLastId updates the last notification ID (in memory)
func (ps *PeriodicSaver) UpdateLastId(id int) {
	atomic.StoreInt64(&ps.lastId, int64(id))
}

// GetLastId returns the current last notification ID
func (ps *PeriodicSaver) GetLastId() int {
	return int(atomic.LoadInt64(&ps.lastId))
}

// Start begins the periodic saving process
func (ps *PeriodicSaver) Start(ctx context.Context) error {
	ticker := time.NewTicker(ps.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Save final state before shutdown
			return ps.saveIfNeeded()
		case <-ps.stopChan:
			return ps.saveIfNeeded()
		case <-ticker.C:
			if err := ps.saveIfNeeded(); err != nil {
				// Log error but continue (don't fail the entire process)
				// The error will be logged by the caller
				continue
			}
		}
	}
}

// Stop stops the periodic saver
func (ps *PeriodicSaver) Stop() error {
	close(ps.stopChan)
	return ps.saveIfNeeded()
}

// saveIfNeeded saves the state only if the ID has changed
func (ps *PeriodicSaver) saveIfNeeded() error {
	currentId := atomic.LoadInt64(&ps.lastId)
	lastSaved := atomic.LoadInt64(&ps.lastSavedId)

	if currentId != lastSaved {
		if err := ps.config.UpdateLastNotificationId(int(currentId)); err != nil {
			return err
		}
		atomic.StoreInt64(&ps.lastSavedId, currentId)
	}

	return nil
}
