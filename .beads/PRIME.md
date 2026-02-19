# Beads Workflow Context

> **Context Recovery**: Run `bd prime` after compaction, clear, or new session
> Hooks auto-call this in Claude Code when .beads/ detected

## 🕵️ CODE REVIEW & REFINEMENT PROTOCOL

**1. Initiating Review**
When a task is functionally complete:
- Do NOT close the task yet.
- Create a review ticket: `bd create --title="Review: <Task Name>" --type=task`
- Link dependency: `bd dep add beads-<original_task_id> beads-<review_id>` (Original task blocks on Review)
- STOP: An agent never works on a review ticket he himself created. You're on standby now awaiting the next command.

**2. Performing Review (The Reviewer)**
- **Criteria**: Assess quality, maintainability, smells, and patterns.
- **Scoring**: Assign a mental score (0.0 - 10.0).
- **Unrelated Issues**: If you see unrelated defects, `bd create` separate issues for each immediately.
- **Decision Logic**:
    - **Score < 8.5**: MANDATORY Refinement.
        - `bd create --title="Refinement: <Task Name>" --type=task` (or `bug` if purely defects)
        - Description MUST list *every* defect and *what* needs fixing.
    - **Score >= 8.5**: Reviewer judgment. Pass or optional minor cleanup.

**3. Executing Refinement (The Implementer)**
- Implement fixes specified in the Refinement ticket.
- **Re-Review Decision**: Do I need a second review?
    - **NO**: If changes were small or structural refactors (e.g., renaming 1 variable in 20 files = 1 small change). -> Close Refinement & Original Task.
    - **YES**: If changes were "Too Much" (Lots of *different* small changes OR few large logic changes). -> Create new Review ticket (Type: Task).
## Core Rules
- **Default**: Use beads for ALL task tracking (`bd create`, `bd ready`, `bd close`)
- **Prohibited**: Do NOT use TodoWrite, TaskCreate, or markdown files for task tracking
- **Workflow**: Create beads issue BEFORE writing code, mark in_progress when starting
- Persistence you don't need beats lost context
- Git workflow: hooks auto-sync, run `bd sync` at session end
- Session management: check `bd ready` for available work

### Workflow Pattern

1. **Start**: Run `bd ready` to find actionable work
2. **Claim**: Use `bd update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Notice**: Any issues you discover should be captured with `bd create` 
5. **Pre-complete:** Now is the time to engage the CODE REVIEW & REFINEMENT PROTOCOL to see if you can close up shop or have extra steps
6. **Complete**: Use `bd close <id>`
7. **Sync**: Always run `bd sync` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `bd ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `bd dep add <issue> <depends-on>` to add dependencies

### Best Practices

- Check `bd ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `bd create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `bd sync` before ending session

## Essential Commands

### Finding Work
- `bd ready` - Show issues ready to work (no blockers)
- `bd list --status=open` - All open issues
- `bd list --status=in_progress` - Your active work
- `bd show <id>` - Detailed issue view with dependencies

### Creating & Updating
- `bd create --title="Summary of this issue" --description="Why this issue exists and what needs to be done" --type=task|bug|feature --priority=2` - New issue
  - Priority: 0-4 or P0-P4 (0=critical, 2=medium, 4=backlog). NOT "high"/"medium"/"low"
- `bd update <id> --status=in_progress` - Claim work
- `bd update <id> --assignee=username` - Assign to someone
- `bd update <id> --title/--description/--notes/--design` - Update fields inline
- `bd close <id>` - Mark complete
- `bd close <id1> <id2> ...` - Close multiple issues at once (more efficient)
- `bd close <id> --reason="explanation"` - Close with reason
- **Tip**: When creating multiple issues/tasks/epics, use parallel subagents for efficiency
- **WARNING**: Do NOT use `bd edit` - it opens $EDITOR (vim/nano) which blocks agents

### Dependencies & Blocking
- `bd dep add <issue> <depends-on>` - Add dependency (issue depends on depends-on)
- `bd blocked` - Show all blocked issues
- `bd show <id>` - See what's blocking/blocked by this issue

### Sync & Collaboration
- `bd sync` - Sync with git remote (run at session end)
- `bd sync --status` - Check sync status without syncing

### Project Health
- `bd stats` - Project statistics (open/closed/blocked counts)
- `bd doctor` - Check for issues (sync problems, missing hooks)

## Common Workflows

**Starting work:**
```bash
bd ready           # Find available work
bd show <id>       # Review issue details
bd update <id> --status=in_progress  # Claim it
```

**Creating dependent work:**
```bash
# Run bd create commands in parallel (use subagents for many items)
bd create --title="Implement feature X" --description="Why this issue exists and what needs to be done" --type=feature
bd create --title="Write tests for X" --description="Why this issue exists and what needs to be done" --type=task
bd dep add beads-yyy beads-xxx  # Tests depend on Feature (Feature blocks tests)
```
