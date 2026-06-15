package downloader

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestIsRemote(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"http://example.com/image.qcow2", true},
		{"https://example.com/image.qcow2", true},
		{"ftp://example.com/image.qcow2", true},
		{"smb://example.com/image.qcow2", true},
		{"file:///home/user/image.qcow2", false}, // file scheme treated as local/unsupported for download
		{"/home/user/image.qcow2", false},
		{"relative/path/image.qcow2", false},
		{"image.qcow2", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsRemote(tt.path); got != tt.want {
				t.Errorf("IsRemote(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestVerifySHA256(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "blob")
	if err := os.WriteFile(tmp, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	const helloHash = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

	if err := VerifySHA256(tmp, helloHash); err != nil {
		t.Errorf("correct hash rejected: %v", err)
	}
	if err := VerifySHA256(tmp, "ABCDEF"+helloHash[6:]); err == nil {
		t.Error("wrong hash accepted")
	}
	if err := VerifySHA256(tmp, "tooShort"); err == nil {
		t.Error("short hash accepted")
	}
	if err := VerifySHA256(tmp, "zzzz"+helloHash[4:]); err == nil {
		t.Error("non-hex hash accepted")
	}
	if err := VerifySHA256("/no/such/file", helloHash); err == nil {
		t.Error("missing file accepted")
	}
}

func TestProgressReaderNonTTY(t *testing.T) {
	// Force non-TTY mode; should not crash and should still consume the stream.
	origTTY := IsTTY
	IsTTY = func() bool { return false }
	defer func() { IsTTY = origTTY }()

	data := bytes.Repeat([]byte{0xAB}, 1024)
	pr := &ProgressReader{Reader: bytes.NewReader(data), Total: int64(len(data))}
	n, err := io.Copy(io.Discard, pr)
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(len(data)) {
		t.Errorf("read %d bytes, want %d", n, len(data))
	}
}
