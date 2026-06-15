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

Interface + composition. `VM` interface in `vm.go`; each arch (`x86_64.go`, `aarch64.go`, `ppc64le.go`, `s390x.go`) embeds `*BaseVM` and supplies:
- `QEMUBinary()`, `GetArchArgs()`, `GetDiskArgs()`, `GetNetworkArgs()`, `GetGraphicalArgs()`
- optionally overrides `GetNonGraphicalDisplayArgs()` (e.g. x86_64 uses `-nographic` when a logfile is set; default is `-display curses`)
- thin `Validate()` and `Run()` that delegate back to `BaseVM.Validate(vm)` / `BaseVM.run(vm)` so the base can call arch methods via the interface

`BaseVM.buildArgs` is the single place QEMU args are assembled; arch-specific code only returns its slice. Display/monitor handling has subtle interactions: graphical implemented via `-nographic` (s390x) suppresses default monitor; non-graphical adds `-snapshot` unless `WriteMode`; logfile mode rewires `-chardev`/`-mon`/`-serial` to mux stdio into the file.

`CreateVM` and `detector.DetectArchitecture` are declared as `var = func(...)` (not `func`) so tests can swap them. `internal/vm/mock_vm.go` provides a test double.

### Config (`internal/config`)

`VMConfig` carries both `json` and `mapstructure` tags (viper unmarshals via mapstructure; JSON is for the persisted config file). Validation enforces CPU/RAM/port ranges and `disk_path` existence.

### Adding a new architecture

1. New `internal/vm/<arch>.go` embedding `*BaseVM`, implementing the interface methods above.
2. Add the case to `CreateVM` in `factory.go` **and** the literal in `SupportedArchitectures()`.
3. Add the same string to `detector.SupportedArchitectures` and (if needed) an ELF match branch in `detectByVirtCat`.
4. Extend `GetInstallationInstructions` in `vm.go` with the per-distro package names.

## Notes

- `cmd/q2boot/check.go` aliases `os.ReadFile`/`os.Stat`/`os.OpenFile` as package vars for test injection — preserve this when editing.
- Defaults live in two places by design: viper `SetDefault` in `initConfig` for the config-file path, and the `Default*` constants in `internal/config` / `internal/vm` for programmatic construction.
