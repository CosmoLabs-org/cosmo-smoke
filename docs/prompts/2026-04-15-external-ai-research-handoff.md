---
title: External AI Research — Universal Smoke Testing Landscape
created: "2026-04-15"
type: handoff
target_ais: [Grok, Gemini, ChatGPT, Perplexity]
purpose: Research open source smoke testing tools and patterns
---

# Research Handoff: Universal Smoke Testing Solutions

## Context

I'm building **cosmo-smoke**, a standalone universal smoke test runner. It's a Go binary that reads `.smoke.yaml` from any project root and runs lightweight "does it turn on?" verification tests.

**Current capabilities:**
- 7 assertion types: exit_code, stdout/stderr contains/matches, file_exists, env_exists
- 3 output formats: terminal (colored), JSON, JUnit XML
- Auto-detection: Go, Node, Python, Docker, Rust project scaffolding
- CLI: `smoke run`, `smoke init`, `smoke version`

**What makes it different:** Config-driven (YAML), language-agnostic, focused purely on smoke tests (not unit tests), designed for portfolio-scale use (95+ projects).

## Research Questions

### 1. Existing Open Source Tools

Find me open source projects that do similar things:
- Universal/polyglot smoke test runners
- Config-driven test execution frameworks
- "Health check" or "sanity test" tools
- Build verification tools

For each tool found, tell me:
- Name, repo URL, stars/activity
- What it does well
- What it lacks
- How it compares to our YAML-config approach

### 2. Assertion Patterns

What assertion types do other smoke/health check tools support that we might be missing?

We have: exit_code, stdout_contains, stdout_matches, stderr_contains, stderr_matches, file_exists, env_exists

What else exists in the wild?
- HTTP endpoint checks?
- JSON/XML response parsing?
- Database connectivity?
- Service mesh health?
- Container/pod status?
- Performance thresholds?

### 3. CI/CD Integration Patterns

How do other tools integrate with CI/CD pipelines?
- GitHub Actions patterns
- GitLab CI patterns
- Jenkins patterns
- Reusable workflow designs

### 4. Portfolio-Scale Patterns

For organizations with 50-100+ repos:
- How do they aggregate smoke test results?
- Dashboard solutions?
- Cross-repo test orchestration?
- Monorepo vs polyrepo patterns?

### 5. Configuration Patterns

What config formats/patterns work well?
- YAML vs TOML vs JSON
- Config inheritance/includes
- Environment-specific overrides
- Secret handling in configs

### 6. Interesting Features We Haven't Thought Of

What innovative features exist in similar tools that we should consider?
- Plugin systems?
- Test generation from OpenAPI specs?
- AI-assisted test creation?
- Visual regression for CLIs?
- Anything surprising or clever?

## Output Format

Please structure your response as:

```markdown
## Tool: [Name]
- **Repo**: [URL]
- **Stars/Activity**: [X stars, last commit Y]
- **Strengths**: [bullet points]
- **Weaknesses**: [bullet points]
- **Relevant patterns to steal**: [specific features]

## Assertion Ideas
[Table of assertion types found in the wild]

## CI Integration Patterns
[Specific patterns worth copying]

## Portfolio-Scale Solutions
[Tools or patterns for multi-repo scenarios]

## Innovative Features
[Surprising/clever ideas we should consider]

## Recommended Next Steps
[Prioritized list of what to research deeper or implement]
```

## Specific Tools to Investigate (if they exist)

- Bats (Bash Automated Testing System)
- Terratest
- Testinfra / Goss
- Pact (contract testing)
- Karate DSL
- Robot Framework
- Gauge
- Any "smoke" or "sanity" specific tools

## What I'll Do With This Research

Feed findings back into cosmo-smoke's roadmap. Looking for:
1. Features we should add
2. Patterns we should follow
3. Mistakes we should avoid
4. Integration opportunities

---

**Copy this entire prompt to Grok, Gemini, ChatGPT, or Perplexity and ask them to research.**
