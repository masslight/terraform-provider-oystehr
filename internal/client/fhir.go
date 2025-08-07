package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	fhirBaseURL = "https://fhir-api.zapehr.com"
)

type entryResult struct {
	Resource any
	Error    error
}

type bundleEntry struct {
	Method          string
	URL             string
	IfMatch         string
	Body            []byte
	ResponseChannel chan entryResult
}

type fhirClient struct {
	config       ClientConfig
	entries      []bundleEntry
	entryMutex   *sync.Mutex
	processMutex *sync.Mutex
}

func newFhirClient(config ClientConfig) *fhirClient {
	return &fhirClient{config, []bundleEntry{}, &sync.Mutex{}, &sync.Mutex{}}
}

func (c *fhirClient) enqueueBundleEntry(ctx context.Context, method, url string, body []byte, ifMatch string, channel chan entryResult) {
	c.entryMutex.Lock()
	defer c.entryMutex.Unlock()
	c.entries = append(c.entries, bundleEntry{Method: method, URL: url, Body: body, IfMatch: ifMatch, ResponseChannel: channel})
	go c.processBundleEntries(ctx)
}

func (c *fhirClient) processBundleEntries(ctx context.Context) error {
	c.processMutex.Lock()
	defer c.processMutex.Unlock()
	c.entryMutex.Lock()
	entries := slices.Clone(c.entries)
	c.entries = []bundleEntry{}
	c.entryMutex.Unlock()

	if len(entries) == 0 {
		tflog.Debug(ctx, "No entries to process")
		return nil
	}

	tflog.Debug(ctx, "Assembling FHIR bundle", map[string]any{
		"entry_count": len(entries),
	})
	// Assemble bundle from entries
	bundle := map[string]any{
		"resourceType": "Bundle",
		"type":         "batch",
		"entry":        make([]map[string]any, len(entries)),
	}
	for i, entry := range entries {
		bundle["entry"].([]map[string]any)[i] = map[string]any{
			"request": map[string]any{
				"method": entry.Method,
				"url":    entry.URL,
			},
		}
		if entry.Body != nil {
			bundle["entry"].([]map[string]any)[i]["resource"] = json.RawMessage(entry.Body)
		}
		if entry.IfMatch != "" {
			bundle["entry"].([]map[string]any)[i]["request"].(map[string]any)["ifMatch"] = entry.IfMatch
		}
	}
	tflog.Debug(ctx, "Assembled FHIR bundle", map[string]any{
		"bundle": bundle,
	})

	// Marshal bundle to JSON
	bundleJSON, err := json.Marshal(bundle)
	if err != nil {
		sendErrorToAllEntries(entries, fmt.Errorf("failed to marshal bundle: %w", err))
		return fmt.Errorf("failed to marshal bundle: %w", err)
	}

	// Send bundle request
	responseBody, err := request(ctx, c.config, http.MethodPost, fhirBaseURL, bundleJSON)
	if err != nil {
		sendErrorToAllEntries(entries, fmt.Errorf("failed to send bundle request: %w", err))
		return fmt.Errorf("failed to send bundle request: %w", err)
	}

	// Process response
	var bundleResponse map[string]any
	if err := json.Unmarshal(responseBody, &bundleResponse); err != nil {
		sendErrorToAllEntries(entries, fmt.Errorf("failed to decode bundle response: %w", err))
		return fmt.Errorf("failed to decode bundle response: %w", err)
	}
	tflog.Debug(ctx, "Decoded bundle response", map[string]any{
		"bundleResponse": bundleResponse,
	})
	if bundleResponse["entry"] == nil {
		sendErrorToAllEntries(entries, fmt.Errorf("bundle response does not contain 'entry': %s", string(responseBody)))
		return fmt.Errorf("bundle response does not contain 'entry': %s", string(responseBody))
	}
	entriesResponse := bundleResponse["entry"].([]any)
	if len(entriesResponse) != len(entries) {
		sendErrorToAllEntries(entries, fmt.Errorf("bundle response entry count mismatch: expected %d, got %d", len(entries), len(entriesResponse)))
		return fmt.Errorf("bundle response entry count mismatch: expected %d, got %d", len(entries), len(entriesResponse))
	}

	// Fan out responses
	for i, entry := range entries {
		resource := entriesResponse[i].(map[string]any)["resource"]
		response := entriesResponse[i].(map[string]any)["response"]
		status := response.(map[string]any)["status"]
		statusStr := status.(string)
		statusCode, _ := strconv.Atoi(statusStr)
		if statusCode < 200 || statusCode >= 300 {
			select {
			case entry.ResponseChannel <- entryResult{Resource: resource, Error: fmt.Errorf("unexpected status code: %d, response body: %+v", statusCode, response)}:
			default:
			}
		} else {
			select {
			case entry.ResponseChannel <- entryResult{Resource: resource, Error: nil}:
			default:
			}
		}
	}
	return nil
}

func (c *fhirClient) CreateResource(ctx context.Context, resourceType string, data map[string]any) (map[string]interface{}, error) {
	url := fmt.Sprintf("/%s", resourceType)

	if data["resourceType"] == nil {
		data["resourceType"] = resourceType
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	responseChannel := make(chan entryResult, 1)
	defer close(responseChannel)
	c.enqueueBundleEntry(ctx, http.MethodPost, url, jsonData, "", responseChannel)

	response := <-responseChannel
	if response.Error != nil {
		return nil, fmt.Errorf("failed to create resource: %w", response.Error)
	}

	result, ok := response.Resource.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to decode response")
	}

	return result, nil
}

func (c *fhirClient) UpdateResource(ctx context.Context, resourceType, resourceID string, versionID string, data map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("/%s/%s", resourceType, resourceID)

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

	responseChannel := make(chan entryResult, 1)
	defer close(responseChannel)
	c.enqueueBundleEntry(ctx, http.MethodPut, url, jsonData, fmt.Sprintf(`W/"%s"`, versionID), responseChannel)

	response := <-responseChannel
	if response.Error != nil {
		return nil, fmt.Errorf("failed to update resource: %w", response.Error)
	}

	result, ok := response.Resource.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to decode response")
	}

	return result, nil
}

func (c *fhirClient) GetResource(ctx context.Context, resourceType, resourceID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("/%s/%s", resourceType, resourceID)

	responseChannel := make(chan entryResult, 1)
	defer close(responseChannel)
	c.enqueueBundleEntry(ctx, http.MethodGet, url, nil, "", responseChannel)

	response := <-responseChannel
	if response.Error != nil {
		return nil, fmt.Errorf("failed to get resource: %w", response.Error)
	}

	result, ok := response.Resource.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to decode response")
	}

	return result, nil
}

func (c *fhirClient) DeleteResource(ctx context.Context, resourceType, resourceID string) error {
	url := fmt.Sprintf("/%s/%s", resourceType, resourceID)

	responseChannel := make(chan entryResult, 1)
	defer close(responseChannel)
	c.enqueueBundleEntry(ctx, http.MethodDelete, url, nil, "", responseChannel)

	response := <-responseChannel
	if response.Error != nil {
		return fmt.Errorf("failed to delete resource: %w", response.Error)
	}

	return nil
}

func sendErrorToAllEntries(entries []bundleEntry, err error) {
	for _, entry := range entries {
		select {
		case entry.ResponseChannel <- entryResult{Resource: nil, Error: err}:
		default:
		}
	}
}
