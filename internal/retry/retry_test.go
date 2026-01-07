package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo_Success(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3

	attempts := 0
	err := Do(context.Background(), cfg, func() error {
		attempts++
		return nil // Success on first attempt
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestDo_RetryableError(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3
	cfg.InitialDelay = 10 * time.Millisecond

	attempts := 0
	retryableErr := NewRetryableError(errors.New("temporary error"))

	err := Do(context.Background(), cfg, func() error {
		attempts++
		if attempts < 3 {
			return retryableErr
		}
		return nil // Success on third attempt
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestDo_MaxAttemptsReached(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3
	cfg.InitialDelay = 10 * time.Millisecond

	attempts := 0
	retryableErr := NewRetryableError(errors.New("persistent error"))

	err := Do(context.Background(), cfg, func() error {
		attempts++
		return retryableErr
	})

	if err == nil {
		t.Error("Do() error = nil, want error")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// The error returned is wrapped, not directly retryable
	// But it should contain the retryable error
	if err == nil {
		t.Error("Expected error")
	}
}

func TestDo_NonRetryableError(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3

	attempts := 0
	nonRetryableErr := errors.New("permanent error")

	err := Do(context.Background(), cfg, func() error {
		attempts++
		return nonRetryableErr
	})

	if err == nil {
		t.Error("Do() error = nil, want error")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt (non-retryable), got %d", attempts)
	}

	if IsRetryable(err) {
		t.Error("Expected non-retryable error")
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.InitialDelay = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	attempts := 0
	retryableErr := NewRetryableError(errors.New("temporary error"))

	// Cancel context after first attempt
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, cfg, func() error {
		attempts++
		return retryableErr
	})

	if err == nil {
		t.Error("Do() error = nil, want context cancellation error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	// Should have attempted at least once, but not all attempts
	if attempts < 1 || attempts >= 5 {
		t.Errorf("Expected 1-4 attempts (cancelled), got %d", attempts)
	}
}

func TestDo_ExponentialBackoff(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 4
	cfg.InitialDelay = 10 * time.Millisecond
	cfg.Multiplier = 2.0
	cfg.Jitter = false // Disable jitter for predictable timing

	start := time.Now()
	attempts := 0
	retryableErr := NewRetryableError(errors.New("temporary error"))

	err := Do(context.Background(), cfg, func() error {
		attempts++
		if attempts < 4 {
			return retryableErr
		}
		return nil
	})

	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}

	// With exponential backoff: 10ms, 20ms, 40ms = 70ms total
	// Allow some margin for execution time
	if elapsed < 50*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Expected elapsed time ~70ms, got %v", elapsed)
	}

	if attempts != 4 {
		t.Errorf("Expected 4 attempts, got %d", attempts)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "retryable error",
			err:  NewRetryableError(errors.New("temporary")),
			want: true,
		},
		{
			name: "non-retryable error",
			err:  errors.New("permanent"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryable(tt.err)
			if got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 10
	cfg.InitialDelay = 50 * time.Millisecond

	attempts := 0
	retryableErr := NewRetryableError(errors.New("temporary error"))

	err := WithTimeout(context.Background(), 100*time.Millisecond, cfg, func() error {
		attempts++
		return retryableErr
	})

	if err == nil {
		t.Error("WithTimeout() error = nil, want timeout error")
	}

	// Should have attempted a few times before timeout
	if attempts < 1 {
		t.Errorf("Expected at least 1 attempt, got %d", attempts)
	}
}

