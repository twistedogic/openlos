# openlos — Design Document

> Virtual assistant workflow app built on opencode

---

## Overview

`openlos` is a **pure opencode configuration layer** — there is no separate CLI binary or build step. It lives entirely inside the `.opencode/` directory and activates whenever `opencode` is run from this repo. Users interact through the opencode TUI using custom `/` commands. The AI understands the productivity domain through a specialized agent, reusable skills, and custom TypeScript tools that read and write flat JSON data files.

---

## Goals

- Log and organize ideas, tasks, and goals via natural language
- Track goal progress and task status persistently
- Generate a daily digest (morning brief) on demand
- Break down goals into prioritized, actionable tasks
- Require no installation steps beyond `opencode` itself

---

## Architecture

```
openlos/
├── data/                          # gitignored — persistent user data
│   ├── ideas.json
│   ├── tasks.json
│   ├── goals.json
│   └── schedule.json
├── .opencode/
│   ├── opencode.json              # config: agent wiring, permissions, commands
│   ├── agents/
│   │   └── assistant.md           # "assistant" primary agent
│   ├── skills/
│   │   ├── goal-tracking/
│   │   │   └── SKILL.md
│   │   ├── task-planning/
│   │   │   └── SKILL.md
│   │   └── daily-digest/
│   │       └── SKILL.md
│   ├── tools/
│   │   ├── ideas.ts               # ideas_log, ideas_list
│   │   ├── tasks.ts               # tasks_add, tasks_list, tasks_update
│   │   ├── goals.ts               # goals_add, goals_list, goals_update
│   │   └── schedule.ts            # schedule_read, schedule_write
│   └── commands/
│       ├── log.md                 # /log <idea text>
│       ├── tasks.md               # /tasks
│       ├── goals.md               # /goals
│       ├── plan.md                # /plan <goal>
│       └── digest.md              # /digest
└── .gitignore                     # data/ already ignored
```

---

## Data Layer

All user data is stored as flat JSON files in `data/`. This directory is gitignored. Tools locate the `data/` directory at runtime using `context.worktree` from the opencode tool context.

### Schemas

```jsonc
// data/ideas.json — array of raw captured ideas
[
  {
    "id": "uuid-v4",
    "text": "...",
    "created": "ISO 8601 timestamp"
  }
]

// data/tasks.json — array of tasks, optionally linked to a goal
[
  {
    "id": "uuid-v4",
    "title": "...",
    "goal_id": "uuid-v4 | null",
    "status": "open | done",
    "created": "ISO 8601 timestamp",
    "due": "ISO 8601 timestamp | null"
  }
]

// data/goals.json — array of goals with lifecycle status
[
  {
    "id": "uuid-v4",
    "title": "...",
    "description": "...",
    "status": "active | completed | paused",
    "created": "ISO 8601 timestamp"
  }
]

// data/schedule.json — keyed by ISO date (YYYY-MM-DD)
{
  "2026-02-24": {
    "focus": "one-line theme for the day",
    "blocks": [
      { "time": "09:00", "activity": "..." }
    ]
  }
}
```

---

## Components

### 1. Custom Agent — `assistant`

**File:** `.opencode/agents/assistant.md`  
**Mode:** `primary` (available alongside the default `build` agent, switchable via Tab)

The `assistant` agent is a primary agent pre-prompted to behave as a personal productivity assistant. It:

- Understands the flat-file storage schema and tool surface
- Proactively invokes custom tools when the user makes natural-language requests
- Loads the appropriate skill (`goal-tracking`, `task-planning`, `daily-digest`) when performing domain-specific reasoning
- Avoids making code changes; it is focused entirely on productivity workflows

### 2. Custom Tools

Tools are TypeScript files in `.opencode/tools/`. Each file exports multiple named functions, following the opencode `<filename>_<export>` naming convention. opencode executes these via Bun internally — no `package.json` or install step is required.

