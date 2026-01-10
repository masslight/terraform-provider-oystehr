package retry

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	tt := []struct {
		name          string
		numFailures   int
		baseBackoff   time.Duration
		maxBackoff    time.Duration
		maxDuration   time.Duration
		maxAttempts   int
		disableJitter bool
		expectedRes   int
		expectedErr   error
	}{
		{
			name:        "defaults success",
			numFailures: 1,
			baseBackoff: BaseBackoffDefault,
			maxBackoff:  MaxBackoffDefault,
			maxDuration: MaxDurationDefault,
			maxAttempts: MaxAttemptsDefault,
			expectedRes: 42,
			expectedErr: nil,
		},
		{
			name:        "defaults error",
			numFailures: MaxAttemptsDefault + 1, // 1 more than our attempts
			baseBackoff: BaseBackoffDefault,
			maxBackoff:  MaxBackoffDefault,
			maxDuration: MaxDurationDefault,
			maxAttempts: MaxAttemptsDefault,
			expectedRes: 0,
			expectedErr: fmt.Errorf("some error"),
		},
		{
			name:          "no max duration success",
			numFailures:   5,                 // 1 less than our attempts
			baseBackoff:   MaxBackoffDefault, // always 8 seconds wait
			maxBackoff:    MaxBackoffDefault,
			maxDuration:   Disabled,
			maxAttempts:   6,    // 6 tries, 5 waits; wait 5 times is longer than 30 seconds and has a check > 30 seconds
			disableJitter: true, // wait full time to prove we don't use duration
			expectedRes:   42,
			expectedErr:   nil,
		},
		{
			name:          "no max duration error",
			numFailures:   7,                 // 1 more than our attempts
			baseBackoff:   MaxBackoffDefault, // always 8 seconds wait
			maxBackoff:    MaxBackoffDefault,
			maxDuration:   Disabled,
			maxAttempts:   6,    // 6 tries, 5 waits; wait 5 times is longer than 30 seconds and has a check > 30 seconds
			disableJitter: true, // wait full time to prove we don't use duration
			expectedRes:   0,
			expectedErr:   fmt.Errorf("some error"),
		},
		{
			name:          "no max attempts success",
			numFailures:   3,                 // (3 - 1) * 8 = 16 seconds
			baseBackoff:   MaxBackoffDefault, // always 8 seconds wait
			maxBackoff:    MaxBackoffDefault,
			maxDuration:   20 * time.Second,
			maxAttempts:   Disabled,
			disableJitter: true, // wait full time to prove we don't use attempts
			expectedRes:   42,
			expectedErr:   nil,
		},
		{
			name:          "no max attempts error",
			numFailures:   4,                 // (4 - 1) * 8 = 24 seconds
			baseBackoff:   MaxBackoffDefault, // always 8 seconds wait
			maxBackoff:    MaxBackoffDefault,
			maxDuration:   20 * time.Second,
			maxAttempts:   Disabled,
			disableJitter: true, // wait full time to prove we don't use attempts
			expectedRes:   0,
			expectedErr:   fmt.Errorf("some error"),
		},
	}
	wg := sync.WaitGroup{}
	for _, tc := range tt {
		wg.Add(1)
		go t.Run(tc.name, func(t *testing.T) {
			var i int
			op := func() (int, error) {
				if i > tc.numFailures-1 {
					return 42, nil
				}
				i++
				return 0, fmt.Errorf("some error")
			}
			res, err := RetryWithBackoff(t.Context(), op, RetryConfig{
				BaseBackoff:   tc.baseBackoff,
				MaxBackoff:    tc.maxBackoff,
				MaxDuration:   tc.maxDuration,
				MaxAttempts:   tc.maxAttempts,
				DisableJitter: tc.disableJitter,
			})
			assert.Equal(t, tc.expectedRes, res, "unexpected result")
			assert.Equal(t, tc.expectedErr, err, "unexpected error")
			wg.Done()
		})
	}
	wg.Wait()
}
