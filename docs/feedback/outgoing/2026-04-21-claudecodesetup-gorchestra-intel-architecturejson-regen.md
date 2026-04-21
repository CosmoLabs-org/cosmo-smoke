---
id: FB-617
title: GOrchestra intel architecture.json regenerates after every commit, creating commit noise
type: improvement
status: pending
priority: low
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-21T16:27:29.878208-03:00"
updated: "2026-04-21T16:27:29.878208-03:00"
suggested_conversion: feature
converted_to: null
related_issues: []
brainstorm_ref: null
session: 2027
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

# FB-617: GOrchestra intel architecture.json regenerates after every commit, creating commit noise

Post-commit hook regenerates GOrchestra/intel/architecture.json after every commit, creating an infinite cycle of uncommitted changes.

Current behavior:
After any commit, git status immediately shows:
  M GOrchestra/intel/architecture.json

The file updates its generated_at timestamp, last_analyzed_commit SHA, and summary text. This forces an extra commit every time, polluting git history. In the session where this was observed, 4 of 9 commits (44%) were just 'chore: regenerate GOrchestra intel metadata' — zero informational value.

Expected behavior:
Either: (a) include the regenerated file in the same commit batch automatically, (b) add it to .gitignore since it regenerates on every commit anyway, or (c) skip regeneration when the delta is only timestamp/SHA (no structural change).

Why it matters:
Git history becomes noisy. Release notes and changelog generation pick up these commits. Code archaeology (git log, git blame) gets diluted. The signal-to-noise ratio degrades with every session.

Priority justification:
Low — not blocking any workflow, but accumulates technical debt in git history. Each session adds 2-4 meaningless commits.

Reproduction steps:
1. Make any commit (e.g. ccs commit-batch)
2. Run git status
3. Observe GOrchestra/intel/architecture.json modified
4. Commit it
5. Run git status again — it's modified again

Affected files:
- GOrchestra/intel/architecture.json (the regenerating file)
- The post-commit hook that triggers regeneration (location unknown — possibly cosmohooks)
- .gitignore (potential fix location)

Suggested implementation:
Add GOrchestra/intel/architecture.json to .gitignore. It regenerates automatically and is never manually edited. Alternatively, commit-batch could auto-include it as a final step in every batch.

