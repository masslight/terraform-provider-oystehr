package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	fhirBaseURL  = "https://fhir-api.zapehr.com"
	maxBatchSize = 100 // Maximum number of entries to process in a single batch
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
	config     *ClientConfig
	entries    []bundleEntry
	entryMutex *sync.Mutex
}

func newFhirClient(config *ClientConfig) *fhirClient {
	c := &fhirClient{config, []bundleEntry{}, &sync.Mutex{}}
	go c.processBundleEntries()
	return c
}

func (c *fhirClient) enqueueBundleEntry(method, url string, body []byte, ifMatch string, channel chan entryResult) {
	c.entryMutex.Lock()
	defer c.entryMutex.Unlock()
	c.entries = append(c.entries, bundleEntry{Method: method, URL: url, Body: body, IfMatch: ifMatch, ResponseChannel: channel})
}

func (c *fhirClient) processBundleEntries() {
	for {
		ctx := tflog.SetField(context.Background(), "oystehr_req_id", uuid.NewString())
		c.entryMutex.Lock()
		var entries []bundleEntry
		if len(c.entries) > 0 {
			if maxBatchSize <= 0 {
				entries = slices.Clone(c.entries)
				c.entries = []bundleEntry{}
			} else if len(c.entries) <= maxBatchSize {
				// If the number of entries is less than or equal to maxBatchSize, process all of them
				entries = slices.Clone(c.entries)
				c.entries = []bundleEntry{}
			} else {
				// If there are more entries than maxBatchSize, process only the first maxBatchSize entries
				entries = slices.Clone(c.entries[:maxBatchSize])
				c.entries = c.entries[maxBatchSize:]
			}
		}
		c.entryMutex.Unlock()

		if len(entries) == 0 {
			tflog.Debug(ctx, "No entries to process")
			continue
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
			sendErrorToAllEntries(ctx, entries, fmt.Errorf("failed to marshal bundle: %w", err))
			continue
		}

		// Send bundle request
		requestCtx, cancel := context.WithTimeout(ctx, time.Second*30)
		responseBody, err := request(requestCtx, c.config, http.MethodPost, fhirBaseURL, bundleJSON)
		cancel()
		if err != nil {
			sendErrorToAllEntries(ctx, entries, fmt.Errorf("failed to send bundle request: %w", err))
			continue
		}

		// Process response
		var bundleResponse map[string]any
		if err := json.Unmarshal(responseBody, &bundleResponse); err != nil {
			sendErrorToAllEntries(ctx, entries, fmt.Errorf("failed to decode bundle response: %w", err))
			continue
		}
		tflog.Debug(ctx, "Decoded bundle response", map[string]any{
			"bundleResponse": bundleResponse,
		})
		if bundleResponse["entry"] == nil {
			sendErrorToAllEntries(ctx, entries, fmt.Errorf("bundle response does not contain 'entry': %s", string(responseBody)))
			continue
		}
		entriesResponse := bundleResponse["entry"].([]any)
		if len(entriesResponse) != len(entries) {
			sendErrorToAllEntries(ctx, entries, fmt.Errorf("bundle response entry count mismatch: expected %d, got %d", len(entries), len(entriesResponse)))
			continue
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
	}
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
	c.enqueueBundleEntry(http.MethodPost, url, jsonData, "", responseChannel)

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
	c.enqueueBundleEntry(http.MethodPut, url, jsonData, fmt.Sprintf(`W/"%s"`, versionID), responseChannel)

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
	c.enqueueBundleEntry(http.MethodGet, url, nil, "", responseChannel)

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
	c.enqueueBundleEntry(http.MethodDelete, url, nil, "", responseChannel)

	response := <-responseChannel
	if response.Error != nil {
		return fmt.Errorf("failed to delete resource: %w", response.Error)
	}

	return nil
}

func sendErrorToAllEntries(ctx context.Context, entries []bundleEntry, err error) {
	tflog.Error(ctx, "Error processing FHIR bundle", map[string]any{
		"error": err,
	})
	for _, entry := range entries {
		select {
		case entry.ResponseChannel <- entryResult{Resource: nil, Error: err}:
		default:
		}
	}
}
