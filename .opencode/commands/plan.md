---
description: Break a goal into actionable tasks using the task-planning skill
agent: assistant
---

Load the `task-planning` skill.

The user wants to plan tasks for the following goal: $ARGUMENTS

Follow the decomposition process from the task-planning skill:
1. Check if a goal with a matching title already exists via `goals_list`; if not, create it with `goals_add` first
2. Load any existing tasks for this goal via `tasks_list` to avoid duplicates
3. Ask any necessary clarifying questions before proceeding (deadline, time per day, etc.)
4. Present the proposed task list to the user for confirmation
5. Once confirmed, persist each task via `tasks_add` linked to the goal's ID
6. Summarize what was created
