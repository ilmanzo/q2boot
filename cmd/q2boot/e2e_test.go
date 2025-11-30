//go:build e2e

package e2e_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestE2E is the main end-to-end test for q2boot. It builds the binary,
// then runs a subtest for each supported architecture.
func TestE2E(t *testing.T) {
	tempDir := t.TempDir()
	q2bootPath := filepath.Join(tempDir, "q2boot_test_binary")

	// Build the q2boot binary
	buildCmd := exec.Command("go", "build", "-o", q2bootPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build q2boot binary: %v", err)
	}

	// Define architectures and their disk images
	diskImages := map[string]string{
		"x86_64":  "../../diskimages/sle-16.0-x86_64-135.5-textmode@64bit.qcow2",
		"aarch64": "../../diskimages/sle-16.0-aarch64-135.5-textmode@aarch64.qcow2",
		"ppc64le": "../../diskimages/sle-16.0-ppc64le-135.5-textmode@ppc64le-p10-virtio.qcow2",
		"s390x":   "../../diskimages/sle-16.0-s390x-135.5-textmode@s390x-kvm.qcow2",
	}

	// Convert to absolute paths
	for arch, path := range diskImages {
		absPath, err := filepath.Abs(path)
		if err != nil {
			t.Fatalf("Failed to get absolute path for disk image %s: %v", path, err)
		}
		diskImages[arch] = absPath
	}

	// Run tests for each architecture
	for arch, diskImage := range diskImages {
		// Capture arch and diskImage in local variables for the closure
		arch := arch
		diskImage := diskImage
		t.Run(arch, func(t *testing.T) {
			t.Parallel() // Run architecture tests in parallel
			if _, err := os.Stat(diskImage); os.IsNotExist(err) {
				t.Skipf("Disk image for %s not found at %s, skipping test", arch, diskImage)
			}
			runQ2BootAndCheck(t, q2bootPath, arch, diskImage)
		})
	}
}

// runQ2BootAndCheck starts q2boot, waits for a login prompt in the logs,
// and then shuts down the VM via the QEMU monitor.
func runQ2BootAndCheck(t *testing.T, q2bootPath string, arch string, diskImage string) {
	// Overall timeout for the entire test run for this architecture
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Use a temporary config directory to avoid races
	configDir := t.TempDir()
	testConfigDir = configDir // This sets the global var in the main package

	// Find a free port for the monitor
	monitorPort, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to find a free monitor port: %v", err)
	}

	// Find a free port for SSH forwarding
	sshPort, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to find a free SSH port: %v", err)
	}

	// Prepare the q2boot command
	cmd := exec.CommandContext(ctx, q2bootPath,
		"-d", diskImage,
		"-a", arch,
		"--monitor-port", fmt.Sprintf("%d", monitorPort),
		"--ssh-port", fmt.Sprintf("%d", sshPort),
		"--confirm=false", // Ensure we don't wait for user input
	)

	// Capture stdout to check for login prompt
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	// Capture stderr for better error reporting
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Start the command and create a channel to receive its exit error
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start q2boot command: %v", err)
	}
	t.Logf("q2boot process started with PID %d", cmd.Process.Pid)

	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	// Channel to signal when the login prompt is found
	loginFound := make(chan struct{})

	// Goroutine to scan stdout for the login prompt
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			t.Logf("[stdout] %s", line)
			if strings.Contains(line, "login:") {
				t.Log("Login prompt found in stdout")
				close(loginFound) // Signal that we found it
				return
			}
		}
	}()

	bootTimeout := time.After(60 * time.Second)

	// Wait for boot, command exit, or timeout
	select {
	case <-loginFound:
		// Boot successful, now quit via monitor
		t.Logf("Attempting to quit VM via monitor on port %d", monitorPort)
		if err := quitViaMonitor(monitorPort); err != nil {
			t.Errorf("Failed to send quit command via monitor: %v", err)
			// If we can't quit gracefully, we have to be more forceful
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}

	case err := <-cmdDone:
		// Command exited before login prompt was found
		t.Fatalf("q2boot exited prematurely with error: %v. Stderr:\n%s", err, stderr.String())

	case <-bootTimeout:
		// Timed out waiting for login
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("Timed out waiting for login prompt after 60s. Stderr:\n%s", stderr.String())
	}

	// Wait for the command to exit after sending quit
	select {
	case err := <-cmdDone:
		if err != nil {
			// QEMU often exits with a non-zero status when quitting via monitor,
			// which is expected. We just log it.
			t.Logf("q2boot exited with: %v. This is often expected after quitting via monitor.", err)
		} else {
			t.Log("q2boot exited gracefully.")
		}
	case <-time.After(10 * time.Second):
		t.Error("q2boot did not exit within 10 seconds after quit command")
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
}

// quitViaMonitor connects to the QEMU monitor and sends the quit command.
func quitViaMonitor(port int) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 5*time.Second)
	if err != nil {
		return fmt.Errorf("could not connect to monitor: %w", err)
	}
	defer conn.Close()

	// QEMU monitor might send a welcome message, read it to clear buffer
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reader := bufio.NewReader(conn)
	_, _ = reader.ReadString('\n')

	// Send the quit command
	_, err = conn.Write([]byte("quit\n"))
	if err != nil {
		return fmt.Errorf("could not send quit command: %w", err)
	}

	return nil
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
