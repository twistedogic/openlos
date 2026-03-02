# openlos — Design Document

> Virtual assistant workflow app built on picoclaw

---

## Overview

`openlos` is a **picoclaw-powered productivity assistant** — it uses picoclaw as the AI agent with a Go CLI backend for data persistence via SQLite. The app lives in a `.picoclaw/` directory and activates when `picoclaw agent` is run from the workspace. Users interact with the AI through natural language, and the assistant manages their ideas, tasks, goals, and schedule using structured SQLite storage.

---

## Goals

- Log and organize ideas, tasks, and goals via natural language
- Track goal progress and task status persistently (SQLite)
- Generate a daily digest (morning brief) on demand
- Break down goals into prioritized, actionable tasks
- Run anywhere picoclaw runs ($10 hardware, any Linux board)

---

## Architecture

```
openlos/
├── data/                          # gitignored — SQLite database
│   └── openlos.db
├── .picoclaw/                     # picoclaw workspace
│   ├── config.json                # picoclaw configuration
│   ├── AGENTS.md                 # "assistant" agent definition
│   ├── TOOLS.md                  # custom tool definitions
│   ├── skills/
│   │   ├── goal-tracking.md
│   │   ├── task-planning.md
│   │   └── daily-digest.md
│   └── bin/
│       └── openlos               # Go CLI binary
└── .gitignore                    # data/ already ignored
```

---

## Data Layer

All user data is stored in SQLite via the Go CLI (`data/openlos.db`). The CLI exposes commands for CRUD operations, and picoclaw tools invoke these commands via `exec`.

### Schema (via SQLite)

```sql
-- ideas table
CREATE TABLE ideas (
  id TEXT PRIMARY KEY,
  text TEXT NOT NULL,
  created INTEGER NOT NULL
);

-- goals table
CREATE TABLE goals (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'active',
  created INTEGER NOT NULL
);

-- tasks table
CREATE TABLE tasks (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  goal_id TEXT REFERENCES goals(id),
  status TEXT NOT NULL DEFAULT 'open',
  created INTEGER NOT NULL,
  due TEXT
);

-- schedule table
CREATE TABLE schedule (
  date TEXT PRIMARY KEY,
  focus TEXT NOT NULL,
  blocks TEXT
);
```

---

## Components

### 1. Custom Agent — `assistant`

**File:** `.picoclaw/AGENTS.md`

The assistant is a picoclaw agent pre-prompted to behave as a personal productivity assistant. It:

- Understands the tool surface and SQLite-backed data model
- Proactively invokes custom tools when the user makes natural-language requests
- Loads the appropriate skill (`goal-tracking`, `task-planning`, `daily-digest`) when performing domain-specific reasoning
- Avoids making code changes; focused entirely on productivity workflows

### 2. Custom Tools

**File:** `.picoclaw/TOOLS.md`

Tools are defined in picoclaw's markdown format and execute the Go CLI via `exec`. The CLI is invoked as `./bin/openlos <command>` from the workspace.

| Tool | CLI Command | Description |
|---|---|---|
| `ideas_log` | `ideas log --text` | Log a new idea |
| `ideas_list` | `ideas list` | List all ideas |
| `tasks_add` | `tasks add --title [--goal-id] [--due]` | Add a new task |
| `tasks_list` | `tasks list [--status] [--goal-id]` | List tasks with optional filters |
| `tasks_update` | `tasks update --id --status [--due]` | Update task status or due date |
| `goals_add` | `goals add --title [--description]` | Add a new goal |
| `goals_list` | `goals list [--status]` | List goals with optional status filter |
| `goals_update` | `goals update --id [--status] [--description]` | Update goal status or description |
| `schedule_read` | `schedule read [--date]` | Read schedule for a date |
| `schedule_write` | `schedule write --date --focus [--blocks]` | Write schedule for a date |

### 3. Skills

**Directory:** `.picoclaw/skills/`

Skills are markdown files with YAML frontmatter, loaded by picoclaw via the native `skill` tool.

| Skill | File | Purpose |
|---|---|---|
| `goal-tracking` | `goal-tracking.md` | How to evaluate goal progress, ask clarifying questions, and suggest milestones |
| `task-planning` | `task-planning.md` | How to decompose a goal into prioritized, time-estimated tasks |
| `daily-digest` | `daily-digest.md` | Format rules and content guidelines for the morning brief |

### 4. Configuration

**File:** `.picoclaw/config.json`

```jsonc
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "model": "claude-sonnet-4.6"
    }
  },
  "model_list": [
    {
      "model_name": "claude-sonnet-4.6",
      "model": "anthropic/claude-sonnet-4.6",
      "api_key": "your-anthropic-key"
    }
  ],
  "tools": {
    "exec": {
      "enabled": true,
      "allowed_commands": [".openlos"]
    }
  }
}
```

---

## Implementation Order

1. Rename `assets/opencode/` → `assets/picoclaw/`
2. Update `assets/assets.go` to embed from `picoclaw`
3. Convert TypeScript tools → picoclaw TOOLS.md format
4. Migrate skills to picoclaw markdown format (with frontmatter)
5. Create AGENTS.md for the assistant persona
6. Update Go CLI:
   - Rename `OPENLOS_WORKTREE` → `PICOCLAW_WORKSPACE`
   - Update default paths
7. Update install logic in main.go:
   - Change `.opencode` → `.picoclaw`
   - Copy Go binary to `.picoclaw/bin/openlos`

