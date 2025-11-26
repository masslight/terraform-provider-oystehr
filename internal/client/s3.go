package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/masslight/terraform-provider-oystehr/internal/fs"
	"github.com/masslight/terraform-provider-oystehr/internal/retry"
)

func uploadToS3(ctx context.Context, url string, source string) error {
	_, err := retry.RetryWithBackoff(ctx, func() (bool, error) {
		var body bytes.Buffer
		data, err := os.ReadFile(fs.CleanPath(source))
		if err != nil {
			return false, fmt.Errorf("failed to read source file: %w", err)
		}
		body.Write(data)
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &body)
		if err != nil {
			return false, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/zip")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, fmt.Errorf("failed to upload source code: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return false, fmt.Errorf("failed to upload source code, status code: %d", resp.StatusCode)
		}

		return true, nil
	}, retry.DefaultRetryConfig)
	return err
}
