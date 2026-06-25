# 001 — Fix LUKS partition detection (always returns false)

**Status:** DONE (2026-06-25)
**Written against:** working tree on top of commit `33ecdbb` (the code being
fixed is **uncommitted** — it is part of the current `git diff`, not in HEAD).
**Effort:** S · **Risk:** Low · **Confidence:** High (reproduced + fix verified
against real LUKS2 images)

## Symptom (reported by user)

> "encrypt partition detection is not working, I'm not getting the message to
> use --graphical"

Booting a disk with a LUKS-encrypted partition in console mode never prints the
intended early error:

```
disk contains LUKS-encrypted partitions: use --graphical (-g) ...
```

## Root cause

`internal/detector/detector.go`, `HasLUKSPartition` shells out to:

```
virt-filesystems --filesystems -a <disk>
```

and greps the output for `crypto_LUKS`. That string is **never present**:

1. `virt-filesystems --filesystems` lists *mountable* filesystems. A **closed**
   LUKS container is not mountable, so it is excluded entirely — output is empty
   for a LUKS-only disk.
2. Even `--long`, `--all`, and `--extra` do **not** emit `crypto_LUKS` for a
   closed container (verified on guestfs-tools 1.55.8). The plain output also
   carries no type column at all, so the grep could never match regardless.

Therefore `HasLUKSPartition` **always returns false** → the guard in
`cmd/q2boot/main.go` never fires.

### Evidence (reproduced)

Built a real LUKS2 image and probed it:

```
$ virt-filesystems --filesystems -a luks.img          # current code
(empty)
$ virt-filesystems --all --long -a luks.img | grep crypto_LUKS
(no match)
$ guestfish --ro -a luks.img run : vfs-type /dev/sda1  # works
crypto_LUKS
```

`vfs-type` is the only reliable signal. It works for both partition-level LUKS
(`/dev/sda1`) and whole-disk LUKS (`/dev/sda`) — both confirmed.

## The fix

Replace `HasLUKSPartition` in `internal/detector/detector.go`. Enumerate
partitions + whole devices (one appliance boot), then ask `vfs-type` for each
(second boot) and look for `crypto_LUKS`. Switch the tool probe from
`virt-filesystems` to `guestfish` (same `guestfs-tools` package).

### Current code (lines ~41–53)

```go
// HasLUKSPartition returns true if the disk image contains at least one
// LUKS-encrypted partition. Uses virt-filesystems from guestfs-tools;
// returns false (no false positives) when the tool is unavailable.
var HasLUKSPartition = func(diskPath string) bool {
	if _, err := exec.LookPath("virt-filesystems"); err != nil {
		return false
	}
	out, err := exec.Command("virt-filesystems", "--filesystems", "-a", diskPath).Output()
	if err != nil {
		return false
	}
	return bytes.Contains(out, []byte("crypto_LUKS"))
}
```

### Replacement

```go
// HasLUKSPartition reports whether the disk image contains a LUKS-encrypted
// block device. A *closed* LUKS container is not a mountable filesystem, so
// `virt-filesystems --filesystems` (and --all/--extra/--long) never lists it;
// the only reliable signal is libguestfs' vfs-type, which returns
// "crypto_LUKS" for an encrypted device. We enumerate partitions + whole
// devices, then probe each. Returns false when guestfish is missing or the
// probe fails, so a missing guestfs-tools install never blocks a launch.
//
// ponytail: two appliance boots (~3s) on every non-graphical launch when
// guestfish is installed. Acceptable for the safety net; if launch latency
// matters, fold this into the same appliance boot as arch detection.
var HasLUKSPartition = func(diskPath string) bool {
	if _, err := exec.LookPath("guestfish"); err != nil {
		return false
	}
	out, err := exec.Command("guestfish", "--ro", "-a", diskPath,
		"run", ":", "list-partitions", ":", "list-devices").Output()
	if err != nil {
		return false
	}
	args := []string{"--ro", "-a", diskPath, "run"}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if dev := strings.TrimSpace(line); dev != "" {
			args = append(args, ":", "vfs-type", dev)
		}
	}
	if len(args) == 4 { // nothing to probe (empty/unreadable image)
		return false
	}
	types, err := exec.Command("guestfish", args...).Output()
	if err != nil {
		return false
	}
	return bytes.Contains(types, []byte("crypto_LUKS"))
}
```

Imports: `bytes`, `strings`, `os/exec` are all already imported — no import
changes needed. (`vfs-type` on a whole partitioned disk returns empty, not an
error, so including `/dev/sda` in the chain is safe and does not abort it —
verified.)

## Verification

A LUKS image is required to test the positive path. Build one with guestfish
(no root needed; the key is read from the next stdin line):

```sh
guestfish <<'EOF'
sparse /tmp/luks-test.img 64M
run
part-disk /dev/sda mbr
luks-format /dev/sda1 0
secret123
EOF
```

Then:

1. `make test` — all existing tests still pass (they stub `HasLUKSPartition`, so
   they keep passing; this fix is not covered by them — see Test plan).
2. `go vet ./...` — clean.
3. Manual positive check (requires guestfish installed):
   ```sh
   go run ./cmd/q2boot /tmp/luks-test.img --dry-run
   ```
   **Expected:** exits with
   `disk contains LUKS-encrypted partitions: use --graphical (-g) ...`
   (no QEMU command printed).
4. Manual passthrough check:
   ```sh
   go run ./cmd/q2boot /tmp/luks-test.img -g --dry-run
   ```
   **Expected:** LUKS guard skipped (because `-g`); proceeds to arch
   detect / dry-run output.
5. Negative check on a normal image (e.g. one under `diskimages/`):
   ```sh
   go run ./cmd/q2boot diskimages/<non-luks>.qcow2 --dry-run
   ```
   **Expected:** no LUKS error; normal dry-run output.

## Test plan

The existing tests in `internal/detector/detector_test.go`
(`TestHasLUKSPartitionMissing`/`Detected`) only reassign the `HasLUKSPartition`
var to a stub — they verify the indirection, not the real probe, which is why
the bug shipped untested. Leave them as-is (they still pass).

Optional but recommended: add an `e2e`-tagged test mirroring the existing
`cmd/q2boot/e2e_test.go` pattern that, when `guestfish` is on PATH, builds the
LUKS image above and asserts `detector.HasLUKSPartition(path) == true` and that
a non-LUKS image returns `false`. Gate it behind the `e2e` build tag and skip
when `guestfish` is absent (`t.Skip`). Do **not** add it to the default `make
test` run — it needs guestfs-tools and is slow.

## Scope

- **In scope:** `HasLUKSPartition` in `internal/detector/detector.go` only.
- **Out of scope:** the guard call site in `cmd/q2boot/main.go` (already
  correct), arch detection, the rest of the in-flight diff, `check.go` (it
  probes `virt-cat`, also part of guestfs-tools — fine to leave; optionally add
  a `guestfish` mention later).

## Escape hatches

- If `guestfish --ro -a <img> run : list-partitions : list-devices` errors on a
  valid image in your environment (e.g. SELinux/AppArmor denying the appliance),
  STOP and report — do not silently widen permissions.
- If you find a single-command tool that reliably prints `crypto_LUKS` for a
  *closed* container in your guestfs-tools version, prefer it (simpler, one
  boot) and note the version — but verify against the build script above first;
  `virt-filesystems` does not, as of 1.55.8.

## Maintenance note

Watch launch latency: this adds two appliance boots to every non-graphical
launch (when guestfish is present), on top of the arch-detection boot. If that
becomes painful, the fold-into-one-boot optimization noted in the code comment
is the upgrade path.
