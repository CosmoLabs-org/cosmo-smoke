# smoke validate

Validate smoke test config without running tests.

## Usage

```bash
smoke validate [-f path]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --file` | `.smoke.yaml` | Config file path |

## Description

Loads and validates `.smoke.yaml` configuration. Reports all errors at once rather than stopping at the first — useful for fixing multiple config issues in one pass.

## Examples

```bash
smoke validate                         # Validate .smoke.yaml in current dir
smoke validate -f staging.smoke.yaml   # Validate a specific config
smoke validate -f . && smoke run       # Validate then run
```
