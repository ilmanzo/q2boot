//go:build e2e

package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestHasLUKSPartitionReal exercises the real guestfish probe. Skipped when
// guestfish is absent. Slow (~3s per image due to appliance boot).
func TestHasLUKSPartitionReal(t *testing.T) {
	if _, err := exec.LookPath("guestfish"); err != nil {
		t.Skip("guestfish not in PATH")
	}
	dir := t.TempDir()

	runFish := func(script string) {
		t.Helper()
		sf := filepath.Join(dir, "script.fish")
		if err := os.WriteFile(sf, []byte(script), 0600); err != nil {
			t.Fatal(err)
		}
		if out, err := exec.Command("guestfish", "-f", sf).CombinedOutput(); err != nil {
			t.Fatalf("guestfish: %v\n%s", err, out)
		}
	}

	luks := filepath.Join(dir, "luks.img")
	runFish("sparse " + luks + " 64M\nrun\npart-disk /dev/sda mbr\nluks-format /dev/sda1 0\nsecret123\n")
	if !HasLUKSPartition(luks) {
		t.Error("expected true for LUKS image")
	}

	plain := filepath.Join(dir, "plain.img")
	runFish("sparse " + plain + " 32M\nrun\n")
	if HasLUKSPartition(plain) {
		t.Error("expected false for plain image")
	}
}