---

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Runtime | picoclaw | Ultra-lightweight ($10 hardware, <10MB RAM), fast startup, AI bootstrapped |
| Storage | SQLite via Go CLI | Reliable, ACID-compliant, no data corruption |
| Interface | picoclaw CLI/chat | Works with Telegram, Discord, WhatsApp, etc. |
| Tools | Exec-based | Reuse existing Go CLI; no rewrite needed |
| Skills vs system prompt | Both | System prompt sets persona; skills provide on-demand deep instructions |
| Data location | `data/` (gitignored) | Keeps personal data out of version control |

---

## Conventions

### Commit Messages

Format: `type: description (lowercase)`

Types:
- `feat`: New features
- `refactor`: Code refactoring (no behavioral change)
- `chore`: Maintenance, tooling, dependencies
- `agent`: Changes to agent definitions or prompts

Examples:
```
feat: add tasksStatusChecker and scheduleSuggester CLI tools
refactor: target picoclaw instead of opencode
chore: add opencode agent
```

---

## Non-Goals

- No web UI or API server (use picoclaw channels instead)
- No push notifications or background scheduling (use picoclaw heartbeat)
- No integration with external calendar or task management services

---

## Future Extensions

### Financial Spending Tracking

Track personal expenses with natural language input and generate spending insights.

#### Data Model

```sql
-- accounts table
CREATE TABLE accounts (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL, -- 'checking', 'savings', 'credit', 'cash'
  balance REAL DEFAULT 0,
  created INTEGER NOT NULL
);

-- categories table
CREATE TABLE categories (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  color TEXT, -- hex color for UI
  budget REAL DEFAULT 0
);

-- transactions table
CREATE TABLE transactions (
  id TEXT PRIMARY KEY,
  account_id TEXT REFERENCES accounts(id),
  category_id TEXT REFERENCES categories(id),
  amount REAL NOT NULL, -- negative for expenses, positive for income
  description TEXT,
  date INTEGER NOT NULL,
  created INTEGER NOT NULL
);
```

#### Tools

| Tool | CLI Command | Description |
|---|---|---|
| `account_add` | `account add --name --type [--balance]` | Add a new account |
| `account_list` | `account list` | List all accounts with balances |
| `category_add` | `category add --name [--budget] [--color]` | Add a spending category |
| `category_list` | `category list` | List categories with budgets |
| `transaction_add` | `transaction add --account --category --amount --description [--date]` | Log a transaction |
| `transaction_list` | `transaction list [--account] [--category] [--start-date] [--end-date]` | List transactions with filters |
| `spending_report` | `spending report [--period month\|year] [--category]` | Generate spending summary |

#### Skills

- **budget-management**: How to set category budgets, track variance, and suggest adjustments
- **spending-analysis**: How to analyze spending patterns and provide actionable insights

#### Example Interactions

- "I just spent $45 at Target on groceries" → Logs expense to checking account, category: groceries
- "How much did I spend on dining out this month?" → Queries transactions, returns total
- "Generate a spending report for January" → Shows breakdown by category vs budget

#### Implementation Notes

- Categories should have sensible defaults (Food, Transport, Entertainment, Utilities, etc.)
- Support recurring transactions (subscriptions, rent) via scheduled tasks
- Integrate with daily-digest to show spending highlights

---

### Periodic Task & Schedule Assistant

Leverage picoclaw's native heartbeat to periodically check task status and proactively suggest the daily schedule.

#### How It Works

1. Create `HEARTBEAT.md` in the workspace
2. Picoclaw reads it every 30 minutes (default interval)
3. Executes tasks using available tools
4. Outputs results to CLI

#### Configuration

```json
{
  "heartbeat": {
    "enabled": true,
    "interval": 30
  }
}
```

#### HEARTBEAT.md

```markdown
# Periodic Tasks

## Quick Tasks (respond directly)
- Run task_status_checker to check for overdue and due-today tasks
- If tasks are overdue, reschedule them to today and notify the user

## Long Tasks (use spawn for async)
- Run schedule_suggester to analyze tasks and goals, then suggest today's schedule if none exists
```

#### Tools

| Tool | Description |
|------|-------------|
| `task_status_checker` | Check for overdue and due-today tasks; auto-reschedule overdue tasks to today |
| `schedule_suggester` | Analyze tasks and goals; suggest today's schedule if none exists |

#### task_status_checker Logic

1. Query all open tasks via `tasks_list`
2. Identify overdue tasks (due date < today)
3. For each overdue task: update due date to today via `tasks_update`
4. Output summary: "Found X overdue task(s) (rescheduled to today), Y task(s) due today"

#### schedule_suggester Logic

1. Check if today's schedule exists via `schedule_read`
2. If exists: skip (output "Schedule already set for today")
3. If not: generate suggestion using daily-digest logic
4. Output suggestion to CLI, ask user to confirm

#### Data Flow

```
Picoclaw Heartbeat (every 30 min)
    │
    ├─→ task_status_checker
    │   ├─→ Query open tasks
    │   ├─→ Find overdue → reschedule to today
    │   └─→ CLI output: "2 overdue tasks moved to today"
    │
    └─→ schedule_suggester (spawned)
        ├─→ Check today's schedule
        ├─→ If none: generate suggestion
        └─→ CLI output: "No schedule set. Suggested focus: ..."
```

#### Files to Modify/Create

| File | Action |
|------|--------|
| `.picoclaw/HEARTBEAT.md` | Create |
| `.picoclaw/TOOLS.md` | Add `task_status_checker`, `schedule_suggester` tools |
| `main.go` | Add `tasks reschedule` command |
| `.picoclaw/skills/daily-digest.md` | Update to mention auto-reschedule |

---

### Other Extensions

- Add more chat channels (Telegram, Discord, WhatsApp) via picoclaw config
- Integrate with picoclaw's memory system for long-term context
- Export command to emit Markdown reports
