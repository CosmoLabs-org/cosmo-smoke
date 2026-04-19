# Issues — How to Use

Track feature requests, bugs, and tasks with deterministic IDs and YAML format.

## Commands

| Command | What it does |
|---------|--------------|
| `ccs issues` | List all issues |
| `ccs issues create --type feature --description "..." "Title"` | Create a feature issue |
| `ccs issues update ID --status in-progress` | Update issue status |
| `/feature "title"` | Create feature via skill |
| `/bug "title"` | Create bug via skill |
| `/task "title"` | Create task via skill |

## Workflow

1. Create issue via `/feature`, `/bug`, or `/task` commands
2. Issue gets deterministic ID (FEAT-NNN, BUG-NNN, TASK-NNN)
3. Update status as work progresses
4. Close when complete

## Conventions

- Never create files in `docs/issues/` directly — use commands
- Each issue is a YAML file with frontmatter
- `docs/issues.yaml` is the index (auto-updated)

## Related

- [Issue tracking instructions](~/.claude/instructions/issue-tracking.md)
