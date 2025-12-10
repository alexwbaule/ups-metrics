package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryWithSuccess(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxAttempts = 3

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := WithRetry(context.Background(), config, fn)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryWithMaxAttemptsExceeded(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxAttempts = 2

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		return errors.New("temporary failure")
	}

	err := WithRetry(context.Background(), config, fn)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryWithNonRetryableError(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxAttempts = 3

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		return errors.New("non-retryable error")
	}

	err := WithRetry(context.Background(), config, fn)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestCalculateDelay(t *testing.T) {
	config := RetryConfig{
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1 * time.Second}, // Capped at MaxDelay
	}

	for _, test := range tests {
		actual := calculateDelay(test.attempt, config)
		if actual != test.expected {
			t.Errorf("Attempt %d: expected %v, got %v", test.attempt, test.expected, actual)
		}
	}
}
