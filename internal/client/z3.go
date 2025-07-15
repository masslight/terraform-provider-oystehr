package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Bucket struct {
	ID   *string `json:"id"`
	Name *string `json:"name"`
}

const (
	z3BaseURL = "https://z3-api.zapehr.com/v1"
)

type z3Client struct {
	config ClientConfig
}

func newZ3Client(config ClientConfig) *z3Client {
	return &z3Client{config}
}

func (c *z3Client) CreateBucket(ctx context.Context, bucket *Bucket) (*Bucket, error) {
	url := fmt.Sprintf("%s/%s", z3BaseURL, *bucket.Name)

	responseBody, err := request(ctx, c.config, http.MethodPut, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bucket: %w", err)
	}

	var createdBucket Bucket
	if err := json.Unmarshal(responseBody, &createdBucket); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdBucket, nil
}

func (c *z3Client) GetBucket(ctx context.Context, bucketName string) (*Bucket, error) {
	url := z3BaseURL

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Bucket: %w", err)
	}

	var buckets []Bucket
	if err := json.Unmarshal(responseBody, &buckets); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, bucket := range buckets {
		if *bucket.Name == bucketName {
			return &bucket, nil
		}
	}

	return nil, fmt.Errorf("bucket not found: %s", bucketName)
}

func (c *z3Client) DeleteBucket(ctx context.Context, bucketName string) error {
	url := fmt.Sprintf("%s/%s", z3BaseURL, bucketName)

	responseBody, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete Bucket: %w", err)
	}

	if len(responseBody) > 0 {
		return fmt.Errorf("unexpected response body when deleting Bucket: %s", responseBody)
	}

	return nil
}
