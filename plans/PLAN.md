# Backlog — q2boot improvement findings

Audit date: 2026-06-25 · against working tree on top of `33ecdbb` (in-flight
uncommitted changes present).

LUKS detection (the urgent bug) has its own full plan:
[001-fix-luks-detection.md](001-fix-luks-detection.md). Everything else lives
here as a ranked backlog — enough detail to act on later, not full executor
specs. Promote any item to its own `00N-*.md` when you pick it up.

| # | Finding | Cat | Effort | Risk | Status |
|---|---------|-----|--------|------|--------|
| 001 | LUKS detection always false | bug | S | Low | DONE (2026-06-25) |
| A | e2e tests silently skip (stale image names) | tests | S | Low | DONE (2026-06-25) |
| B | Default log file litters cwd | DX/debt | S | Low | DONE (2026-06-25) |
| C | `http.Get` has no timeout | security | S | Low | DONE (2026-06-25) |
| D | UEFI vars file sized from CODE blob | bug-latent | S | Low | DONE (2026-06-25) |
| E | LUKS unit tests assert nothing real | tests | XS | None | DONE (2026-06-25) |
| F | `IsArchSupported` — no production caller | cleanup | XS | None | DONE (2026-06-25) |
| G | `MockVM.RunFunc` dead field | cleanup | XS | None | DONE (2026-06-25) |
| H | `MockVM.Configure` redundant override | cleanup | XS | None | DONE (2026-06-25) |
| I | `SerialConsoleStdio` dead const | cleanup | XS | None | DONE (2026-06-25) |
| J | `VM` interface + `self` collapse (contested) | cleanup | M | Med | skip — mock earns it |

---

## A — e2e tests silently skip; disk-image filenames drifted

**Where:** `cmd/q2boot/e2e_test.go:33-37`.

The `diskImages` map points at `sle-16.0-<arch>-135.5-textmode...qcow2`, but the
files in `diskimages/` are version **160.4**. `os.Stat` misses → every arch
subtest hits `t.Skipf` (`e2e_test.go:55-57`). `make e2e-test` reports success
while testing nothing.

**Fix:** stop hard-coding the version. Glob per arch instead, e.g.
`filepath.Glob("../../diskimages/sle-16.0-"+arch+"-*-textmode*.qcow2")` and take
the first match (still `t.Skip` if none). Survives the next version bump.

**Verify:** `make e2e-test` actually runs the QEMU boot subtests (look for
`login:` progress in `-v` output), not `--- SKIP`. Requires the real images +
QEMU locally.

**Note:** `diskimages/` is gitignored (not in the repo), so CI without the
images will still skip — that's expected; this fix is for local/dev runs.

## B — Default log file litters the working directory

**Where:** `config.go:15` (`DefaultLogFile = "q2boot.log"`), consumed in
`vm.go` `RunVM`; `.gitignore` does not list it.

`LogFile` defaults to the relative path `q2boot.log`, so **every** run writes it
into the current directory. The `?? q2boot.log` untracked file in the repo is
this firing during development. Run q2boot inside any git repo and it drops an
untracked file there.

**Decision needed (pick one):**
1. Default `LogFile = ""` (off); users opt in with `--log-file`. Cleanest, but
   the U7 "tail log on QEMU failure" feature then needs an explicit log file to
   have anything to tail. Changes current behavior.
2. Keep logging on by default but write to a non-cwd location —
   `$XDG_STATE_HOME/q2boot/q2boot.log` (fallback `~/.local/state/q2boot/`) or
   `os.TempDir()`. Keeps U7 working, stops cwd pollution.
3. Minimum: add `q2boot.log` to `.gitignore` and document the cwd behavior.

**Recommendation:** option 2. Defensively also do option 3 (gitignore) so a
stray relative log never gets committed.

