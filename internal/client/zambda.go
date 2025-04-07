package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ZambdaSchedule struct {
	Expression  string       `json:"expression" validate:"required" description:"A cron expression that determines when the Zambda Function is invoked. The expression must conform to the format here, https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-cron-expressions.html."`
	Start       string       `json:"start,omitempty" description:"An optional start date and time when the cron expression will go into effect."`
	End         string       `json:"end,omitempty" description:"An optional end date and time when the cron expression will no longer be in effect."`
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty" description:"Optional retry policy that defines the behavior of retry attempts and event age."`
}

type RetryPolicy struct {
	MaximumEventAge int `json:"maximumEventAge,omitempty" description:"The maximum age of a request that Zambda sends to a function for processing, in seconds." validate:"min=60,max=86400"`
	MaximumRetry    int `json:"maximumRetry,omitempty" description:"The maximum number of times to retry when the function returns an error." validate:"min=0,max=185"`
}

type ZambdaFunction struct {
	ID            string          `json:"id" validate:"required,uuid" readOnly:"true"`
	Name          string          `json:"name" validate:"required" description:"A name for the Zambda Function. May contain letters, numbers, dashes, and underscores. Must be unique within the project."`
	Runtime       string          `json:"runtime,omitempty" description:"The runtime of the Zambda Function."`
	Status        string          `json:"status" validate:"required" readOnly:"true" description:"The Zambda Function status provides information about the Functions state including whether it is ready to be invoked."`
	TriggerMethod string          `json:"triggerMethod" validate:"required" description:"The trigger method for the Zambda Function determines how the Function is invoked. Learn more about the different types here, https://docs.oystehr.com/oystehr/services/zambda/#types-of-zambdas."`
	Schedule      *ZambdaSchedule `json:"schedule,omitempty"`
	FileInfo      *FileInfo       `json:"fileInfo,omitempty"`
}

type FileInfo struct {
	Name         string `json:"name,omitempty" description:"The name of the zip file that was uploaded for the Zambda Function."`
	Size         int    `json:"size,omitempty" description:"The size of the zip file that was uploaded for the Zambda Function."`
	LastModified string `json:"lastModified,omitempty" description:"The date and time when the Z3 Object was last modified in ISO 8601 format."`
}

const (
	zambdaBaseURL = "https://zambda-api.zapehr.com/v1/zambda"
)

type zambdaClient struct {
	config ClientConfig
}

func newZambdaClient(config ClientConfig) *zambdaClient {
	return &zambdaClient{config}
}

func (c *zambdaClient) CreateZambda(ctx context.Context, zambda *ZambdaFunction) (*ZambdaFunction, error) {
	url := zambdaBaseURL

	body, err := json.Marshal(zambda)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ZambdaFunction: %w", err)
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

	var createdZambda ZambdaFunction
	if err := json.NewDecoder(resp.Body).Decode(&createdZambda); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdZambda, nil
}

func (c *zambdaClient) GetZambda(ctx context.Context, id string) (*ZambdaFunction, error) {
	url := fmt.Sprintf("%s/%s", zambdaBaseURL, id)

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

	var zambda ZambdaFunction
	if err := json.NewDecoder(resp.Body).Decode(&zambda); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &zambda, nil
}

func (c *zambdaClient) UpdateZambda(ctx context.Context, id string, zambda *ZambdaFunction) (*ZambdaFunction, error) {
	url := fmt.Sprintf("%s/%s", zambdaBaseURL, id)

	body, err := json.Marshal(zambda)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ZambdaFunction: %w", err)
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

	var updatedZambda ZambdaFunction
	if err := json.NewDecoder(resp.Body).Decode(&updatedZambda); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedZambda, nil
}

func (c *zambdaClient) DeleteZambda(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", zambdaBaseURL, id)

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
