# Sessions — How to Use

Track session transcripts and session-level documentation.

## Commands

| Command | What it does |
|---------|--------------|
| `ccs sessions list` | List sessions |
| `ccs sessions show ID` | Show session details |

## Workflow

1. Sessions are auto-tracked via transcripts
2. Session-end creates summaries and documentation
3. Reference past sessions for context

## Conventions

- Transcripts stored as JSONL files
- Session summaries in `transcripts/` subdirectory
- Don't edit transcripts manually

## Related

- [Session recovery](~/.claude/instructions/session-recovery.md)
