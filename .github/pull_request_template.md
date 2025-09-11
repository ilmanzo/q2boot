## Description
Brief description of the changes in this pull request.

## Type of Change
Please check the relevant option(s):

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update (changes to documentation only)
- [ ] 🔧 Refactoring (code changes that neither fix bugs nor add features)
- [ ] ⚡ Performance improvement
- [ ] 🧪 Test improvements (adding or updating tests)
- [ ] 🚀 Build/CI improvements
- [ ] 🧹 Chore (maintenance, dependency updates, etc.)

## Related Issues
Closes #(issue_number)
Fixes #(issue_number)
Related to #(issue_number)

## Changes Made
Please describe the changes made in detail:

- Change 1
- Change 2
- Change 3

## Testing Performed
Describe the testing you have performed:

- [ ] Unit tests added/updated and passing
- [ ] Integration tests passing
- [ ] Manual testing performed
- [ ] Tested on multiple platforms (specify which ones)
- [ ] Performance impact assessed

### Test Commands Run
```bash
make test
make test-verbose
dub build --build=release
./qboot --help
```

## Configuration Changes
If this PR introduces configuration changes:

- [ ] Updated default configuration
- [ ] Added new configuration options
- [ ] Deprecated existing options
- [ ] Updated configuration documentation

Example configuration:
```json
{
  "new_option": "default_value"
}
```

## Breaking Changes
If this is a breaking change, describe:

- What breaks
- Migration path for users
- Deprecation timeline (if applicable)

## Documentation
- [ ] Code is self-documenting with clear variable/function names
- [ ] Added/updated inline code comments where needed
- [ ] Updated README.md (if applicable)
- [ ] Updated TESTING.md (if applicable)
- [ ] Updated CONTRIBUTING.md (if applicable)
- [ ] Updated CHANGELOG.md

## Code Quality
- [ ] Code follows the project's coding standards
- [ ] Self-review of code completed
- [ ] Code is properly formatted (`make format`)
- [ ] No lint warnings (`make lint`)
- [ ] No compiler warnings
- [ ] Memory safety considerations addressed

## Performance Impact
- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance regression (justified and documented)
- [ ] Performance benchmarks run (attach results)

## Security Considerations
- [ ] No security impact
- [ ] Security improvement
- [ ] Potential security implications (documented and reviewed)
- [ ] Input validation added/updated
- [ ] Error handling doesn't leak sensitive information

## Platform Compatibility
Tested on:
- [ ] Linux (Ubuntu/Debian)
- [ ] Linux (CentOS/RHEL)
- [ ] macOS
- [ ] Windows
- [ ] Other: _______________

## Dependencies
- [ ] No new dependencies added
- [ ] New dependencies added (justified and documented)
- [ ] Dependencies updated
- [ ] Dependencies removed

## Deployment Notes
Any special deployment considerations:
- Environment variables
- Configuration file changes
- Database migrations
- Service restarts required

## Screenshots/Examples
If applicable, add screenshots or examples of the changes:

```bash
# Example command
qboot -d disk.img --new-feature
```

## Checklist
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## Additional Notes
Add any additional notes, concerns, or questions for reviewers:

---

**For Reviewers:**
Please ensure all checklist items are completed before approving.