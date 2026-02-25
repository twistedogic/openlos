---
name: task-planning
description: Instructions for decomposing goals into prioritized, time-estimated tasks
license: MIT
compatibility: opencode
---

## What I do

- Break a goal down into a concrete, ordered list of actionable tasks
- Assign priority order and suggest due dates
- Ensure every task is specific, independently completable, and unambiguous
- Avoid tasks that are too large (should take no more than 1–2 days of focused work)
- Link each created task to its parent goal via `goal_id`

## Decomposition process

1. Load the goal using `goals_list` to confirm it exists and read its description
2. Load any existing tasks for this goal using `tasks_list` with `goal_id` filter — avoid duplicating work already planned
3. Ask the user for any missing context if the goal is ambiguous (see clarifying questions below)
4. Draft the task list (internally) before writing — present it to the user first
5. After user confirms, persist each task using `tasks_add`, linking `goal_id`

## Task quality rules

A well-formed task must be:
- **Specific**: starts with an action verb (Write, Research, Set up, Review, Draft, Test, Send, etc.)
- **Scoped**: completable in a single focused work session (≤ 2 days)
- **Verifiable**: it's clear when the task is done
- **Atomic**: does not depend on another incomplete task within the same session

**Bad:** "Work on the report"
**Good:** "Draft the executive summary section of the Q1 report"

## Suggested ordering

Order tasks so that:
1. Tasks that unblock other tasks come first
2. Research/exploration tasks precede execution tasks
3. Review/polish tasks come last

## Due date heuristics

If the goal has no explicit deadline:
- Spread tasks across the next 2 weeks by default
- Assign 1–3 tasks per day, leaving buffer days
- Ask the user if they have a preferred completion date before assigning dates

If a deadline is known:
- Work backward from the deadline
- Flag if the number of tasks makes the deadline unrealistic

## Clarifying questions

Ask before decomposing if the goal description is thin:
- "What does 'done' look like for this goal?"
- "Are there any dependencies or blockers you know about?"
- "How much time per day can you realistically dedicate to this?"
- "Is there a hard deadline, or is this open-ended?"
