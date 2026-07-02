package provider

import (
	"context"
	"fmt"
	"testing"
)

func TestExtractZ3StatusCode(t *testing.T) {
	tt := []struct {
		name  string
		err   error
		code  int
		ok    bool
		is500 bool
		is404 bool
		is403 bool
		is400 bool
	}{
		{
			name:  "extracts 500",
			err:   fmt.Errorf("failed request: unexpected status code: 500, response body: boom"),
			code:  500,
			ok:    true,
			is500: true,
		},
		{
			name:  "extracts 404",
			err:   fmt.Errorf("unexpected status code: 404, response body: not found"),
			code:  404,
			ok:    true,
			is404: true,
		},
		{
			name:  "extracts 403",
			err:   fmt.Errorf("unexpected status code: 403, response body: forbidden"),
			code:  403,
			ok:    true,
			is403: true,
		},
		{
			name:  "extracts 400",
			err:   fmt.Errorf("unexpected status code: 400, response body: non-empty"),
			code:  400,
			ok:    true,
			is400: true,
		},
		{
			name: "no status code marker",
			err:  fmt.Errorf("some other error"),
			ok:   false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			code, ok := extractZ3StatusCode(tc.err)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if ok && code != tc.code {
				t.Fatalf("expected code=%d, got %d", tc.code, code)
			}
			if got := isZ3StatusCodeError(tc.err, 500); got != tc.is500 {
				t.Fatalf("expected is500=%v, got %v", tc.is500, got)
			}
			if got := isZ3StatusCodeError(tc.err, 404); got != tc.is404 {
				t.Fatalf("expected is404=%v, got %v", tc.is404, got)
			}
			if got := isZ3StatusCodeError(tc.err, 403); got != tc.is403 {
				t.Fatalf("expected is403=%v, got %v", tc.is403, got)
			}
			if got := isZ3StatusCodeError(tc.err, 400); got != tc.is400 {
				t.Fatalf("expected is400=%v, got %v", tc.is400, got)
			}
		})
	}
}

func TestRetryZ3OnServerErrorsRetries5xx(t *testing.T) {
	attempts := 0
	err := retryZ3OnServerErrors(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("unexpected status code: 500, response body: transient")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryZ3OnServerErrorsNoRetryFor4xx(t *testing.T) {
	attempts := 0
	err := retryZ3OnServerErrors(context.Background(), func() error {
		attempts++
		return fmt.Errorf("unexpected status code: 403, response body: forbidden")
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt for non-retriable error, got %d", attempts)
	}
}

func TestRetryZ3OnServerErrorsStopsAfterMaxAttempts(t *testing.T) {
	attempts := 0
	err := retryZ3OnServerErrors(context.Background(), func() error {
		attempts++
		return fmt.Errorf("unexpected status code: 503, response body: unavailable")
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
