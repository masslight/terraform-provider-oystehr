package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Effect string

const (
	EffectAllow Effect = "Allow"
	EffectDeny  Effect = "Deny"
	roleBaseURL        = "https://iam-api.zapehr.com/v1/iam/role"
)

type Rule struct {
	Resource  []string       `json:"resource"`
	Action    []string       `json:"action"`
	Effect    Effect         `json:"effect"`
	Condition map[string]any `json:"condition,omitempty"`
}

type AccessPolicy struct {
	Rule []Rule `json:"rule"`
}

type Role struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	AccessPolicy AccessPolicy `json:"accessPolicy"`
}

type roleClient struct {
	config ClientConfig
}

func newRoleClient(config ClientConfig) *roleClient {
	return &roleClient{config}
}

func (c *roleClient) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	url := roleBaseURL

	body, err := json.Marshal(role)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal role: %w", err)
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

	var createdRole Role
	if err := json.NewDecoder(resp.Body).Decode(&createdRole); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdRole, nil
}

func (c *roleClient) GetRole(ctx context.Context, id string) (*Role, error) {
	url := fmt.Sprintf("%s/%s", roleBaseURL, id)

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

	var role Role
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &role, nil
}

func (c *roleClient) UpdateRole(ctx context.Context, id string, role *Role) (*Role, error) {
	url := fmt.Sprintf("%s/%s", roleBaseURL, id)

	body, err := json.Marshal(role)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal role: %w", err)
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

	var updatedRole Role
	if err := json.NewDecoder(resp.Body).Decode(&updatedRole); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedRole, nil
}

func (c *roleClient) DeleteRole(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", roleBaseURL, id)

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
