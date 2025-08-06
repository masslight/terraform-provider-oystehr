package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

func request(ctx context.Context, config ClientConfig, method, url string, body []byte) ([]byte, error) {
	return requestWithHeaders(ctx, config, method, url, body, nil)
}

func requestWithHeaders(ctx context.Context, config ClientConfig, method, url string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	accessToken, err := getAccessToken(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("x-oystehr-project-id", *config.ProjectID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}
