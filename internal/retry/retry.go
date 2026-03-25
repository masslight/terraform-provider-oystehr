package retry

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	BaseBackoffDefault time.Duration = 500 * time.Millisecond
	MaxBackoffDefault  time.Duration = 8000 * time.Millisecond
	MaxDurationDefault time.Duration = 30 * time.Second
	MaxAttemptsDefault               = 3
	Disabled                         = 0
)

var (
	// DefaultRetryConfig defines the default configuration for retrying operations.
	DefaultRetryConfig = RetryConfig{
		BaseBackoff:   BaseBackoffDefault,
		MaxBackoff:    MaxBackoffDefault,
		MaxDuration:   MaxDurationDefault,
		MaxAttempts:   MaxAttemptsDefault,
		DisableJitter: false,
	}
)

type RetryConfig struct {
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	// Set to retry.Disable to disable max duration
	MaxDuration time.Duration
	// Set to retry.Disable to disable max attempts
	MaxAttempts int
	// Set to true to wait the full backoff
	DisableJitter bool
}

func RetryWithBackoff[T any](ctx context.Context, operation func() (T, error), config RetryConfig) (T, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	start := time.Now()

	for attempt := range int(math.Max(float64(config.MaxAttempts), 1000)) {
		res, err := operation()
		if err == nil {
			return res, nil
		}
		// If this was the last attempt, return the error
		if config.MaxAttempts > Disabled && attempt == config.MaxAttempts-1 {
			var zero T
			return zero, err
		}
		// If we have elapsed max duration, return the error
		if config.MaxDuration > Disabled && time.Since(start) > config.MaxDuration {
			var zero T
			return zero, err
		}
		// Compute exponential backoff and apply full jitter
		backoff, jitter := calculateBackoffAndJitter(rng, config, attempt)
		tflog.Debug(ctx, "Retrying operation after backoff", map[string]any{
			"attempt":      attempt + 1,
			"max_attempts": config.MaxAttempts,
			"backoff":      backoff,
			"jitter":       jitter,
		})
		time.Sleep(jitter)
	}

	var zero T
	return zero, nil
}

func calculateBackoffAndJitter(rng *rand.Rand, config RetryConfig, attempt int) (backoff time.Duration, jitter time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			// panic, jitter too big
			backoff = config.MaxBackoff
			jitter = config.MaxBackoff
		}
	}()
	backoff = min(config.BaseBackoff*(1<<attempt), config.MaxBackoff)
	if config.DisableJitter {
		return backoff, backoff
	}
	return backoff, time.Duration(rng.Int63n(int64(backoff)))
}
