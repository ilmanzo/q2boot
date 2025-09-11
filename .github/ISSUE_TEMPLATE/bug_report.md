---
name: Bug Report
about: Create a report to help us improve QBoot
title: '[BUG] '
labels: bug
assignees: ''

---

## Bug Description
A clear and concise description of what the bug is.

## Environment
- **OS**: [e.g. Ubuntu 22.04, macOS 13.0, Windows 11]
- **Architecture**: [e.g. x86_64, arm64]
- **D Compiler**: [e.g. DMD v2.105.0, LDC2 v1.34.0]
- **QEMU Version**: [e.g. 7.2.0]
- **QBoot Version**: [e.g. 1.0.0, commit hash]

## Steps to Reproduce
1. Go to '...'
2. Run command '...'
3. Use configuration '...'
4. See error

## Expected Behavior
A clear and concise description of what you expected to happen.

## Actual Behavior
A clear and concise description of what actually happened.

## Error Messages
```
Paste any error messages, stack traces, or log output here
```

## Configuration File
If relevant, paste your `~/.config/qboot/config.json`:
```json
{
  "cpu": 2,
  "ram_gb": 4,
  "ssh_port": 2222,
  "log_file": "console.log",
  "headless_saves_changes": false
}
```

## Command Used
```bash
qboot -d disk.img -c 4 -r 8
```

## Additional Context
Add any other context about the problem here:
- Screenshots
- Related issues
- Workarounds you've tried
- Any other relevant information

## Checklist
- [ ] I have searched existing issues to ensure this is not a duplicate
- [ ] I have provided all the required environment information
- [ ] I have included steps to reproduce the issue
- [ ] I have included relevant error messages/logs
- [ ] I have tested with the latest version of QBoot