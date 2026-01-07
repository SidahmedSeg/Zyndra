package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 3
	cfg.SuccessThreshold = 2
	cfg.Timeout = 100 * time.Millisecond

	cb := NewCircuitBreaker(cfg)

	// Initially closed
	if cb.State() != StateClosed {
		t.Errorf("Initial state = %v, want %v", cb.State(), StateClosed)
	}

	// Fail multiple times to open circuit
	for i := 0; i < 3; i++ {
		err := cb.Call(context.Background(), func() error {
			return errors.New("failure")
		})
		if err == nil {
			t.Error("Expected error from failing function")
		}
	}

	// Circuit should be open
	if cb.State() != StateOpen {
		t.Errorf("State after failures = %v, want %v", cb.State(), StateOpen)
	}

	// Call should fail immediately
	err := cb.Call(context.Background(), func() error {
		return nil
	})
	if err != ErrCircuitOpen {
		t.Errorf("Call() error = %v, want %v", err, ErrCircuitOpen)
	}

	// Wait for timeout to transition to half-open
	// Need to wait a bit longer to ensure state transition
	time.Sleep(200 * time.Millisecond)

	// Try a call to trigger state transition check
	cb.Call(context.Background(), func() error {
		return nil
	})

	// Circuit should be half-open or closed (if success threshold met)
	state := cb.State()
	if state != StateHalfOpen && state != StateClosed {
		t.Errorf("State after timeout = %v, want %v or %v", state, StateHalfOpen, StateClosed)
	}

	// Succeed twice to close circuit
	for i := 0; i < 2; i++ {
		err := cb.Call(context.Background(), func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Call() error = %v, want nil", err)
		}
	}

	// Circuit should be closed
	if cb.State() != StateClosed {
		t.Errorf("State after successes = %v, want %v", cb.State(), StateClosed)
	}
}

func TestCircuitBreaker_FailureInHalfOpen(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 3
	cfg.SuccessThreshold = 2
	cfg.Timeout = 50 * time.Millisecond

	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Call(context.Background(), func() error {
			return errors.New("failure")
		})
	}

	if cb.State() != StateOpen {
		t.Fatalf("Expected circuit to be open")
	}

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Fail in half-open state
	err := cb.Call(context.Background(), func() error {
		return errors.New("failure")
	})

	if err == nil {
		t.Error("Expected error from failing function")
	}

	// Circuit should be open again
	if cb.State() != StateOpen {
		t.Errorf("State after failure in half-open = %v, want %v", cb.State(), StateOpen)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 3

	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Call(context.Background(), func() error {
			return errors.New("failure")
		})
	}

	if cb.State() != StateOpen {
		t.Fatalf("Expected circuit to be open")
	}

	// Reset
	cb.Reset()

	// Circuit should be closed
	if cb.State() != StateClosed {
		t.Errorf("State after reset = %v, want %v", cb.State(), StateClosed)
	}

	stats := cb.GetStats()
	if stats.Failures != 0 {
		t.Errorf("Failures after reset = %d, want 0", stats.Failures)
	}
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 3

	cb := NewCircuitBreaker(cfg)

	stats := cb.GetStats()
	if stats.State != StateClosed {
		t.Errorf("Initial state = %v, want %v", stats.State, StateClosed)
	}

	// Fail once
	cb.Call(context.Background(), func() error {
		return errors.New("failure")
	})

	stats = cb.GetStats()
	if stats.Failures != 1 {
		t.Errorf("Failures = %d, want 1", stats.Failures)
	}

	if stats.State != StateClosed {
		t.Errorf("State = %v, want %v", stats.State, StateClosed)
	}
}

func TestCircuitBreaker_ResetTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 3
	cfg.ResetTimeout = 50 * time.Millisecond

	cb := NewCircuitBreaker(cfg)

	// Fail twice (not enough to open)
	cb.Call(context.Background(), func() error {
		return errors.New("failure")
	})
	cb.Call(context.Background(), func() error {
		return errors.New("failure")
	})

	stats := cb.GetStats()
	if stats.Failures != 2 {
		t.Errorf("Failures = %d, want 2", stats.Failures)
	}

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Make a successful call
	cb.Call(context.Background(), func() error {
		return nil
	})

	// Failures should be reset
	stats = cb.GetStats()
	if stats.Failures != 0 {
		t.Errorf("Failures after reset timeout = %d, want 0", stats.Failures)
	}
}

func TestCircuitBreaker_SuccessResetsFailures(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 5

	cb := NewCircuitBreaker(cfg)

	// Fail twice
	cb.Call(context.Background(), func() error {
		return errors.New("failure")
	})
	cb.Call(context.Background(), func() error {
		return errors.New("failure")
	})

	stats := cb.GetStats()
	if stats.Failures != 2 {
		t.Errorf("Failures = %d, want 2", stats.Failures)
	}

	// Success should reset failures
	cb.Call(context.Background(), func() error {
		return nil
	})

	stats = cb.GetStats()
	if stats.Failures != 0 {
		t.Errorf("Failures after success = %d, want 0", stats.Failures)
	}
}

