package retry

import (
	"context"
	"fmt"
	"time"
)

type Config struct {
	MaxAttempts int
	Delay       time.Duration
	Backoff     time.Duration
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		Delay:       1 * time.Second,
		Backoff:     2 * time.Second,
	}
}

func Do(ctx context.Context, config Config, operation func() error) error {
	var lastErr error
	delay := config.Delay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if attempt > 1 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
			delay += config.Backoff
		}

		lastErr = operation()
		if lastErr == nil {
			return nil
		}

		if attempt < config.MaxAttempts {
			continue
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}
