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

type LabRouteAddress struct {
	Address1          *string `json:"address1"`
	Address2          *string `json:"address2,omitempty"`
	City              *string `json:"city"`
	StateProvinceCode *string `json:"stateProvinceCode"`
	PostalCode        *string `json:"postalCode"`
}

type LabRoute struct {
	RouteGUID                 *string          `json:"routeGuid"`
	AccountNumber             *string          `json:"accountNumber"`
	LabGUID                   *string          `json:"labGuid"`
	LabName                   *string          `json:"labName,omitempty"`
	PrimaryID                 *string          `json:"primaryId,omitempty"`
	PrimaryName               *string          `json:"primaryName,omitempty"`
	PrimaryAddress            *LabRouteAddress `json:"primaryAddress,omitempty"`
	ClientSiteID              *string          `json:"clientSiteId,omitempty"`
	EULAVersion               *string          `json:"eulaVersion,omitempty"`
	EULAAccepterFullName      *string          `json:"eulaAccepterFullName,omitempty"`
	EULAAcceptanceDateTimeUTC *string          `json:"eulaAcceptanceDateTimeUtc,omitempty"`
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

	var createdRouteOutput *struct {
		RouteGUID *string `json:"routeGuid"`
	}
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
	url := fmt.Sprintf("%s/route/%s", labBaseURL, routeGUID)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get LabRoute: %w", err)
	}

	var route *LabRoute
	if err := json.Unmarshal(responseBody, &route); err != nil {
		return nil, fmt.Errorf("failed to decode LabRoute response: %w", err)
	}

	if route == nil {
		return nil, fmt.Errorf("LabRoute not found")
	}
	return route, nil
}

func (c *labClient) DeleteLabRoute(ctx context.Context, routeGUID string, labRoute *LabRoute) error {
	url := fmt.Sprintf("%s/route/%s", labBaseURL, routeGUID)

	body, err := json.Marshal(labRoute)
	if err != nil {
		return fmt.Errorf("failed to marshal LabRoute: %w", err)
	}

	_, err = request(ctx, c.config, http.MethodDelete, url, body)
	if err != nil {
		return fmt.Errorf("failed to delete LabRoute: %w", err)
	}

	return nil
}
