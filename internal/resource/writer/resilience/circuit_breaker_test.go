package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerClosed(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)

	// Should allow execution when closed
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be closed, got %v", cb.GetState())
	}
}

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker(config)

	// First failure
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be closed after first failure, got %v", cb.GetState())
	}

	// Second failure should open the circuit
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be open after threshold failures, got %v", cb.GetState())
	}
}

func TestCircuitBreakerRejectsWhenOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 1,
		RecoveryTimeout:  1 * time.Second,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker(config)

	// Cause failure to open circuit
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	// Should reject execution when open
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err == nil {
		t.Error("Expected circuit breaker to reject execution when open")
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 1,
		RecoveryTimeout:  10 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker(config)

	// Cause failure to open circuit
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	// Wait for recovery timeout
	time.Sleep(20 * time.Millisecond)

	// Should allow execution after timeout (half-open)
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected success after recovery timeout, got error: %v", err)
	}

	// Should be closed after successful execution
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be closed after successful recovery, got %v", cb.GetState())
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 1,
		RecoveryTimeout:  1 * time.Second,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker(config)

	// Cause failure to open circuit
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be open, got %v", cb.GetState())
	}

	// Reset should close the circuit
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be closed after reset, got %v", cb.GetState())
	}
}
