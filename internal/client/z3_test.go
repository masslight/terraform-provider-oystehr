package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeleteBucketObjectsSuccess(t *testing.T) {
	originalBaseURL := z3BaseURL
	t.Cleanup(func() { z3BaseURL = originalBaseURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/test-bucket/objects" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deletedCount":7}`))
	}))
	t.Cleanup(server.Close)

	z3BaseURL = server.URL + "/v1"
	projectID := "project-1"
	accessToken := "token-1"
	z3Client := newZ3Client(&ClientConfig{ProjectID: &projectID, AccessToken: &accessToken})

	resp, err := z3Client.DeleteBucketObjects(context.Background(), "test-bucket")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response, got nil")
	}
	if resp.DeletedCount != 7 {
		t.Fatalf("expected deletedCount=7, got %d", resp.DeletedCount)
	}
}

func TestDeleteBucketObjectsError(t *testing.T) {
	originalBaseURL := z3BaseURL
	t.Cleanup(func() { z3BaseURL = originalBaseURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	t.Cleanup(server.Close)

	z3BaseURL = server.URL + "/v1"
	projectID := "project-1"
	accessToken := "token-1"
	z3Client := newZ3Client(&ClientConfig{ProjectID: &projectID, AccessToken: &accessToken})

	_, err := z3Client.DeleteBucketObjects(context.Background(), "test-bucket")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "failed to delete Bucket objects") || !strings.Contains(got, "unexpected status code: 403") {
		t.Fatalf("unexpected error: %s", got)
	}
}
