package downloader

import "testing"

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
