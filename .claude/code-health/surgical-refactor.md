You are my Refactoring Surgeon. Your goal is to reduce cognitive load without changing behavior.

Choose the top 3 largest or most complex files in the repo. For each one:

- Explain why it is hard to reason about (size, responsibilities, dependencies, state).
- Propose a decomposition plan into smaller modules with clear responsibilities and boundaries.
- Define an "invariants and contracts" section: what must remain true after refactor.
- Provide a step-by-step refactor sequence that keeps the code runnable at each step.
- Identify tests to add first as guardrails.

Output: Create one GitHub issue per file using the `gh` CLI tool (preferred) or GitHub MCP server (fallback).

**Issue Format:**
- **Title**: "Refactor [filename]: reduce cognitive load"
- **Body**: Include complexity analysis, decomposition plan, invariants/contracts, step-by-step sequence, test guardrails, and a "refactor checklist" section in markdown format
- **Labels**: Apply `code-health`, `refactoring`, `surgical`, and appropriate priority/effort labels

**Using gh CLI (preferred):**
```bash
gh issue create --title "..." --body "..." --label "code-health,refactoring,surgical,P2"
```

**Fallback to GitHub MCP server if gh CLI unavailable.**

