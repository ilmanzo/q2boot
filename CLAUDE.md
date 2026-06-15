# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
make build              # fmt + vet + build -> build/q2boot
make test               # go test -v ./...
make test-coverage      # writes coverage.html
make e2e-test           # go test -v -tags=e2e ./cmd/q2boot
make lint               # golangci-lint (if installed)

# Single test / package
go test -v -run TestName ./internal/vm
go test -v ./internal/config
```

Release builds bake `version`/`commit`/`buildTime` via `-ldflags` (see Makefile).

## Architecture

Entry: `cmd/q2boot/main.go` (cobra + viper). Subcommands: root `runQ2Boot`, `version`, `check` (`check.go`).

Flow in `runQ2BootE`:
1. If disk path is a remote URL (`http/https/ftp/smb`), `downloader.Download` fetches to a temp file with a `cleanup` func (deferred).
2. `applyFlagOverrides` merges CLI flags into the viper-loaded `*config.VMConfig`. Precedence: **CLI flag > config file > viper default**. Flag-vs-config detection uses `cmd.Flags().Changed(name)` for bools/ports because their zero values are ambiguous.
3. If `--arch` was not set, `detector.DetectArchitecture` runs: tries `virt-cat <disk> /bin/sh | file -` (ELF arch sniff), falls back to filename substring match (`-arch`, `_arch`, `@arch`).
4. `cfg.Validate()` → `vm.ValidatePortsAvailable` → `vm.IsArchSupported` → `vm.CreateVM(arch)` → `Configure(cfg)` → `Validate()` → `Run()`.

### VM package (`internal/vm`)

Data-table architecture. **All four QEMU architectures share one implementation** (`archVM` in `specs.go`); the only per-arch surface is an `ArchSpec` entry in `archSpecs` describing:
- binary name + machine/cpu args
- disk/net bus device + IDs + extra device options
- whether `--graphical` should fall back to a serial console (ppc64le, s390x)
- optional UEFI firmware search paths (aarch64)
- per-distro package names (Ubuntu/SUSE/Arch) for install hints

`BaseVM` holds the shared runtime state (CPU/RAM/ports/etc.) plus a `self VM` pointer set in each constructor so the promoted `Validate()` / `Run()` methods can dispatch into the concrete type's arch-specific methods. **Forgetting `self = v` in a new constructor will nil-deref at Validate/Run time.**

`BaseVM.buildArgs` is the single place QEMU args are assembled; arch-specific code only returns its slice. Display/monitor handling has subtle interactions: graphical implemented via `-nographic` (s390x) suppresses default monitor; non-graphical adds `-snapshot` unless `WriteMode`; logfile mode rewires `-chardev`/`-mon`/`-serial` to mux stdio into the file.

`CreateVM` and `detector.DetectArchitecture` are declared as `var = func(...)` (not `func`) so tests can swap them. `internal/vm/mock_vm.go` provides a test double.

### Config (`internal/config`)

`VMConfig` carries both `json` and `mapstructure` tags (viper unmarshals via mapstructure; JSON is for the persisted config file). Validation enforces CPU/RAM/port ranges and `disk_path` existence.

### Adding a new architecture

1. Append an `ArchSpec` entry to `archSpecs` in `internal/vm/specs.go`.
2. Add the arch token to the regex in `internal/detector/detector.go` (`archFilenameRe`) and add a virt-cat ELF-match branch in `detectByVirtCat`.

That's it. `factory.go` and the version/check commands derive everything they need from the spec table.

## Notes

- `cmd/q2boot/check.go` aliases `os.ReadFile`/`os.Stat`/`os.OpenFile` as package vars for test injection — preserve this when editing.
- Defaults live in two places by design: viper `SetDefault` in `initConfig` for the config-file path, and the `Default*` constants in `internal/config` / `internal/vm` for programmatic construction.