| File | Exported tools | Description |
|---|---|---|
| `ideas.ts` | `ideas_log`, `ideas_list` | Append a new idea to `data/ideas.json`; list all captured ideas |
| `tasks.ts` | `tasks_add`, `tasks_list`, `tasks_update` | Add a task; list tasks filtered by status/goal; update status or due date |
| `goals.ts` | `goals_add`, `goals_list`, `goals_update` | Add a goal; list goals filtered by status; update status or description |
| `schedule.ts` | `schedule_read`, `schedule_write` | Read schedule for a given date; write/update a day's focus and time blocks |

All tools use `context.worktree` to resolve the absolute path to `data/` at runtime, ensuring they work regardless of where `opencode` is invoked from within the repo.

### 3. Skills

Skills are `SKILL.md` files discovered by opencode from `.opencode/skills/`. The agent loads them on demand via the native `skill` tool. Three skills are defined:

| Skill | Directory | Purpose |
|---|---|---|
| `goal-tracking` | `.opencode/skills/goal-tracking/` | How to evaluate goal progress, ask clarifying questions, and suggest milestones |
| `task-planning` | `.opencode/skills/task-planning/` | How to decompose a goal into prioritized, time-estimated tasks |
| `daily-digest` | `.opencode/skills/daily-digest/` | Format rules and content guidelines for the morning brief |

### 4. Custom Commands

Five slash commands defined as markdown files in `.opencode/commands/`. Each becomes available in the opencode TUI as `/command-name`.

| Command | Template trigger | What it does |
|---|---|---|
| `/log $ARGUMENTS` | User captures a raw idea | Calls `ideas_log` to persist the text; confirms the capture back to the user |
| `/tasks` | User wants to review work | Lists open tasks grouped by goal using `tasks_list` |
| `/goals` | User wants to review objectives | Shows all active goals with their task count and status via `goals_list` |
| `/plan $ARGUMENTS` | User wants to act on a goal | Loads `task-planning` skill; decomposes the named goal into tasks; persists them via `tasks_add` |
| `/digest` | User starts their day | Loads `daily-digest` skill; reads current tasks and goals; generates a focused morning brief |

### 5. Configuration — `opencode.json`

**File:** `.opencode/opencode.json`

Responsibilities:
- Declares the `assistant` agent alongside built-in agents
- Sets `allow` permission on all custom tool calls (no user confirmation prompts for data reads/writes)
- Skill access: `allow` for all three project skills

---

## Implementation Order

1. Seed data files (`data/*.json`)
2. Custom tools (`.opencode/tools/*.ts`)
3. Skills (`.opencode/skills/*/SKILL.md`)
4. Agent definition (`.opencode/agents/assistant.md`)
5. Command definitions (`.opencode/commands/*.md`)
6. Configuration (`.opencode/opencode.json`)

---

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Runtime | opencode + Bun (internal) | No separate install step; opencode resolves `@opencode-ai/plugin` via its bundled Bun runtime |
| Storage | Flat JSON files | Simple, human-readable, no database dependency, easy to inspect and back up |
| Interface | opencode TUI | User is already in opencode; custom commands and a specialized agent provide a natural UX |
| Agent mode | Additional primary (alongside `build`) | Preserves standard coding workflows; user switches to `assistant` for productivity tasks |
| Skills vs system prompt | Both | System prompt sets persistent persona; skills provide on-demand deep instructions without polluting every context window |
| Data location | `data/` (gitignored) | Keeps personal data out of version control while keeping the configuration shareable |

---

## Non-Goals

- No web UI or API server
- No push notifications or background scheduling
- No integration with external calendar or task management services (e.g. Google Calendar, Todoist)
- No separate CLI binary outside of opencode

---

## Future Extensions

- **MCP server** wrapping the data tools for use from other opencode projects
- **Recurring task generation** via a `/weekly-review` command
- **Goal retrospective** command that summarizes completed vs missed tasks per goal
- **Export command** to emit a Markdown report of all goals and tasks
