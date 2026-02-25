---
description: Generate today's morning brief — focus theme, due tasks, and goal status
agent: assistant
---

Load the `daily-digest` skill.

Generate a morning brief for today following the skill's structure and guidelines exactly:
1. Gather data: `goals_list` (active), `tasks_list` (open), `schedule_read` (today)
2. Identify overdue tasks, tasks due today, and the top 3–5 recommended tasks
3. Compute per-goal progress
4. Choose a focus theme
5. Output the digest in the format specified by the skill

After presenting the digest, ask the user if they want to save the focus theme to today's
schedule via `schedule_write`.
