# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-04-16

### Added
- # FEAT-006: TAP output format

**Type**: feature
**Status**: closed
**Created**: 2026-04-16

## Description

Test Anything Protocol output for broader CI compatibility. Simpler than JUnit, widely supported.
- # FEAT-005: Process running assertion

**Type**: feature
**Status**: closed
**Created**: 2026-04-16

## Description

New assertion: process_running. Check if process exists by name or pattern. Syntax: process_running: 'nginx'. For daemon/service smoke tests.
- add grpc_health assertion via standard health protocol (commit:67532938)
- add redis_ping and memcached_version assertions (commit:ec481107)
- add ssl_cert assertion for TLS certificate validation (commit:1ddd7880)
- add prometheus text-format output (commit:e1912cd8)
- add response_time_ms threshold assertion (commit:d3981133)
- add allow_failure flag for flaky tests (commit:46999819)
- add process_running assertion type (commit:a778725b)

### Fixed
- harden process_running after Opus review (commit:2023700e)
- inject version via ldflags in Makefile (commit:d5d6c5f0)

## [0.2.0] - 2026-04-16

### Added
- add --from-running container inspection (commit:47c889a2)
- add config inheritance with includes and templates (commit:9ea0581c)
- add HTTP endpoint and JSON field assertions (commit:ab31b785)
- add TAP v14 output format (commit:8eb6794a)
- add port_listening assertion type (commit:db28d927)
- add JUnit XML output format (commit:5a2a2cbb)
- add stderr_matches and env_exists assertion types (commit:df0c7e3a)
- add HTTP health endpoint for container probes (commit:bddc019b)

