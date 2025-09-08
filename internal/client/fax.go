package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	faxBaseURL = "https://fax-api.zapehr.com/v1"
)

type FaxNumber struct {
	Number *string `json:"faxNumber,omitempty"`
}

type FaxGetConfigOutput struct {
	Configured *bool    `json:"configured,omitempty"`
	Numbers    []string `json:"faxNumbers,omitempty"`
}

type faxClient struct {
	config *ClientConfig
}

func newFaxClient(config *ClientConfig) *faxClient {
	return &faxClient{config}
}

func (c *faxClient) Onboard(ctx context.Context) (*FaxNumber, error) {
	url := fmt.Sprintf("%s/onboard", faxBaseURL)

	responseBody, err := request(ctx, c.config, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create FaxNumber: %w", err)
	}

	var createdFaxNumber FaxNumber
	if err := json.Unmarshal(responseBody, &createdFaxNumber); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdFaxNumber, nil
}

func (c *faxClient) GetFaxNumber(ctx context.Context, faxNumber string) (*FaxNumber, error) {
	url := fmt.Sprintf("%s/config", faxBaseURL)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get FaxNumber: %w", err)
	}

	var route FaxGetConfigOutput
	if err := json.Unmarshal(responseBody, &route); err != nil {
		return nil, fmt.Errorf("failed to decode FaxNumber response: %w", err)
	}

	if len(route.Numbers) == 0 {
		return nil, fmt.Errorf("FaxNumber not found")
	}
	for _, r := range route.Numbers {
		if r == faxNumber {
			return &FaxNumber{Number: &r}, nil
		}
	}
	return nil, fmt.Errorf("FaxNumber not found")
}

func (c *faxClient) Offboard(ctx context.Context) error {
	url := fmt.Sprintf("%s/offboard", faxBaseURL)

	_, err := request(ctx, c.config, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete LabRoute: %w", err)
	}

	return nil
}
