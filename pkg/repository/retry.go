package repository

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
	Logger          *logrus.Logger
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     20,
		InitialDelay:    1 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"no such host", "connection refused", "timeout"},
		Logger:          logrus.StandardLogger(),
	}
}

func RetryWithExponentialBackoff(operation func() error, config RetryConfig) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		errStr := err.Error()

		retryable := false
		for _, retryableErr := range config.RetryableErrors {
			if contains(errStr, retryableErr) {
				retryable = true
				break
			}
		}

		if !retryable {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		if attempt == config.MaxAttempts {
			break
		}

		config.Logger.WithFields(logrus.Fields{
			"attempt": attempt,
			"delay":   delay.String(),
			"error":   err.Error(),
		}).Warning("Operation failed, retrying...")

		time.Sleep(delay)

		delay = time.Duration(float64(delay) * config.BackoffFactor)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:(len(s)-len(substr)+1)] != substr
}
