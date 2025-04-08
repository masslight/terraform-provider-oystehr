package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func getAccessToken(ctx context.Context, config ClientConfig) (string, error) {
	if config.AccessToken != nil {
		return *config.AccessToken, nil
	}

	if config.ClientID == nil || config.ClientSecret == nil {
		return "", fmt.Errorf("client ID or client secret is not set")
	}

	url := "https://auth.zapehr.com/oauth/token"
	data := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     *config.ClientID,
		"client_secret": *config.ClientSecret,
		"audience":      "https://api.zapehr.com",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("received non-2xx response: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	config.AccessToken = &result.AccessToken
	return result.AccessToken, nil
}
