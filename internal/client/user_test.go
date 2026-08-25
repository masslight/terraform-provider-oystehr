package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestUserClient() *userClient {
	projectID := "project-1"
	accessToken := "token-1"
	return newUserClient(&ClientConfig{ProjectID: &projectID, AccessToken: &accessToken})
}

func TestInviteUserSuccess(t *testing.T) {
	originalBaseURL := userBaseURL
	t.Cleanup(func() { userBaseURL = originalBaseURL })

	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/user/invite" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"user-123","email":"demo@example.com","roles":[{"id":"role-1","name":"admin"}]}`))
	}))
	t.Cleanup(server.Close)

	userBaseURL = server.URL + "/v1/user"
	c := newTestUserClient()

	email := "demo@example.com"
	appID := "app-1"
	user, err := c.InviteUser(context.Background(), &UserInvite{Email: &email, ApplicationID: &appID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected invite endpoint called once, got %d", callCount)
	}
	if user == nil || user.ID == nil || *user.ID != "user-123" {
		t.Fatalf("expected user id user-123, got %+v", user)
	}
	if len(user.Roles) != 1 || user.Roles[0] != "role-1" {
		t.Fatalf("expected roles normalized to [role-1], got %+v", user.Roles)
	}
}

func TestInviteUserError(t *testing.T) {
	originalBaseURL := userBaseURL
	t.Cleanup(func() { userBaseURL = originalBaseURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	t.Cleanup(server.Close)

	userBaseURL = server.URL + "/v1/user"
	c := newTestUserClient()

	email := "demo@example.com"
	appID := "app-1"
	_, err := c.InviteUser(context.Background(), &UserInvite{Email: &email, ApplicationID: &appID})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to invite user") || !strings.Contains(err.Error(), "unexpected status code: 400") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestChangePasswordSuccess(t *testing.T) {
	originalBaseURL := userBaseURL
	t.Cleanup(func() { userBaseURL = originalBaseURL })

	var callCount int
	var receivedPassword string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/user/user-123/change-password" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]string
		_ = json.Unmarshal(body, &payload)
		receivedPassword = payload["password"]
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	userBaseURL = server.URL + "/v1/user"
	c := newTestUserClient()

	if err := c.ChangePassword(context.Background(), "user-123", "s3cret!"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected change-password called once, got %d", callCount)
	}
	if receivedPassword != "s3cret!" {
		t.Fatalf("expected password to be forwarded, got %q", receivedPassword)
	}
}

func TestGetUserNotFound(t *testing.T) {
	originalBaseURL := userBaseURL
	t.Cleanup(func() { userBaseURL = originalBaseURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	t.Cleanup(server.Close)

	userBaseURL = server.URL + "/v1/user"
	c := newTestUserClient()

	_, err := c.GetUser(context.Background(), "missing")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 404") {
		t.Fatalf("expected 404 error, got %s", err.Error())
	}
}

func TestDeleteUserSuccess(t *testing.T) {
	originalBaseURL := userBaseURL
	t.Cleanup(func() { userBaseURL = originalBaseURL })

	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/user/user-123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	userBaseURL = server.URL + "/v1/user"
	c := newTestUserClient()

	if err := c.DeleteUser(context.Background(), "user-123"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected delete called once, got %d", callCount)
	}
}

func TestDeleteUserNotFoundReturns404(t *testing.T) {
	originalBaseURL := userBaseURL
	t.Cleanup(func() { userBaseURL = originalBaseURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	t.Cleanup(server.Close)

	userBaseURL = server.URL + "/v1/user"
	c := newTestUserClient()

	err := c.DeleteUser(context.Background(), "missing")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	// The resource layer relies on this 404 marker to treat delete as success.
	if !strings.Contains(err.Error(), "unexpected status code: 404") {
		t.Fatalf("expected 404 error, got %s", err.Error())
	}
}
