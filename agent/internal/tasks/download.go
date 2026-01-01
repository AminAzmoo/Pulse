package tasks

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

func DownloadFile(params map[string]interface{}) error {
	url, ok := params["url"].(string)
	if !ok || url == "" {
		return errors.New("missing or invalid 'url' parameter")
	}

	dest, ok := params["dest"].(string)
	if !ok || dest == "" {
		return errors.New("missing or invalid 'dest' parameter")
	}

	// Security: Prevent directory traversal
	if filepath.Clean(dest) != dest || filepath.IsAbs(dest) == false {
		// Enforce absolute paths and clean paths
		// Depending on policy, we might want to restrict to specific directories
	}
	// Ideally we restrict to specific allowed directories, but for now we enforce absolute path
	if !filepath.IsAbs(dest) {
		return errors.New("destination path must be absolute")
	}

	expectedHash, _ := params["sha256"].(string)

	zap.L().Info("Starting file download", zap.String("url", url), zap.String("dest", dest))

	return doDownload(url, dest, expectedHash)
}

func doDownload(url, dest, expectedHash string) error {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create temp file first
	tmpDest := dest + ".tmp"
	out, err := os.Create(tmpDest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()
	defer os.Remove(tmpDest) // Cleanup on error

	// Calculate hash while downloading
	hasher := sha256.New()
	writer := io.MultiWriter(out, hasher)

	if _, err = io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Verify hash
	if expectedHash != "" {
		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if actualHash != expectedHash {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
		}
	}

	out.Close() // Close before rename
	if err := os.Rename(tmpDest, dest); err != nil {
		return fmt.Errorf("failed to move file to destination: %w", err)
	}

	zap.L().Info("File download completed successfully", zap.String("dest", dest))
	return nil
}
