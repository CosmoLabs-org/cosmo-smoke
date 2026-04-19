# Feedback — How to Use

Manage incoming and outgoing cross-project feedback with quality standards.

## Commands

| Command | What it does |
|---------|--------------|
| `ccs feedback send` | Send feedback to another project |
| `ccs feedback list` | List all feedback |
| `ccs feedback done ID` | Mark feedback as resolved |
| `ccs feedback convert ID` | Convert feedback to issue |

## Workflow

1. Incoming feedback arrives in `incoming/`
2. Review and triage
3. Convert actionable items to issues or resolve
4. Outgoing feedback goes in `outgoing/`

## Conventions

- Feedback must include: what happened, why it matters, proposed solution
- Never shorter than 5 lines (unless `--bare`)
- `index.yaml` tracks all feedback items

## Related

- [Feedback loop instructions](~/.claude/instructions/feedback-loop.md)
