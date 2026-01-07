package infra

import (
	"context"
	"fmt"

	"github.com/intelifox/click-deploy/internal/retry"
)

// RetryClient wraps an infra Client with retry and circuit breaker logic
type RetryClient struct {
	client         Client
	retryConfig    retry.RetryConfig
	circuitBreaker *retry.CircuitBreaker
}

// NewRetryClient creates a new retry-enabled infra client
func NewRetryClient(client Client) *RetryClient {
	return &RetryClient{
		client:      client,
		retryConfig: retry.DefaultRetryConfig(),
		circuitBreaker: retry.NewCircuitBreaker(retry.DefaultConfig()),
	}
}

// WithRetryConfig sets a custom retry configuration
func (c *RetryClient) WithRetryConfig(cfg retry.RetryConfig) *RetryClient {
	c.retryConfig = cfg
	return c
}

// WithCircuitBreakerConfig sets a custom circuit breaker configuration
func (c *RetryClient) WithCircuitBreakerConfig(cfg retry.Config) *RetryClient {
	c.circuitBreaker = retry.NewCircuitBreaker(cfg)
	return c
}

// CreateInstance wraps CreateInstance with retry and circuit breaker
func (c *RetryClient) CreateInstance(ctx context.Context, req CreateInstanceRequest) (*Instance, error) {
	var result *Instance
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			result, err = c.client.CreateInstance(ctx, req)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to create instance: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return nil, fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return result, err
}

// GetInstance wraps GetInstance with retry
func (c *RetryClient) GetInstance(ctx context.Context, instanceID string) (*Instance, error) {
	var result *Instance
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			result, err = c.client.GetInstance(ctx, instanceID)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to get instance: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return nil, fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return result, err
}

// DeleteInstance wraps DeleteInstance with retry
func (c *RetryClient) DeleteInstance(ctx context.Context, instanceID string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.DeleteInstance(ctx, instanceID)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to delete instance: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

// WaitForInstanceStatus wraps WaitForInstanceStatus with retry
func (c *RetryClient) WaitForInstanceStatus(ctx context.Context, instanceID string, status string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.WaitForInstanceStatus(ctx, instanceID, status)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to wait for instance status: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

// AllocateFloatingIP wraps AllocateFloatingIP with retry
func (c *RetryClient) AllocateFloatingIP(ctx context.Context, req AllocateFloatingIPRequest) (*FloatingIP, error) {
	var result *FloatingIP
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			result, err = c.client.AllocateFloatingIP(ctx, req)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to allocate floating IP: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return nil, fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return result, err
}

// AttachFloatingIP wraps AttachFloatingIP with retry
func (c *RetryClient) AttachFloatingIP(ctx context.Context, fipID string, instanceID string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.AttachFloatingIP(ctx, fipID, instanceID)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to attach floating IP: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

// CreateSecurityGroup wraps CreateSecurityGroup with retry
func (c *RetryClient) CreateSecurityGroup(ctx context.Context, req CreateSecurityGroupRequest) (*SecurityGroup, error) {
	var result *SecurityGroup
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			result, err = c.client.CreateSecurityGroup(ctx, req)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to create security group: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return nil, fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return result, err
}

// CreateDNSRecord wraps CreateDNSRecord with retry
func (c *RetryClient) CreateDNSRecord(ctx context.Context, req CreateDNSRecordRequest) (*DNSRecord, error) {
	var result *DNSRecord
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			result, err = c.client.CreateDNSRecord(ctx, req)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to create DNS record: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return nil, fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return result, err
}

// CreateContainer wraps CreateContainer with retry
func (c *RetryClient) CreateContainer(ctx context.Context, req CreateContainerRequest) (*Container, error) {
	var result *Container
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			result, err = c.client.CreateContainer(ctx, req)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to create container: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return nil, fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return result, err
}

// GetContainerStatus wraps GetContainerStatus with retry
func (c *RetryClient) GetContainerStatus(ctx context.Context, containerID string) (*Container, error) {
	var result *Container
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			result, err = c.client.GetContainerStatus(ctx, containerID)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to get container status: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return nil, fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return result, err
}


// StopContainer wraps StopContainer with retry
func (c *RetryClient) StopContainer(ctx context.Context, containerID string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.StopContainer(ctx, containerID)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to stop container: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

// DeleteContainer wraps DeleteContainer with retry
func (c *RetryClient) DeleteContainer(ctx context.Context, containerID string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.DeleteContainer(ctx, containerID)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to delete container: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

// WaitForContainerStatus wraps WaitForContainerStatus with retry
func (c *RetryClient) WaitForContainerStatus(ctx context.Context, containerID string, status string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.WaitForContainerStatus(ctx, containerID, status)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to wait for container status: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

// CreateVolume wraps CreateVolume with retry
func (c *RetryClient) CreateVolume(ctx context.Context, req CreateVolumeRequest) (*Volume, error) {
	var result *Volume
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			result, err = c.client.CreateVolume(ctx, req)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to create volume: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return nil, fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return result, err
}

// AttachVolume wraps AttachVolume with retry
func (c *RetryClient) AttachVolume(ctx context.Context, volumeID string, instanceID string, device string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.AttachVolume(ctx, volumeID, instanceID, device)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to attach volume: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

// DetachVolume wraps DetachVolume with retry
func (c *RetryClient) DetachVolume(ctx context.Context, volumeID string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.DetachVolume(ctx, volumeID)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to detach volume: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

// DeleteVolume wraps DeleteVolume with retry
func (c *RetryClient) DeleteVolume(ctx context.Context, volumeID string) error {
	var err error

	callErr := c.circuitBreaker.Call(ctx, func() error {
		err = retry.Do(ctx, c.retryConfig, func() error {
			err = c.client.DeleteVolume(ctx, volumeID)
			if err != nil {
				return retry.NewRetryableError(fmt.Errorf("failed to delete volume: %w", err))
			}
			return nil
		})
		return err
	})

	if callErr != nil {
		return fmt.Errorf("circuit breaker error: %w", callErr)
	}

	return err
}

