package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	secretBaseURL = "https://zambda-api.zapehr.com/v1/secret"
)

type Secret struct {
	Name  *string `json:"name"`
	Value *string `json:"value"`
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

	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	var createdSecret Secret
	if err := json.Unmarshal(responseBody, &createdSecret); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdSecret, nil
}

func (c *secretClient) GetSecret(ctx context.Context, name string) (*Secret, error) {
	url := fmt.Sprintf("%s/%s", secretBaseURL, name)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	var secret Secret
	if err := json.Unmarshal(responseBody, &secret); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &secret, nil
}

func (c *secretClient) DeleteSecret(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/%s", secretBaseURL, name)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}
