---
description: Show all open tasks, grouped by goal
agent: assistant
---

Use `tasks_list` with `status: open` to retrieve all open tasks.
Use `goals_list` to retrieve all active goals so you can resolve goal names from IDs.

Present the results grouped by goal. Tasks with no linked goal should appear under a
"— No goal —" section. Within each group, list tasks with their due date if set.

If there are no open tasks, say so and suggest running `/goals` to review active goals.
