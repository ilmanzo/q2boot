package detector

import "testing"

// Filename detection must run first so a recognizably-named image short-circuits
// the slow virt-cat path. We confirm by passing a non-existent disk path whose
// name still encodes the arch — virt-cat would fail, filename match returns.
func TestDetectArchitecturePrefersFilename(t *testing.T) {
	cases := map[string]string{
		"/no/such/path/openSUSE-Tumbleweed-x86_64.qcow2":      "x86_64",
		"/no/such/path/ubuntu-24.04-aarch64.img":              "aarch64",
		"/no/such/path/fedora_ppc64le.raw":                    "ppc64le",
		"/no/such/path/sles15@s390x.qcow2":                    "s390x",
	}
	for path, want := range cases {
		got, err := DetectArchitecture(path)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", path, err)
		}
		if got != want {
			t.Errorf("%s: got %q, want %q", path, got, want)
		}
	}
}

func TestDetectArchitectureEmptyPath(t *testing.T) {
	if _, err := DetectArchitecture(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}
