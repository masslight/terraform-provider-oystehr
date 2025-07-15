package fs

import (
	"crypto/sha256"
	"fmt"
	"os"
)

func sha256Hash(data []byte) (string, error) {
	var hasher = sha256.New()
	_, err := hasher.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to write data to hasher: %w", err)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func Sha256HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}
	return sha256Hash(data)
}
