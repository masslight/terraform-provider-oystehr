package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	appBaseURL = "https://app-api.zapehr.com/v1/application"
)

type Application struct {
	ID                     *string  `json:"id,omitempty"`
	Name                   *string  `json:"name"`
	ClientID               *string  `json:"clientId,omitempty"`
	ConnectionName         *string  `json:"connectionName,omitempty"`
	Description            *string  `json:"description,omitempty"`
	LoginRedirectURI       *string  `json:"loginRedirectUri,omitempty"`
	LoginWithEmailEnabled  *bool    `json:"loginWithEmailEnabled,omitempty"`
	AllowedCallbackUrls    []string `json:"allowedCallbackUrls"`
	AllowedLogoutUrls      []string `json:"allowedLogoutUrls"`
	AllowedWebOriginsUrls  []string `json:"allowedWebOriginsUrls"`
	AllowedCORSOriginsUrls []string `json:"allowedCORSOriginsUrls"`
	PasswordlessSMS        *bool    `json:"passwordlessSMS,omitempty"`
	MFAEnabled             *bool    `json:"mfaEnabled,omitempty"`
	ShouldSendInviteEmail  *bool    `json:"shouldSendInviteEmail,omitempty"`
	LogoURI                *string  `json:"logoUri,omitempty"`
	RefreshTokenEnabled    *bool    `json:"refreshTokenEnabled,omitempty"`
}

type applicationClient struct {
	config *ClientConfig
}

func newApplicationClient(config *ClientConfig) *applicationClient {
	return &applicationClient{
		config,
	}
}

func (c *applicationClient) CreateApplication(ctx context.Context, app *Application) (*Application, error) {
	url := appBaseURL
	body, err := json.Marshal(app)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal application: %w", err)
	}

	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		// return nil, fmt.Errorf("failed to create application: %w", err)
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	var createdApp Application
	if err := json.Unmarshal(responseBody, &createdApp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdApp, nil
}

func (c *applicationClient) GetApplication(ctx context.Context, id string) (*Application, error) {
	url := fmt.Sprintf("%s/%s", appBaseURL, id)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	var app Application
	if err := json.Unmarshal(responseBody, &app); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &app, nil
}

func (c *applicationClient) UpdateApplication(ctx context.Context, id string, app *Application) (*Application, error) {
	url := fmt.Sprintf("%s/%s", appBaseURL, id)
	body, err := json.Marshal(app)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal application: %w", err)
	}

	responseBody, err := request(ctx, c.config, http.MethodPatch, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	var updatedApp Application
	if err := json.Unmarshal(responseBody, &updatedApp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedApp, nil
}

func (c *applicationClient) DeleteApplication(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", appBaseURL, id)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	return nil
}
