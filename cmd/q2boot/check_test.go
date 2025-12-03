package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckVirtCat(t *testing.T) {
	// captureOutput is a helper to redirect stdout and capture what's printed.
	captureOutput := func(f func()) string {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		w.Close()
		os.Stdout = oldStdout
		var buf bytes.Buffer
		io.Copy(&buf, r)
		return buf.String()
	}

	t.Run("virt-cat is found", func(t *testing.T) {
		// 1. Create a temporary directory and a fake virt-cat executable.
		tempDir := t.TempDir()
		fakeVirtCatPath := filepath.Join(tempDir, "virt-cat")
		if err := os.WriteFile(fakeVirtCatPath, []byte("#!/bin/sh"), 0755); err != nil {
			t.Fatalf("Failed to create fake virt-cat executable: %v", err)
		}

		// 2. Temporarily modify the PATH to include our temp directory.
		originalPath := os.Getenv("PATH")
		os.Setenv("PATH", tempDir)
		defer os.Setenv("PATH", originalPath)

		// 3. Run the check and capture its output.
		var result bool
		output := captureOutput(func() {
			result = checkVirtCat()
		})

		// 4. Assert the results.
		if !result {
			t.Error("Expected checkVirtCat to return true, but it returned false")
		}
		if !strings.Contains(output, "✅ virt-cat is installed") {
			t.Errorf("Expected output to confirm virt-cat is installed, but got: %s", output)
		}
	})

	t.Run("virt-cat is not found", func(t *testing.T) {
		// 1. Set a PATH that we know does not contain virt-cat.
		originalPath := os.Getenv("PATH")
		os.Setenv("PATH", "") // An empty path is a reliable way to ensure it's not found.
		defer os.Setenv("PATH", originalPath)

		// 2. Run the check.
		result := checkVirtCat()

		// 3. Assert the result.
		if result {
			t.Error("Expected checkVirtCat to return false, but it returned true")
		}
	})
}

func TestCheckKVM(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping KVM test on non-Linux OS")
	}

	// Keep original functions to restore them after the test
	originalReadFile := osReadFile
	originalStat := osStat
	originalOpenFile := osOpenFile
	defer func() {
		osReadFile = originalReadFile
		osStat = originalStat
		osOpenFile = originalOpenFile
	}()

	// Helper to capture stdout
	captureOutput := func(f func()) string {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		f()
		w.Close()
		os.Stdout = oldStdout
		var buf bytes.Buffer
		io.Copy(&buf, r)
		return buf.String()
	}

	t.Run("KVM is fully available", func(t *testing.T) {
		// Mock filesystem dependencies to simulate a perfect setup
		osReadFile = func(name string) ([]byte, error) {
			return []byte("flags: vmx"), nil // Simulate Intel CPU with virtualization
		}
		osStat = func(name string) (os.FileInfo, error) {
			return nil, nil // Simulate file exists
		}
		osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
			return nil, nil // Simulate successful open for R/W
		}

		var result bool
		output := captureOutput(func() {
			result = checkKVM()
		})

		if !result {
			t.Error("Expected checkKVM to return true, but it returned false")
		}
		if !strings.Contains(output, "✅ KVM is available and ready to use") {
			t.Errorf("Expected success message, but got: %s", output)
		}
	})

	t.Run("CPU does not support virtualization", func(t *testing.T) {
		// Mock only cpuinfo to lack virtualization flags
		osReadFile = func(name string) ([]byte, error) {
			return []byte("flags: fpu"), nil
		}

		var result bool
		output := captureOutput(func() {
			result = checkKVM()
		})

		if result {
			t.Error("Expected checkKVM to return false, but it returned true")
		}
		if !strings.Contains(output, "KVM acceleration is not supported") {
			t.Errorf("Expected 'not supported' message, but got: %s", output)
		}
	})

	t.Run("KVM module is not loaded", func(t *testing.T) {
		osReadFile = func(name string) ([]byte, error) {
			return []byte("flags: svm"), nil // AMD CPU is fine
		}
		osStat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist // Simulate /dev/kvm does not exist
		}

		result := checkKVM()
		if result {
			t.Error("Expected checkKVM to return false, but it returned true")
		}
	})

	t.Run("KVM permissions are incorrect", func(t *testing.T) {
		osReadFile = func(name string) ([]byte, error) {
			return []byte("flags: vmx"), nil // Intel CPU is fine
		}
		osStat = func(name string) (os.FileInfo, error) {
			return nil, nil // /dev/kvm exists
		}
		osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
			return nil, os.ErrPermission // Simulate permission denied
		}

		result := checkKVM()
		if result {
			t.Error("Expected checkKVM to return false, but it returned true")
		}
	})
}
