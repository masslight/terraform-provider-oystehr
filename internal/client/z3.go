package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Bucket struct {
	ID   *string `json:"id"`
	Name *string `json:"name"`
}

type Object struct {
	Bucket       *string `json:"-"`
	Key          *string `json:"key"`
	LastModified *string `json:"lastModified"`
}

const (
	z3BaseURL = "https://z3-api.zapehr.com/v1"
)

type z3Client struct {
	config *ClientConfig
}

func newZ3Client(config *ClientConfig) *z3Client {
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

func (c *z3Client) ListObject(ctx context.Context, bucketName, objectKey string) (*Object, error) {
	url := fmt.Sprintf("%s/%s/%s", z3BaseURL, bucketName, objectKey)

	responseBody, err := request(ctx, c.config, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Object: %w", err)
	}

	var objects []Object
	if err := json.Unmarshal(responseBody, &objects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(objects) == 0 {
		return nil, fmt.Errorf("no objects found in bucket %s matching key %s", bucketName, objectKey)
	}

	// API returns a list of objects, we assume the first one is the desired object
	object := objects[0]

	// API returns bucket and key together as key
	bucket, key, found := strings.Cut(*object.Key, "/")
	if !found {
		return nil, fmt.Errorf("object key %s does not contain a valid bucket prefix", *object.Key)
	}
	object.Bucket = &bucket
	object.Key = &key
	return &object, nil
}

func (c *z3Client) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	url := fmt.Sprintf("%s/%s/%s", z3BaseURL, bucketName, objectKey)

	_, err := request(ctx, c.config, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete Object: %w", err)
	}

	return nil
}

func (c *z3Client) UploadObject(ctx context.Context, bucketName, objectKey, source string) error {
	url := fmt.Sprintf("%s/%s/%s", z3BaseURL, bucketName, objectKey)

	body, err := json.Marshal(map[string]string{"action": "upload"})
	if err != nil {
		return fmt.Errorf("failed to marshal upload request: %w", err)
	}
	responseBody, err := request(ctx, c.config, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to upload Object: %w", err)
	}

	var uploadInfo struct {
		SignedUrl string `json:"signedUrl"`
	}
	if err := json.Unmarshal(responseBody, &uploadInfo); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return uploadToS3(ctx, uploadInfo.SignedUrl, source)
}
