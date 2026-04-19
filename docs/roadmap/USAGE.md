# Roadmap — How to Use

Manage project roadmap with items, priorities, and timeline tracking.

## Commands

| Command | What it does |
|---------|--------------|
| `ccs roadmap` | Show roadmap status |
| `ccs roadmap add "Title"` | Add roadmap item |
| `ccs roadmap update ID --status done` | Update item status |
| `ccs roadmap init` | Initialize roadmap from legacy files |

## Workflow

1. Add roadmap items for planned work
2. Set priority and timeline
3. Link to issues when implementation starts
4. Update status as work progresses

## Conventions

- Never edit `docs/roadmap/` files directly — use `ccs roadmap` commands
- Items stored as YAML in `items/` directory
- `index.yaml` tracks all items with metadata

## Related

- [Roadmap management](~/.claude/instructions/roadmap.md)