**Verify:** run `go run ./cmd/q2boot <disk> --dry-run` from a clean dir → no
`q2boot.log` created there (dry-run shouldn't write one anyway); a real run
writes to the chosen state dir, and `git status` in a repo dir stays clean.

## C — `http.Get` has no timeout

**Where:** `downloader.go:97` (`http.Get(url)` uses `http.DefaultClient`).

A server that accepts the connection then stalls hangs the download forever (no
deadline). Self-inflicted hang, not remote code exec — low severity.

**Fix:** use a custom `http.Client` with **connection-phase** timeouts, NOT a
blanket `Timeout` (that cap would also kill a legitimately long multi-GB body
transfer). Set `Transport` `DialContext` timeout + `TLSHandshakeTimeout` +
`ResponseHeaderTimeout`; leave the body read unbounded (the existing
Content-Length cross-check in `downloadHTTP` already catches truncation).

**Verify:** unit test with an `httptest.Server` handler that sleeps past
`ResponseHeaderTimeout` before writing headers → `downloadHTTP` returns an error
promptly instead of blocking. Add to `downloader_test.go`.

## D — UEFI vars file sized from the CODE blob (latent)

**Where:** `specs.go:124-141` (`uefiArgs`) + `specs.go:175-188`
(`ensureVarsFile`).

The per-disk NVRAM vars file is created with the size of the **firmware CODE**
blob (`info.Size()` of `FirmwarePath`). The vars file should match the **VARS
template** size. It works today only because AAVMF CODE and VARS are both 64MiB
on current distros; a distro shipping mismatched sizes would feed QEMU a
wrong-sized pflash.

**Fix:** locate the matching vars template next to the code blob
(`AAVMF_VARS.fd` beside `AAVMF_CODE.fd`; `aavmf-aarch64-vars.bin` beside
`aavmf-aarch64-code.bin`) and size from it — copying the template is more
correct than zero-filling. If keeping zero-fill, at minimum size from the vars
template, and add a `// ponytail:` comment naming the 64MiB assumption.

**Verify:** aarch64 boot still reaches UEFI/GRUB with an existing and a
freshly-created `.vars` file.

## E — LUKS unit tests assert nothing real

**Where:** `detector_test.go:31-49`
(`TestHasLUKSPartitionMissing`/`Detected`).

Both reassign the `HasLUKSPartition` package var to a stub and then call the
stub — they verify Go lets you reassign a func var, not the detection. This is
why bug 001 shipped green. **Handle as part of 001:** delete these and add the
`e2e`-tagged real-image test described in 001's Test plan.

---

## Cleanup — over-engineering (ponytail-audit, 2026-06-25)

Free `delete:` wins (F–I), do in one commit. J is a contested structural call —
recommend **leaving it**.

### F — `IsArchSupported` has zero production callers
**Where:** `factory.go:22`. Only caller is `TestIsArchSupported` (`vm_test.go:155`).
**Cut:** delete the func and its test. ~-25 lines.

### G — `MockVM.RunFunc` is dead flexibility
**Where:** `mock_vm.go:7,34-39`. The `RunFunc` field is never set by any test;
`MockVM.Run` always falls through to `return nil`.
**Cut:** drop the field and the nil-check; `Run()` becomes `return nil`. ~-5 lines.

### H — `MockVM.Configure` is a redundant override
**Where:** `mock_vm.go:48-50`. Byte-identical to the `BaseVM.Configure` already
promoted via embedding — removing it still satisfies the `VM` interface.
**Cut:** delete the override. ~-3 lines.

### I — `SerialConsoleStdio` is a dead constant
**Where:** `vm.go:39` (`SerialConsoleStdio = "stdio"`). Declared, never referenced.
**Cut:** delete the line. -1 line.

### J — `VM` interface + `self VM` back-dispatch (contested — likely keep)
**Where:** `vm.go:42-52` (interface), `vm.go:58-62` (`self` field).
One production implementation (`archVM`); all 4 arches are the same data-driven
type, so cross-arch polymorphism is unnecessary. **But** the interface is
load-bearing for tests: `main_test.go` swaps `vm.CreateVM` to return a
`MockVM`, and `vm_test.go` drives `buildArgs`/`Run`/`Validate` through the mock.
Collapsing to methods on `*archVM` means reworking 3 test sites for ~-40 lines.
**Recommendation:** leave it — the mock makes it pay rent. Listed only so it's
on record as considered.

Verified-and-rejected: `config.Validate` empty-`DiskPath` guard looked dead but
is tested (`config_test.go:179`) — keep.

**Net cleanup: ~-35 lines (F–I) safe; ~-75 if J is pursued. -0 deps.**

---

## Direction (options for the maintainer)

- **riscv64 support — DEFERRED (2026-06-25, hardware not yet ready).**
  `diskimages/debian-13-...-riscv64-...qcow2` is already committed, but
  `archSpecs` (`specs.go:44`) has only 4 arches. When hardware is ready, adding
  one = an `ArchSpec` entry + a `archFilenameRe`/`detectByVirtCat` branch
  (`detector.go:15,57`). Note: riscv64 needs UEFI firmware paths and has no KVM
  on most hosts (TCG only). Spike-sized. **Do not start until hardware lands.**

- **Collapse libguestfs to one appliance boot.** Arch detection (virt-cat) and
  LUKS detection (guestfish, after 001) each boot the appliance (~1.5s). A
  non-graphical launch without `--arch` pays ~3 boots. If launch latency starts
  to matter, do a single guestfish inspection that returns both the ELF arch and
  the partition vfs-types. Only worth it if users complain.
