package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	userBaseURL = "https://project-api.zapehr.com/v1/user"
)

type UserInvite struct {
	Email         *string        `json:"email"`
	Username      *string        `json:"username,omitempty"`
	ApplicationID *string        `json:"applicationId"`
	Resource      map[string]any `json:"resource,omitempty"`
	Profile       *string        `json:"profile,omitempty"`
	Roles         []string       `json:"roles,omitempty"`
	AccessPolicy  *AccessPolicy  `json:"accessPolicy,omitempty"`
}

type User struct {
	ID           *string       `json:"id"`
	Name         *string       `json:"name"`
	Email        *string       `json:"email"`
	PhoneNumber  *string       `json:"phoneNumber,omitempty"`
	Profile      *string       `json:"profile,omitempty"`
	AccessPolicy *AccessPolicy `json:"accessPolicy,omitempty"`
	Roles        []string      `json:"roles"`
}

type userOutput struct {
	ID           *string       `json:"id"`
	Name         *string       `json:"name"`
	Email        *string       `json:"email"`
	PhoneNumber  *string       `json:"phoneNumber,omitempty"`
	Profile      *string       `json:"profile,omitempty"`
	AccessPolicy *AccessPolicy `json:"accessPolicy,omitempty"`
	Roles        []RoleStub    `json:"roles"`
}

func userOutputToUser(output *userOutput) *User {
	if output == nil {
		return nil
	}

	roles := make([]string, 0, len(output.Roles))
	for _, role := range output.Roles {
		if role.ID != nil {
			roles = append(roles, *role.ID)
		}
	}

	return &User{
		ID:           output.ID,
		Name:         output.Name,
		Email:        output.Email,
		PhoneNumber:  output.PhoneNumber,
		Profile:      output.Profile,
		AccessPolicy: output.AccessPolicy,
		Roles:        roles,
	}
}

type userClient struct {
	config *ClientConfig
}

func newUserClient(config *ClientConfig) *userClient {
	return &userClient{config}
}

func (c *userClient) InviteUser(ctx context.Context, invite *UserInvite) (*User, error) {
	url := fmt.Sprintf("%s/invite", userBaseURL)

	body, err := json.Marshal(invite)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user invite: %w", err)
	}

	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to invite user: %w", err)
	}

	var output userOutput
	if err := json.Unmarshal(responseBody, &output); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return userOutputToUser(&output), nil
}

func (c *userClient) GetUser(ctx context.Context, id string) (*User, error) {
	url := fmt.Sprintf("%s/%s", userBaseURL, id)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var output userOutput
	if err := json.Unmarshal(responseBody, &output); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return userOutputToUser(&output), nil
}

func (c *userClient) ChangePassword(ctx context.Context, id, password string) error {
	url := fmt.Sprintf("%s/%s/change-password", userBaseURL, id)

	body, err := json.Marshal(map[string]string{"password": password})
	if err != nil {
		return fmt.Errorf("failed to marshal change password request: %w", err)
	}

	if _, err := request(ctx, c.config, http.MethodPost, url, body); err != nil {
		return fmt.Errorf("failed to change user password: %w", err)
	}

	return nil
}

func (c *userClient) DeleteUser(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", userBaseURL, id)

	if _, err := request(ctx, c.config, http.MethodDelete, url, nil); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
