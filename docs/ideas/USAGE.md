# Ideas — How to Use

Capture, track, and promote ideas for future features and improvements.

## Commands

| Command | What it does |
|---------|--------------|
| `ccs idea add "Title"` | Add a new idea |
| `ccs idea list` | List all ideas |
| `ccs idea promote SLUG --type feature` | Promote idea to issue |
| `ccs idea backfill-ids` | Add missing IDs to legacy ideas |
| `/idea "title"` | Quick idea capture via skill |

## Workflow

1. Capture ideas as they arise with `/idea` or `ccs idea add`
2. Ideas get dated markdown files with metadata
3. Promote actionable ideas to issues with `ccs idea promote`
4. Track status from capture to implementation

## Conventions

- Each idea has a unique slug and optional Base36 ID
- Include: problem, proposed solution, potential approach
- Never create files directly — use commands

## Related

- [Issue tracking](../issues/USAGE.md)
- [Roadmap](../roadmap/USAGE.md)
