package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TriggerMethod string

const (
	TriggerMethodAuthenticated   TriggerMethod = "http_auth"
	TriggerMethodUnauthenticated TriggerMethod = "http_open"
	TriggerMethodSubscription    TriggerMethod = "subscription"
	TriggerMethodCron            TriggerMethod = "cron"
)

type Runtime string

const (
	RuntimeNodejs18  Runtime = "nodejs18.x"
	RuntimeNodejs20  Runtime = "nodejs20.x"
	RuntimeNodejs22  Runtime = "nodejs22.x"
	RuntimePython313 Runtime = "python3.13"
	RuntimePython312 Runtime = "python3.12"
	RuntimeJava21    Runtime = "java21"
	RuntimeDotnet9   Runtime = "dotnet9"
	RuntimeRuby33    Runtime = "ruby3.3"
)

var ValidRuntimes = []string{
	string(RuntimeNodejs18),
	string(RuntimeNodejs20),
	string(RuntimeNodejs22),
	string(RuntimePython313),
	string(RuntimePython312),
	string(RuntimeJava21),
	string(RuntimeDotnet9),
	string(RuntimeRuby33),
}

type ZambdaSchedule struct {
	Expression  *string      `json:"expression"`
	Start       *string      `json:"start,omitempty"`
	End         *string      `json:"end,omitempty"`
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
}

type RetryPolicy struct {
	MaximumEventAge *int64 `json:"maximumEventAge,omitempty"`
	MaximumRetry    *int64 `json:"maximumRetry,omitempty"`
}

type ZambdaFunction struct {
	ID               *string         `json:"id"`
	Name             *string         `json:"name"`
	Runtime          *Runtime        `json:"runtime,omitempty"`
	MemorySize       *int32          `json:"memorySize,omitempty"`
	TimeoutInSeconds *int32          `json:"timeoutInSeconds,omitempty"`
	TriggerMethod    *TriggerMethod  `json:"triggerMethod,omitempty"`
	Schedule         *ZambdaSchedule `json:"schedule,omitempty"`
	Status           *string         `json:"status,omitempty"`
	FileInfo         *FileInfo       `json:"fileInfo,omitempty"`
}

type FileInfo struct {
	Name         *string `json:"name,omitempty"`
	Size         *int64  `json:"size,omitempty"`
	LastModified *string `json:"lastModified,omitempty"`
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

	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZambdaFunction: %w", err)
	}

	var createdZambda ZambdaFunction
	if err := json.Unmarshal(responseBody, &createdZambda); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdZambda, nil
}

func (c *zambdaClient) GetZambda(ctx context.Context, id string) (*ZambdaFunction, error) {
	url := fmt.Sprintf("%s/%s", zambdaBaseURL, id)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get ZambdaFunction: %w", err)
	}

	var zambda ZambdaFunction
	if err := json.Unmarshal(responseBody, &zambda); err != nil {
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

	responseBody, err := request(ctx, c.config, http.MethodPatch, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update ZambdaFunction: %w", err)
	}

	var updatedZambda ZambdaFunction
	if err := json.Unmarshal(responseBody, &updatedZambda); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedZambda, nil
}

func (c *zambdaClient) DeleteZambda(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", zambdaBaseURL, id)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete ZambdaFunction: %w", err)
	}

	return nil
}

func (c *zambdaClient) UploadZambdaSource(ctx context.Context, id string, source string) error {
	url := fmt.Sprintf("%s/%s/s3-upload", zambdaBaseURL, id)

	filename := path.Base(source)
	tflog.Info(ctx, fmt.Sprintf("Initiating Zambda source upload for %s to %s", filename, url))

	body, err := json.Marshal(map[string]string{"filename": filename})
	if err != nil {
		return fmt.Errorf("failed to marshal upload request: %w", err)
	}

	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to initiate Zambda source upload: %w", err)
	}

	var uploadInfo struct {
		SignedUrl string `json:"signedUrl"`
	}
	if err := json.Unmarshal(responseBody, &uploadInfo); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	tflog.Info(ctx, fmt.Sprintf("Received signed URL for Zambda source upload: %s", uploadInfo.SignedUrl))

	return uploadToS3(ctx, uploadInfo.SignedUrl, source)
}
