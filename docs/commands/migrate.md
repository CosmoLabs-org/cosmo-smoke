# smoke migrate

Migrate configuration files from other test frameworks to cosmo-smoke format.

## Subcommands

### goss

Migrate a Goss YAML configuration to `.smoke.yaml`.

```bash
smoke migrate goss <input.yaml> [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | (stdout) | Output `.smoke.yaml` path |
| `--overwrite` | `false` | Overwrite output file if it exists |
| `--strict` | `false` | Fail on any unmappable assertion |
| `--stats` | `false` | Print mapping stats to stderr |
| `--distro` | `deb` | Linux distro for package commands: `deb`, `rpm`, `apk` |

#### Supported Goss Resources

Core keys (process, port, command, file, http, package, service) map to native cosmo-smoke assertions. Other keys use command fallback with TODO comments for unsupported attributes.

#### Examples

```bash
smoke migrate goss goss.yaml                       # Print to stdout
smoke migrate goss goss.yaml -o .smoke.yaml        # Write to file
smoke migrate goss goss.yaml --strict --stats      # Strict mode with stats
smoke migrate goss goss.yaml --distro rpm          # RPM-based distro
```
