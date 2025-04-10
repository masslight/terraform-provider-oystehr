package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	m2mBaseURL = "https://iam-api.zapehr.com/v1/m2m"
)

type M2M struct {
	ID           *string       `json:"id"`
	ClientID     *string       `json:"clientId"`
	Profile      *string       `json:"profile"`
	JwksURL      *string       `json:"jwksUrl,omitempty"`
	Name         *string       `json:"name"`
	Description  *string       `json:"description,omitempty"`
	AccessPolicy *AccessPolicy `json:"accessPolicy,omitempty"`
	Roles        []RoleStub    `json:"roles"`
}

type RoleStub struct {
	ID   *string `json:"id"`
	Name *string `json:"name"`
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

	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create M2M: %w; %+v, %s", err, m2m, body)
	}

	var createdM2M M2M
	if err := json.Unmarshal(responseBody, &createdM2M); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdM2M, nil
}

func (c *m2mClient) GetM2M(ctx context.Context, id string) (*M2M, error) {
	url := fmt.Sprintf("%s/%s", m2mBaseURL, id)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get M2M: %w", err)
	}

	var m2m M2M
	if err := json.Unmarshal(responseBody, &m2m); err != nil {
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

	responseBody, err := request(ctx, c.config, http.MethodPatch, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update M2M: %w", err)
	}

	var updatedM2M M2M
	if err := json.Unmarshal(responseBody, &updatedM2M); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedM2M, nil
}

func (c *m2mClient) DeleteM2M(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", m2mBaseURL, id)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete M2M: %w", err)
	}

	return nil
}
