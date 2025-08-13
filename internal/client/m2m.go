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

// M2M corresponds to create and update input, and matches the output format for terraform.
type M2M struct {
	ID           *string       `json:"id"`
	ClientID     *string       `json:"clientId"`
	Profile      *string       `json:"profile"`
	JwksURL      *string       `json:"jwksUrl,omitempty"`
	Name         *string       `json:"name"`
	Description  *string       `json:"description,omitempty"`
	AccessPolicy *AccessPolicy `json:"accessPolicy,omitempty"`
	Roles        []string      `json:"roles"`
}

// M2MOutput matches the output format of the API response. It's role field is more complex and less useful than the M2M input type.
type M2MOutput struct {
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

func m2mOutputToM2M(m2mOutput *M2MOutput) *M2M {
	if m2mOutput == nil {
		return nil
	}

	roles := make([]string, len(m2mOutput.Roles))
	for i, role := range m2mOutput.Roles {
		if role.ID != nil {
			roles[i] = *role.ID
		}
	}

	return &M2M{
		ID:           m2mOutput.ID,
		ClientID:     m2mOutput.ClientID,
		Profile:      m2mOutput.Profile,
		JwksURL:      m2mOutput.JwksURL,
		Name:         m2mOutput.Name,
		Description:  m2mOutput.Description,
		AccessPolicy: m2mOutput.AccessPolicy,
		Roles:        roles,
	}
}

type m2mClient struct {
	config *ClientConfig
}

func newM2MClient(config *ClientConfig) *m2mClient {
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

	var createdM2MOutput M2MOutput
	if err := json.Unmarshal(responseBody, &createdM2MOutput); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return m2mOutputToM2M(&createdM2MOutput), nil
}

func (c *m2mClient) GetM2M(ctx context.Context, id string) (*M2M, error) {
	url := fmt.Sprintf("%s/%s", m2mBaseURL, id)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get M2M: %w", err)
	}

	var m2mOutput M2MOutput
	if err := json.Unmarshal(responseBody, &m2mOutput); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return m2mOutputToM2M(&m2mOutput), nil
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

	var updatedM2MOutput M2MOutput
	if err := json.Unmarshal(responseBody, &updatedM2MOutput); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return m2mOutputToM2M(&updatedM2MOutput), nil
}

func (c *m2mClient) DeleteM2M(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", m2mBaseURL, id)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete M2M: %w", err)
	}

	return nil
}
