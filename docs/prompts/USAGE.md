# Prompts — How to Use

Track prompt/plan files with metadata frontmatter and status lifecycle.

## Commands

| Command | What it does |
|---------|--------------|
| `ccs prompts status` | Show prompt statuses |
| `ccs prompts migrate` | Migrate UNKNOWN entries |
| `ccs prompts set-status FILE STATUS` | Update prompt status |
| `ccs prompts init FILE --issue ID` | Enrich with frontmatter |

## Workflow

1. Prompts are created as part of planning or session work
2. Each prompt has frontmatter: issue, status, created date
3. Statuses: PENDING → IN_PROGRESS → COMPLETED
4. UNKNOWN entries need migration (missing frontmatter)

## Conventions

- All prompts should have YAML frontmatter
- Status tracks lifecycle from creation to completion
- Use `ccs prompts migrate` to fix entries without frontmatter

## Related

- [Planning instructions](../../.claude/instructions/planning-mode.md)
