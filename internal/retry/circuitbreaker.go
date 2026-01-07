package retry

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrCircuitHalfOpen is returned when the circuit breaker is half-open
	ErrCircuitHalfOpen = errors.New("circuit breaker is half-open")
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// Config configures circuit breaker behavior
type Config struct {
	FailureThreshold   int           // Number of failures before opening circuit
	SuccessThreshold   int           // Number of successes in half-open to close circuit
	Timeout            time.Duration // Time to wait before attempting half-open
	ResetTimeout       time.Duration // Time to wait before resetting failure count
	MaxConcurrentCalls int           // Maximum concurrent calls (optional, 0 = unlimited)
}

// DefaultConfig returns a default circuit breaker configuration
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		ResetTimeout:     60 * time.Second,
		MaxConcurrentCalls: 0, // Unlimited
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config      Config
	state       CircuitBreakerState
	failures    int
	successes   int
	lastFailure time.Time
	lastReset   time.Time
	mu          sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		config:    config,
		state:     StateClosed,
		lastReset: time.Now(),
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Call executes a function through the circuit breaker
func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
	// Check if we can make the call
	if err := cb.beforeCall(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Update circuit breaker state based on result
	cb.afterCall(err)

	return err
}

// beforeCall checks if a call can be made
func (cb *CircuitBreaker) beforeCall() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// Reset failure count if reset timeout has passed
	if now.Sub(cb.lastReset) > cb.config.ResetTimeout {
		cb.failures = 0
		cb.lastReset = now
	}

	switch cb.state {
	case StateClosed:
		// Allow call
		return nil

	case StateOpen:
		// Check if timeout has passed, transition to half-open
		if now.Sub(cb.lastFailure) > cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Allow call (testing if service is back)
		return nil

	default:
		return ErrCircuitOpen
	}
}

// afterCall updates circuit breaker state based on call result
func (cb *CircuitBreaker) afterCall(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Call failed
		cb.failures++
		cb.lastFailure = time.Now()

		switch cb.state {
		case StateClosed:
			// Check if we should open the circuit
			if cb.failures >= cb.config.FailureThreshold {
				cb.state = StateOpen
				cb.lastFailure = time.Now()
			}

		case StateHalfOpen:
			// Failure in half-open state, go back to open
			cb.state = StateOpen
			cb.lastFailure = time.Now()
			cb.successes = 0
		}
	} else {
		// Call succeeded
		cb.failures = 0

		switch cb.state {
		case StateClosed:
			// Reset failure count on success
			cb.lastReset = time.Now()

		case StateHalfOpen:
			// Count successes in half-open state
			cb.successes++
			if cb.successes >= cb.config.SuccessThreshold {
				// Enough successes, close the circuit
				cb.state = StateClosed
				cb.successes = 0
				cb.lastReset = time.Now()
			}
		}
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.lastReset = time.Now()
}

// Stats returns circuit breaker statistics
type Stats struct {
	State     CircuitBreakerState
	Failures  int
	Successes int
	LastFailure time.Time
}

// GetStats returns current circuit breaker statistics
func (cb *CircuitBreaker) GetStats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return Stats{
		State:       cb.state,
		Failures:    cb.failures,
		Successes:   cb.successes,
		LastFailure: cb.lastFailure,
	}
}

