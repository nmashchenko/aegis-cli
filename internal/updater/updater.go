package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const repo = "nmashchenko/aegis-cli"

type release struct {
	TagName string `json:"tag_name"`
}

// GetLatestVersion fetches the latest release tag from GitHub.
func GetLatestVersion() (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return "", fmt.Errorf("check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("parse release: %w", err)
	}
	return r.TagName, nil
}

// IsNewer returns true if latest is a higher version than current.
func IsNewer(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	if current == "dev" || current == "" {
		return false
	}
	return latest != current
}

// Update downloads the latest release and replaces the current binary.
func Update(latest string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	url := fmt.Sprintf(
		"https://github.com/%s/releases/download/%s/aegis_%s_%s.tar.gz",
		repo, latest, goos, goarch,
	)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Extract the aegis binary from the tar.gz
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("decompress: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("binary not found in archive")
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}
		if hdr.Name == "aegis" {
			break
		}
	}

	// Get path of current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable path: %w", err)
	}

	// Write to a temp file next to the binary, then rename (atomic on same fs)
	tmpPath := execPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(f, tr); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write binary: %w", err)
	}
	f.Close()

	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace binary: %w", err)
	}

	return nil
}
