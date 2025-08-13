package client

import (
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
	Resource  []string       `json:"resource" tfsdk:"resource"`
	Action    []string       `json:"action" tfsdk:"action"`
	Effect    *Effect        `json:"effect" tfsdk:"effect"`
	Condition map[string]any `json:"condition,omitempty" tfsdk:"condition"`
}

type AccessPolicy struct {
	Rule []Rule `json:"rule" tfsdk:"rule"`
}

type Role struct {
	ID           *string       `json:"id"`
	Name         *string       `json:"name"`
	Description  *string       `json:"description,omitempty"`
	AccessPolicy *AccessPolicy `json:"accessPolicy"`
}

type roleClient struct {
	config *ClientConfig
}

func newRoleClient(config *ClientConfig) *roleClient {
	return &roleClient{config}
}

func (c *roleClient) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	url := roleBaseURL
	body, err := json.Marshal(role)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal role: %w", err)
	}

	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	var createdRole Role
	if err := json.Unmarshal(responseBody, &createdRole); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdRole, nil
}

func (c *roleClient) GetRole(ctx context.Context, id string) (*Role, error) {
	url := fmt.Sprintf("%s/%s", roleBaseURL, id)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	var role Role
	if err := json.Unmarshal(responseBody, &role); err != nil {
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

	responseBody, err := request(ctx, c.config, http.MethodPatch, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	var updatedRole Role
	if err := json.Unmarshal(responseBody, &updatedRole); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedRole, nil
}

func (c *roleClient) DeleteRole(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", roleBaseURL, id)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}
