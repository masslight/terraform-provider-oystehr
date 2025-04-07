package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	appBaseURL = "https://app-api.zapehr.com/v1"
)

type Application struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	ClientID               string   `json:"clientId"`
	Description            string   `json:"description,omitempty"`
	LoginRedirectURI       string   `json:"loginRedirectUri,omitempty"`
	LoginWithEmailEnabled  bool     `json:"loginWithEmailEnabled,omitempty"`
	AllowedCallbackUrls    []string `json:"allowedCallbackUrls,omitempty"`
	AllowedLogoutUrls      []string `json:"allowedLogoutUrls,omitempty"`
	AllowedWebOriginsUrls  []string `json:"allowedWebOriginsUrls,omitempty"`
	AllowedCORSOriginsUrls []string `json:"allowedCORSOriginsUrls,omitempty"`
	PasswordlessSMS        bool     `json:"passwordlessSMS,omitempty"`
	MFAEnabled             bool     `json:"mfaEnabled,omitempty"`
	ShouldSendInviteEmail  bool     `json:"shouldSendInviteEmail,omitempty"`
	LogoURI                string   `json:"logoUri,omitempty"`
	RefreshTokenEnabled    bool     `json:"refreshTokenEnabled,omitempty"`
}

type applicationClient struct {
	config ClientConfig
}

func newApplicationClient(config ClientConfig) *applicationClient {
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

	var createdApp Application
	if err := json.NewDecoder(resp.Body).Decode(&createdApp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdApp, nil
}

func (c *applicationClient) GetApplication(ctx context.Context, id string) (*Application, error) {
	url := fmt.Sprintf("%s/%s", appBaseURL, id)

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

	var app Application
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
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

	var updatedApp Application
	if err := json.NewDecoder(resp.Body).Decode(&updatedApp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedApp, nil
}

func (c *applicationClient) DeleteApplication(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", appBaseURL, id)

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
