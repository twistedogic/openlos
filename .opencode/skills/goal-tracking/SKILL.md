---
name: goal-tracking
description: Instructions for evaluating goal progress, asking clarifying questions, and suggesting milestones
license: MIT
compatibility: opencode
---

## What I do

- Evaluate the current status of a goal based on its linked tasks
- Surface goals that are stalled (active but no open tasks)
- Ask clarifying questions when a goal is vague or unmeasurable
- Suggest concrete milestones that break a goal into observable checkpoints
- Recommend when to move a goal from `active` → `paused` or `completed`

## How to evaluate progress

1. Load the goal using `goals_list`
2. Load all tasks linked to that goal using `tasks_list` with the `goal_id` filter
3. Compute:
   - Total tasks: count of all tasks for this goal
   - Completed tasks: count with `status: done`
   - Progress %: (completed / total) × 100, or 0 if no tasks exist
4. Report progress clearly: e.g. "3 of 7 tasks done (43%)"

## Staleness rules

A goal is **stalled** if:
- Status is `active`, AND
- It has zero open tasks (either no tasks at all, or all tasks are `done` but the goal isn't marked `completed`)

When a goal is stalled, prompt the user:
> "Goal '...' has no open tasks. Would you like to add new tasks, mark it completed, or pause it?"

## Clarifying questions to ask

Ask these when adding or reviewing a goal that seems vague:

- "How will you know when this goal is achieved?"
- "What is the target date or timeframe, if any?"
- "What is the first concrete step you could take this week?"

## Milestone suggestions

When asked to suggest milestones for a goal:
- Identify 3–5 observable checkpoints that represent meaningful progress
- Each milestone should be specific and verifiable (avoid "make progress on X")
- Express them as tasks with suggested due dates spaced evenly toward the goal horizon

## When to recommend status changes

| Situation | Recommended action |
|---|---|
| All tasks done, goal achieved | Mark goal `completed` via `goals_update` |
| No activity expected soon | Mark goal `paused` via `goals_update` |
| Goal scope has changed significantly | Update description via `goals_update`, then reassess tasks |
