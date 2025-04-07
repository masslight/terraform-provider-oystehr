package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	m2mBaseURL = "https://iam-api.zapehr.com/v1/m2m"
)

type M2M struct {
	ID           string       `json:"id"`
	ClientID     string       `json:"clientId"`
	Profile      string       `json:"profile"`
	JwksURL      string       `json:"jwksUrl,omitempty"`
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	AccessPolicy AccessPolicy `json:"accessPolicy,omitempty"`
	Roles        []RoleStub   `json:"roles"`
}

type RoleStub struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type m2mClient struct {
	config ClientConfig
}

func newM2MClient(config ClientConfig) *m2mClient {
	return &m2mClient{config}
}

func (c *m2mClient) CreateM2M(ctx context.Context, m2m *M2M) (*M2M, error) {
	url := m2mBaseURL

	body, err := json.Marshal(m2m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal M2M: %w", err)
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

	var createdM2M M2M
	if err := json.NewDecoder(resp.Body).Decode(&createdM2M); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdM2M, nil
}

func (c *m2mClient) GetM2M(ctx context.Context, id string) (*M2M, error) {
	url := fmt.Sprintf("%s/%s", m2mBaseURL, id)

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

	var m2m M2M
	if err := json.NewDecoder(resp.Body).Decode(&m2m); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &m2m, nil
}

func (c *m2mClient) UpdateM2M(ctx context.Context, id string, m2m *M2M) (*M2M, error) {
	url := fmt.Sprintf("%s/%s", m2mBaseURL, id)

	body, err := json.Marshal(m2m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal M2M: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var updatedM2M M2M
	if err := json.NewDecoder(resp.Body).Decode(&updatedM2M); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedM2M, nil
}

func (c *m2mClient) DeleteM2M(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", m2mBaseURL, id)

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
