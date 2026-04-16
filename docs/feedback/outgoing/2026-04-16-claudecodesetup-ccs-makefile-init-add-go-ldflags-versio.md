---
id: FB-466
title: 'ccs makefile init: add Go ldflags version injection'
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-16T09:29:09.138611-03:00"
updated: "2026-04-16T09:29:09.138611-03:00"
suggested_conversion: feature
converted_to: null
related_issues: []
brainstorm_ref: null
suggested_workflow: []
response:
  acknowledged: null
  acknowledged_by: null
  started: null
  implemented: null
  rejected: null
  rejection_reason: null
  notes: ""
---

# FB-466: ccs makefile init: add Go ldflags version injection

## Problem
`ccs makefile init` generates Go build targets without ldflags for version injection.

## Current output (cosmo-smoke)
```makefile
build:
    go build -o bin/$(BINARY_NAME) .
```

## Expected
```makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || cat .version 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.Version=$(VERSION)

build:
    go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .
```

## Why it matters
- Binaries have no embedded version (e.g., `smoke version` returns empty)
- Missing `-s -w` means ~30% larger binaries
- Project CLAUDE.md documents ldflags pattern but generated Makefile ignores it

## Additional gaps
1. No `release` or `release-all` target for cross-compilation
2. No integration with .version-registry.json (could read version from there)
3. Minor: `ccs read Makefile` returns empty summary — Makefile not recognized as parseable filetype

