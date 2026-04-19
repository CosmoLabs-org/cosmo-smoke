# Parallel-Agent Merge-Conflict SOP

Handling merge conflicts when multiple GLM agents modify overlapping files during GOrchestra execution.

## Detection

After `ccs merge --all`, conflicts surface as:

```
CONFLICT (content): Merge conflict in internal/schema/schema.go
Automatic merge failed; fix conflicts and commit the result.
```

`ccs verify-worktree --all` also flags conflicting files pre-merge.

## Resolution Strategies

### Strategy 1: Sequential Re-merge

When conflict is minor (adjacent additions, not overlapping edits):

1. `git mergetool` or manual edit to resolve
2. `git add <files> && git commit`
3. Re-run `ccs verify-worktree` on the resolved merge

### Strategy 2: Rebase Agent onto Master

When agent A's work depends on agent B's output:

1. Merge agent B first: `ccs merge <agent-b>`
2. Rebase agent A: `git -C <worktree-a-dir> rebase master`
3. Resolve conflicts in worktree context
4. `ccs verify-worktree <agent-a> && ccs merge <agent-a>`

### Strategy 3: Cherry-Pick

When one agent's changes to a file are more complete:

1. Merge the more complete agent first
2. Cherry-pick specific commits from the other: `git cherry-pick <sha>`
3. Resolve conflicts per commit

## Prevention

### File Partitioning

The primary prevention strategy. In the plan phase, assign agents to non-overlapping files:

- Agent A: `internal/schema/` + `cmd/`
- Agent B: `internal/runner/` + `internal/reporter/`
- Agent C: `docs/` + `tests for A+B`

### Interface Contracts

When agents must share types, define the interface in the plan:

```go
// Both agents use this type — defined in plan, implemented once
type TraceResult struct {
    TraceID  string
    SpanCount int
}
```

Agent who "owns" the defining file implements it. Other agents depend on the interface.

### Shared-Stub Pattern

For new types both agents need:

1. Agent A creates a minimal stub in a shared file
2. Agent B works against the stub
3. Post-merge, the full implementation replaces the stub

### Tag-Based Isolation

Use Go build tags for truly independent implementations:

```go
// agent_a.go
//go:build agentA

// agent_b.go
//go:build !agentA
```

## Escalation

If resolution requires architectural decisions not in the plan:

1. Abort the conflicting merge: `git merge --abort`
2. Document the conflict: `ccs feedback send . "merge conflict between agents X and Y on file Z"`
3. Resolve in a focused session with full context

## Checklist

- [ ] Review plan for file overlap before dispatching agents
- [ ] Run `ccs verify-worktree --all` before batch merge
- [ ] Merge agents with shared files last
- [ ] Test after every merge resolution
