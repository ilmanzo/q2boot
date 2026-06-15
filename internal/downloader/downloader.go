package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VerifySHA256 returns nil iff the file at path hashes to expectedHex
// (case-insensitive). Use this for downloaded images: a fast Content-Length
// check catches truncation, but only a hash catches a tampered/wrong-file
// upstream.
func VerifySHA256(path, expectedHex string) error {
	expectedHex = strings.ToLower(strings.TrimSpace(expectedHex))
	if len(expectedHex) != 64 {
		return fmt.Errorf("invalid SHA-256: expected 64 hex chars, got %d", len(expectedHex))
	}
	if _, err := hex.DecodeString(expectedHex); err != nil {
		return fmt.Errorf("invalid SHA-256: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hash %s: %w", path, err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != expectedHex {
		return fmt.Errorf("SHA-256 mismatch for %s:\n  got      %s\n  expected %s", path, got, expectedHex)
	}
	return nil
}

func IsRemote(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ftp" || u.Scheme == "smb"
}

// Download returns (localPath, cleanup, err). cleanup deletes the temp file
// and must be called by the caller even on success.
func Download(remoteURL string) (string, func(), error) {
	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid URL: %w", err)
	}

	ext := filepath.Ext(u.Path)
	if ext == "" {
		ext = ".qcow2"
	}

	tmpFile, err := os.CreateTemp("", "q2boot-download-*"+ext)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	cleanup := func() { os.Remove(tmpPath) }

	fmt.Printf("Downloading %s to %s...\n", remoteURL, tmpPath)

	switch u.Scheme {
	case "http", "https":
		if err := downloadHTTP(remoteURL, tmpPath); err != nil {
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

	pr := &ProgressReader{Reader: resp.Body, Total: resp.ContentLength}

	// io.Copy surfaces partial-read errors (incl. unexpected EOF on a dropped
	// TCP stream) as a non-nil err, so a silent truncation is not possible
	// here. The post-copy length cross-check guards against a server that
	// advertises Content-Length but closes the stream cleanly mid-body.
	_, err = io.Copy(out, pr)
	fmt.Println()
	if err != nil {
		return err
	}
	if pr.Total > 0 && pr.Current != pr.Total {
		return fmt.Errorf("incomplete download: got %d bytes, expected %d", pr.Current, pr.Total)
	}
	return nil
}

type ProgressReader struct {
	io.Reader
	Total       int64
	Current     int64
	LastPercent int
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)

	if pr.Total > 0 {
		percent := int(float64(pr.Current) / float64(pr.Total) * 100)
		if percent > pr.LastPercent {
			pr.LastPercent = percent
			fmt.Printf("\rDownloading... %d%%", percent)
		}
	} else {
		const mb = 1024 * 1024
		if pr.Current/mb > (pr.Current-int64(n))/mb {
			fmt.Printf("\rDownloading... %d MB", pr.Current/mb)
		}
	}

	return n, err
}

func downloadCurl(url, dest string) error {
	if _, err := exec.LookPath("curl"); err != nil {
		return fmt.Errorf("curl is required for %s downloads but was not found in PATH", url)
	}
	cmd := exec.Command("curl", "-L", "-f", "-o", dest, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
