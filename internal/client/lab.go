package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	labBaseURL = "https://labs-api.zapehr.com/v1"
)

type LabRoute struct {
	RouteGUID     *string `json:"routeGuid"`
	LabGUID       *string `json:"labGuid"`
	AccountNumber *string `json:"accountNumber"`
}

type CreateLabRouteOutput struct {
	RouteGUID *string `json:"routeGuid"`
}

type labClient struct {
	config *ClientConfig
}

func newLabClient(config *ClientConfig) *labClient {
	return &labClient{config}
}

func (c *labClient) CreateLabRoute(ctx context.Context, route *LabRoute) (*LabRoute, error) {
	url := fmt.Sprintf("%s/route", labBaseURL)

	body, err := json.Marshal(route)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LabRoute: %w", err)
	}

	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create LabRoute: %w; %+v, %s", err, route, body)
	}

	var createdRouteOutput CreateLabRouteOutput
	if err := json.Unmarshal(responseBody, &createdRouteOutput); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	createdRoute, err := c.GetLabRoute(ctx, *createdRouteOutput.RouteGUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created LabRoute: %w", err)
	}

	return createdRoute, nil
}

func (c *labClient) GetLabRoute(ctx context.Context, routeGUID string) (*LabRoute, error) {
	url := fmt.Sprintf("%s/route", labBaseURL)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get LabRoute: %w", err)
	}

	var route []LabRoute
	if err := json.Unmarshal(responseBody, &route); err != nil {
		return nil, fmt.Errorf("failed to decode LabRoute response: %w", err)
	}

	if len(route) == 0 {
		return nil, fmt.Errorf("LabRoute not found")
	}
	for _, r := range route {
		if r.RouteGUID != nil && *r.RouteGUID == routeGUID {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("LabRoute not found")
}

func (c *labClient) DeleteLabRoute(ctx context.Context, routeGUID string) error {
	url := fmt.Sprintf("%s/route/%s", labBaseURL, routeGUID)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete LabRoute: %w", err)
	}

	return nil
}
