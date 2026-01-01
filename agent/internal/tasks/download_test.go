package tasks

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadFile_Success(t *testing.T) {
	content := []byte("test content")
	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.txt")
	
	params := map[string]interface{}{
		"url":    server.URL,
		"dest":   dest,
		"sha256": hashStr,
	}

	err := DownloadFile(params)
	assert.NoError(t, err)

	data, err := os.ReadFile(dest)
	assert.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestDownloadFile_ChecksumMismatch(t *testing.T) {
	content := []byte("test content")
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "mismatch.txt")
	
	params := map[string]interface{}{
		"url":    server.URL,
		"dest":   dest,
		"sha256": "wronghash",
	}

	err := DownloadFile(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")

	_, err = os.Stat(dest)
	assert.True(t, os.IsNotExist(err))
}

func TestDownloadFile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "notfound.txt")
	
	params := map[string]interface{}{
		"url":  server.URL,
		"dest": dest,
	}

	err := DownloadFile(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status: 404")
}
