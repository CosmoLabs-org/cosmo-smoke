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

Supported project types (31 total):

**Languages:** Go, Node (bun/npm), Python, Rust, Java (Maven), Java (Gradle), .NET/C#, Ruby, PHP, Deno, Scala, Elixir, Swift (server), Dart (server), Zig, Haskell, Lua, C/C++ (Make), C/C++ (CMake)

**Mobile:** React Native, Flutter, iOS, Android

**Infrastructure:** Docker, Terraform, Helm, Kustomize, Serverless

**Static Sites:** Hugo, Astro, Jekyll

## Examples

```bash
smoke init                             # Auto-detect and generate
smoke init --force                     # Overwrite existing config
smoke init --from-running my-container # Generate from running container
```
