# smoke init

Auto-detect project type and generate a `.smoke.yaml` configuration.

## Usage

```bash
smoke init [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --force` | `false` | Overwrite existing `.smoke.yaml` |
| `--from-running` | (none) | Generate config by inspecting a running Docker container |

## Description

By default, `smoke init` scans the current directory for project markers (Go modules, Node packages, Python projects, Dockerfiles, Rust Cargo.toml) and generates a tailored `.smoke.yaml` with appropriate smoke tests.

Supported project types: Go, Node (bun/npm), Python, Docker, Rust.

## Examples

```bash
smoke init                             # Auto-detect and generate
smoke init --force                     # Overwrite existing config
smoke init --from-running my-container # Generate from running container
```
