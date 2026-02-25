---
description: Show all active goals with task progress
agent: assistant
---

Load the `goal-tracking` skill.

Use `goals_list` with `status: active` to retrieve all active goals.
For each goal, use `tasks_list` with the goal's `goal_id` to count total and completed tasks.

Present a summary table:

  Goal title | Status | Progress (done/total) | Notes

Under the table, flag any stalled goals (active, zero open tasks) and suggest next actions
per the goal-tracking skill instructions.

If there are no active goals, say so and suggest running `/plan <goal name>` to get started.
