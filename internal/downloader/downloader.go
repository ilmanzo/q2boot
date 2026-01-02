package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

// IsRemote checks if the path is a remote URL supported for downloading
func IsRemote(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ftp" || u.Scheme == "smb"
}

// Download downloads the file from the URL to a temporary file.
// It returns the path to the temporary file, a cleanup function, and an error.
func Download(remoteURL string) (string, func(), error) {
	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create a temporary file with the same extension as the original file
	ext := filepath.Ext(u.Path)
	if ext == "" {
		ext = ".qcow2" // Default to qcow2 if no extension
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "q2boot-download-*"+ext)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close() // Close immediately, let tools open it

	cleanup := func() {
		os.Remove(tmpPath)
	}

	fmt.Printf("Downloading %s to %s...\n", remoteURL, tmpPath)

	switch u.Scheme {
	case "http", "https":
		// Try internal Go downloader first for HTTP(S)
		if err := downloadHTTP(remoteURL, tmpPath); err != nil {
			// If it fails, maybe try curl? No, net/http is reliable.
			cleanup()
			return "", nil, err
		}
	case "ftp", "smb":
		if err := downloadCurl(remoteURL, tmpPath); err != nil {
			cleanup()
			return "", nil, err
		}
	default:
		cleanup()
		return "", nil, fmt.Errorf("unsupported protocol: %s", u.Scheme)
	}

	fmt.Println("Download complete.")
	return tmpPath, cleanup, nil
}

func downloadHTTP(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Using io.Copy to stream the download
	// TODO: Add progress bar if needed, but for now simple copy
	_, err = io.Copy(out, resp.Body)
	return err
}

func downloadCurl(url, dest string) error {
	// Check if curl exists
	_, err := exec.LookPath("curl")
	if err != nil {
		return fmt.Errorf("curl is required for %s downloads but was not found in PATH", url)
	}

	// Use curl to download
	// -L: Follow redirects
	// -o: Output file
	// -f: Fail silently (no output at all) on server errors
	cmd := exec.Command("curl", "-L", "-f", "-o", dest, url)

	// Connect stdout/stderr to show progress bar from curl
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
