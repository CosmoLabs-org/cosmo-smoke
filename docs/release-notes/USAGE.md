# Release Notes — How to Use

Per-version release notes documenting features, fixes, and changes.

## Commands

| Command | What it does |
|---------|--------------|
| `ccs changelog audit` | Audit release notes for completeness |
| `ccs changelog migrate-all` | Migrate all changelog entries |

## Workflow

1. Changes tracked in `docs/changelog/unreleased.yaml`
2. On release, entries become release notes
3. Each version gets a dedicated release notes file
4. Audit checks for missing versions or placeholders

## Conventions

- Named `cosmo-smoke-vX.Y.Z-ReleaseNotes-*.md`
- Never edit CHANGELOG.md directly — use `ccs changelog` commands
- Include features, fixes, and breaking changes

## Related

- [Changelog management](~/.claude/instructions/changelog-management.md)
