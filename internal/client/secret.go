package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	secretBaseURL = "https://zambda-api.zapehr.com/v1/secret"
)

type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type secretClient struct {
	config ClientConfig
}

func newSecretClient(config ClientConfig) *secretClient {
	return &secretClient{config}
}

func (c *secretClient) SetSecret(ctx context.Context, secret *Secret) (*Secret, error) {
	url := secretBaseURL

	body, err := json.Marshal(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal secret: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var createdSecret Secret
	if err := json.NewDecoder(resp.Body).Decode(&createdSecret); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdSecret, nil
}

func (c *secretClient) GetSecret(ctx context.Context, name string) (*Secret, error) {
	url := fmt.Sprintf("%s/%s", secretBaseURL, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var secret Secret
	if err := json.NewDecoder(resp.Body).Decode(&secret); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &secret, nil
}

func (c *secretClient) DeleteSecret(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/%s", secretBaseURL, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
