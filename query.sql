-- name: CreateIdea :one
INSERT INTO ideas (id, text, created)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListIdeas :many
SELECT * FROM ideas
ORDER BY created DESC
LIMIT ?;

-- name: CreateGoal :one
INSERT INTO goals (id, title, description, status, created)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListGoals :many
SELECT * FROM goals
ORDER BY created DESC;

-- name: ListGoalsByStatus :many
SELECT * FROM goals
WHERE status = ?
ORDER BY created DESC;

-- name: GetGoal :one
SELECT * FROM goals
WHERE id = ?;

-- name: UpdateGoal :one
UPDATE goals
SET status = ?, description = ?
WHERE id = ?
RETURNING *;

-- name: CreateTask :one
INSERT INTO tasks (id, title, goal_id, status, created, due)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY created DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE status = ?
ORDER BY created DESC;

-- name: ListTasksByGoal :many
SELECT * FROM tasks
WHERE goal_id = ?
ORDER BY created DESC;

-- name: ListTasksByStatusAndGoal :many
SELECT * FROM tasks
WHERE status = ? AND goal_id = ?
ORDER BY created DESC;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = ?;

-- name: UpdateTask :one
UPDATE tasks
SET status = ?, due = ?
WHERE id = ?
RETURNING *;

-- name: GetSchedule :one
SELECT * FROM schedule
WHERE date = ?;

-- name: UpsertSchedule :one
INSERT INTO schedule (date, focus, blocks)
VALUES (?, ?, ?)
ON CONFLICT(date) DO UPDATE SET
    focus = excluded.focus,
    blocks = excluded.blocks
RETURNING *;

-- name: ListSchedules :many
SELECT * FROM schedule
ORDER BY date DESC;
