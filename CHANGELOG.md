# Changelog

All notable changes to QBoot will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure
- Comprehensive unit testing framework
- Integration testing suite
- Performance benchmarking tests

### Changed
- Improved error handling and validation
- Enhanced code documentation

### Fixed
- Various bug fixes and stability improvements

## [1.0.0] - 2025-01-XX

### Added
- Core QEMU VM launcher functionality
- JSON-based configuration system with `~/.config/qboot/config.json`
- Command-line interface with comprehensive options
- Interactive mode with GUI support
- Headless mode for server deployments
- Snapshot mode (changes discarded) and persistent mode
- KVM hardware acceleration support
- SSH port forwarding configuration
- Virtio drivers for optimal performance
- Serial console logging
- Automatic hugepages memory allocation
- Configuration validation and error handling
- Comprehensive test suite with >95% coverage
- Unit tests for all core functions
- Integration tests for complete workflows
- Edge case testing and error condition handling
- Performance stress testing
- Mock QEMU execution for safe testing
- Build system with DUB and Make
- Cross-platform compatibility (Linux, macOS, Windows)
- Detailed documentation and examples
- Contributing guidelines and development setup
- Apache 2.0 license

### Features
- **Zero Configuration**: Works out of the box with sensible defaults
- **Flexible Configuration**: Override any setting via config file or command line
- **Safe Testing**: Comprehensive test suite prevents regressions
- **Developer Friendly**: Clear code structure and extensive documentation
- **Production Ready**: Robust error handling and validation

### Command Line Options
- `-d, --disk` - Specify disk image path (required)
- `-c, --cpu` - Set number of CPU cores (default: 2)
- `-r, --ram` - Set RAM in GB (default: 4)
- `-i, --interactive` - Enable GUI mode (default: headless)
- `-S, --no-snapshot` - Persist changes to disk (default: snapshot mode)
- `-l, --log` - Set log file path (default: console.log)
- `--ssh-port` - Set SSH port forwarding (default: 2222)
- `-h, --help` - Show help information

### Configuration Options
- `cpu` - Number of CPU cores (1-32)
- `ram_gb` - RAM in gigabytes (1-128)
- `ssh_port` - SSH forwarding port (1024-65535)
- `log_file` - Path to serial console log
- `headless_saves_changes` - Whether headless mode persists changes

### Technical Details
- Written in D programming language
- Uses DUB for dependency management and building
- Supports DMD, LDC2, and GDC compilers
- Generates optimized QEMU command lines
- Automatic KVM detection and configuration
- Virtio network and storage drivers
- Memory optimization with hugepages
- Safe parameter validation
- Graceful error handling and recovery

### Testing Infrastructure
- Unit tests for individual functions
- Integration tests for complete workflows
- Edge case testing for boundary conditions
- Performance testing and benchmarking
- Mock execution for safe testing
- Temporary file management for test isolation
- Comprehensive test coverage reporting
- Automated test running with Make targets

### Documentation
- Comprehensive README with usage examples
- Detailed API documentation
- Testing guide and best practices
- Contributing guidelines
- Code style and formatting rules
- Development setup instructions
- Troubleshooting and FAQ

[Unreleased]: https://github.com/yourusername/qboot/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yourusername/qboot/releases/tag/v1.0.0