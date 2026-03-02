---
name: daily-digest
description: Format rules and content guidelines for generating a focused morning brief
---

## What I do

- Generate a concise, actionable morning brief for the current day
- Surface the most important open tasks based on due dates and goal priority
- Highlight any overdue tasks or stalled goals that need attention
- Suggest a single "focus theme" for the day
- Optionally write the digest to the schedule for today via `schedule_write`

## Data to gather before generating

Run these tool calls first:
1. `goals_list` with `status: active` — get all active goals
2. `tasks_list` with `status: open` — get all open tasks
3. `schedule_read` with today's date — check if a schedule already exists

## Digest structure

Output the digest in this exact order:

```
## Morning Brief — <YYYY-MM-DD>

### Focus for Today
<One sentence describing the recommended theme or intention for the day>

### Overdue Tasks
<List any tasks whose due date is before today — or "None" if clear>

### Due Today
<List tasks due today — or "None">

### Recommended Tasks (Top 3–5)
<Prioritized list of open tasks to tackle today, with their linked goal name>

### Active Goals — Quick Status
<For each active goal: title, progress (X/Y tasks done), and a one-line status note>

### Captured Ideas (Recent)
<Optional: show 2–3 most recent ideas if any were logged in the last 3 days>
```

## Prioritization logic for "Recommended Tasks"

Rank open tasks by:
1. Overdue (past due date) — highest priority
2. Due today
3. Due within the next 3 days
4. Tasks linked to goals with the most incomplete tasks (most momentum needed)
5. Tasks with no due date — lowest priority

Limit to 5 tasks maximum. The user should leave the brief feeling focused, not overwhelmed.

## Focus theme heuristics

Choose the focus theme based on:
- The goal with the most urgent or overdue tasks
- If everything is on track: the goal the user has been making the least progress on
- If no tasks exist: suggest a theme around capturing ideas and setting goals

Keep the focus theme to one sentence, starting with a verb or framing word:
- "Today: make headway on [goal]."
- "Focus: unblock [task] to move [goal] forward."
- "Priority: clear overdue items before taking on new work."

## Tone guidelines

- Be direct and brief — no filler phrases like "Great news!" or "Let's dive in!"
- Use plain language; avoid jargon
- If everything looks on track, say so concisely
- If there are problems (overdue tasks, stalled goals), name them plainly without alarm

## After generating

Ask the user:
> "Shall I save this as today's schedule focus?"

If yes, call `schedule_write` with `focus` set to the focus theme and `blocks` left empty (or populated if the user provides time blocks).
