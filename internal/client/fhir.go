package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type fhirClient struct {
	config ClientConfig
}

func newFhirClient(config ClientConfig) *fhirClient {
	return &fhirClient{config}
}

func (c *fhirClient) CreateResource(ctx context.Context, resourceType string, data map[string]any) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://fhir-api.zapehr.com/%s", resourceType)

	if data["resourceType"] == nil {
		data["resourceType"] = resourceType
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	responseBody, err := request(ctx, c.config, http.MethodPost, url, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func (c *fhirClient) UpdateResource(ctx context.Context, resourceType, resourceID string, data map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://fhir-api.zapehr.com/%s/%s", resourceType, resourceID)

	if data["resourceType"] == nil {
		data["resourceType"] = resourceType
	}
	if data["id"] == nil {
		data["id"] = resourceID
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	responseBody, err := request(ctx, c.config, http.MethodPut, url, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func (c *fhirClient) GetResource(ctx context.Context, resourceType, resourceID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://fhir-api.zapehr.com/%s/%s", resourceType, resourceID)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func (c *fhirClient) DeleteResource(ctx context.Context, resourceType, resourceID string) error {
	url := fmt.Sprintf("https://fhir-api.zapehr.com/%s/%s", resourceType, resourceID)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}
