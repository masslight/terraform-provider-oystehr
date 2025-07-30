package fs

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"strings"
)

func CleanPath(p string) string {
	ret := p
	if strings.Contains(p, "~") {
		hd, err := os.UserHomeDir()
		if err != nil {
			return fmt.Sprintf("Error getting user home directory: %v. Please provide a path without ~.", err)
		}
		ret = strings.Replace(p, "~", hd, 1)
	}
	return path.Clean(ret)
}

func sha256Hash(data []byte) (string, error) {
	var hasher = sha256.New()
	_, err := hasher.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to write data to hasher: %w", err)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func Sha256HashFile(path string) (string, error) {
	data, err := os.ReadFile(CleanPath(path))
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}
	return sha256Hash(data)
}
