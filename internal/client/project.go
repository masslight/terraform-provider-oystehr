package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	projectBaseURL = "https://project-api.zapehr.com/v1/project"
)

type Project struct {
	ID                 *string `json:"id,omitempty"`
	Name               *string `json:"name,omitempty"`
	Description        *string `json:"description,omitempty"`
	SignupEnabled      *bool   `json:"signupEnabled,omitempty"`
	DefaultPatientRole struct {
		ID   *string `json:"id,omitempty"`
		Name *string `json:"name,omitempty"`
	} `json:"defaultPatientRole,omitempty"`
	FhirVersion *string `json:"fhirVersion,omitempty"`
	Sandbox     *bool   `json:"sandbox,omitempty"`
}

type ProjectUpdateParams struct {
	Name                 *string `json:"name,omitempty"`
	Description          *string `json:"description,omitempty"`
	SignupEnabled        *bool   `json:"signupEnabled,omitempty"`
	DefaultPatientRoleId *string `json:"defaultPatientRoleId,omitempty"`
}
type projectClient struct {
	config *ClientConfig
}

func newProjectClient(config *ClientConfig) *projectClient {
	return &projectClient{
		config,
	}
}

func (c *projectClient) GetProject(ctx context.Context) (*Project, error) {
	url := projectBaseURL

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	var project Project
	if err := json.Unmarshal(responseBody, &project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

func (c *projectClient) UpdateProject(ctx context.Context, project *ProjectUpdateParams) (*Project, error) {
	url := projectBaseURL
	body, err := json.Marshal(project)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project: %w", err)
	}

	tflog.Info(ctx, "Updating project configuration", map[string]interface{}{
		"body": string(body),
	})

	responseBody, err := request(ctx, c.config, http.MethodPatch, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	var updatedProject Project
	if err := json.Unmarshal(responseBody, &updatedProject); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedProject, nil
}
