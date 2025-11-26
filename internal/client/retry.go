package client

import (
	"context"
	"math/rand"
	"time"
)

func retryWithBackoff(ctx context.Context, operation func() error, config struct {
	baseBackoff int64
	maxBackoff  int64
	maxAttempts int
}) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	baseBackoff := time.Duration(config.baseBackoff) * time.Millisecond
	maxBackoff := time.Duration(config.maxBackoff) * time.Millisecond

	for attempt := range config.maxAttempts {
		err := operation()
		if err == nil {
			return nil
		}
		// If this was the last attempt, return the error
		if attempt == config.maxAttempts-1 {
			return err
		}
		// Compute exponential backoff and apply full jitter
		backoff := baseBackoff * (1 << attempt)
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		jitter := time.Duration(rng.Int63n(int64(backoff)))
		time.Sleep(jitter)
	}

	return nil
}
