package retry

import (
	"context"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	BaseBackoffDefault time.Duration = 500
	MaxBackoffDefault  time.Duration = 8000
	MaxAttemptsDefault               = 3
)

var (
	// DefaultRetryConfig defines the default configuration for retrying operations.
	DefaultRetryConfig = RetryConfig{
		BaseBackoff: BaseBackoffDefault,
		MaxBackoff:  MaxBackoffDefault,
		MaxAttempts: MaxAttemptsDefault,
	}
)

type RetryConfig struct {
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	MaxAttempts int
}

func RetryWithBackoff[T any](ctx context.Context, operation func() (T, error), config RetryConfig) (T, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	baseBackoff := time.Duration(config.BaseBackoff) * time.Millisecond
	maxBackoff := time.Duration(config.MaxBackoff) * time.Millisecond

	for attempt := range config.MaxAttempts {
		res, err := operation()
		if err == nil {
			return res, nil
		}
		// If this was the last attempt, return the error
		if attempt == config.MaxAttempts-1 {
			var zero T
			return zero, err
		}
		// Compute exponential backoff and apply full jitter
		backoff := baseBackoff * (1 << attempt)
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		jitter := time.Duration(rng.Int63n(int64(backoff)))
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
