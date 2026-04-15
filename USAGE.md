# smoke — How to Use

Runs lightweight smoke tests defined in `.smoke.yaml` to verify a project is functional.

## Commands

| Command | What it does |
|---------|--------------|
| `smoke run` | Run all smoke tests in `.smoke.yaml` |
| `smoke run --tag <tag>` | Run only tests matching the given tag |
| `smoke run --exclude-tag <tag>` | Run all tests except those with the given tag |
| `smoke run --format json` | Output results as JSON (for CI pipelines) |
| `smoke run --fail-fast` | Stop immediately on the first failure |
| `smoke run --timeout <dur>` | Override per-test timeout (e.g. `60s`, `2m`) |
| `smoke run --dry-run` | List matching tests without executing them |
| `smoke run -f <path>` | Use a config file at a non-default path |
| `smoke init` | Auto-detect project type and generate `.smoke.yaml` |
| `smoke init --force` | Overwrite an existing `.smoke.yaml` |
| `smoke version` | Print the binary version |

## Workflow

1. Run `smoke init` to generate a `.smoke.yaml` in your project root.
2. Edit the generated config — add real commands, adjust timeouts, tag tests.
3. Run `smoke run` to execute all tests.
4. Use `--tag` to run focused subsets (e.g. `--tag build` in CI, `--tag runtime` locally).
5. Use `--format json` to integrate results into CI pipelines or monitoring systems.

## Flags Reference

### `smoke run`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --file` | string | `.smoke.yaml` | Config file path |
| `--tag` | string (repeatable) | — | Include only tests with this tag |
| `--exclude-tag` | string (repeatable) | — | Exclude tests with this tag |
| `--format` | string | `terminal` | Output format: `terminal` or `json` |
| `--fail-fast` | bool | `false` | Stop on first failure |
| `--timeout` | duration | _(from config)_ | Per-test timeout override |
| `--dry-run` | bool | `false` | List tests without running |

### `smoke init`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f, --force` | bool | `false` | Overwrite existing `.smoke.yaml` |

## Conventions

- Config file is `.smoke.yaml` at the project root. Use `-f` to override.
- Commands in `run` and `prerequisites.check` execute via the system shell (`sh -c`).
- All commands run from the directory containing the config file, not the caller's cwd.
- `file_exists` paths are relative to the config file's directory.
- Timeouts use Go duration strings: `30s`, `2m`, `1m30s`.
- Tags are free-form strings. A test with no tags is always included unless `--tag` is set.
- `cleanup` runs after the test regardless of pass/fail, but its exit code is ignored.
- Prerequisites run before all tests. If any prerequisite fails, the run aborts.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All tests passed |
| `1` | One or more tests failed |
| `2` | Config error or invalid arguments |

## Related

- [README.md](./README.md) — Overview and quick start
- [SPEC.md](./SPEC.md) — Full schema reference
- [.smoke.yaml](./.smoke.yaml) — This project's own smoke tests
