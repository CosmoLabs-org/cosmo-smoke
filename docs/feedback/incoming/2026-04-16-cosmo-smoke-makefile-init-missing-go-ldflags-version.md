---
id: FB-001
title: makefile init missing Go ldflags version injection
type: idea
status: implemented
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: cosmo-smoke
to_target: self
created: "2026-04-16T09:29:01.672619-03:00"
updated: "2026-04-16T15:36:10.742569-03:00"
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

# FB-001: makefile init missing Go ldflags version injection

## Problem
`ccs makefile init` generates a Go build target without ldflags for version injection.

## Current output
```makefile
build:
    go build -o bin/$(BINARY_NAME) .
```

## Expected
```makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.Version=$(VERSION)

build:
    go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .
```

## Why it matters
- Binaries have no embedded version (`smoke version` returns empty)
- No strip flags means ~30% larger binaries
- CLAUDE.md documents ldflags but Makefile ignores it

## Additional gaps
1. No `release` target for cross-compilation (linux/darwin/windows)
2. No integration with .version-registry.json
3. `ccs read Makefile` returns empty summary (not a recognized filetype)

