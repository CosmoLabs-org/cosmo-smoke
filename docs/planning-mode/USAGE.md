# Planning Mode — How to Use

Store implementation plans with structured format and issue linkage.

## Commands

| Command | What it does |
|---------|--------------|
| `ccs prompts init FILE --issue ID` | Enrich plan with frontmatter |
| `ccs prompts status` | Check plan statuses |

## Workflow

1. Create plan in `docs/planning-mode/YYYY-MM-DD-title.md`
2. Add frontmatter with issue ID and roadmap linkage
3. Execute plan steps, updating status as you go
4. Mark COMPLETED when done

## Conventions

- Plans start with clear goal/objective
- End with specific implementation steps
- Always save to this directory (never `.claude/plans/`)
- Use `/continuation-prompt` for session handoff

## Related

- [Planning instructions](~/.claude/rules/planning.md)
