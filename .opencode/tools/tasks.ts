import { tool } from "@opencode-ai/plugin"
import { readFileSync, writeFileSync, existsSync } from "fs"
import { join } from "path"
import { randomUUID } from "crypto"

type TaskStatus = "open" | "done"

interface Task {
  id: string
  title: string
  goal_id: string | null
  status: TaskStatus
  created: string
  due: string | null
}

function dataPath(worktree: string): string {
  return join(worktree, "data", "tasks.json")
}

function load(worktree: string): Task[] {
  const p = dataPath(worktree)
  if (!existsSync(p)) return []
  return JSON.parse(readFileSync(p, "utf-8")) as Task[]
}

function save(worktree: string, tasks: Task[]): void {
  writeFileSync(dataPath(worktree), JSON.stringify(tasks, null, 2) + "\n", "utf-8")
}

export const add = tool({
  description:
    "Add a new task. Optionally link it to a goal by providing goal_id. Optionally set a due date (ISO 8601).",
  args: {
    title: tool.schema.string().describe("Short, actionable task title"),
    goal_id: tool.schema
      .string()
      .uuid()
      .optional()
      .describe("UUID of the goal this task belongs to, if any"),
    due: tool.schema
      .string()
      .optional()
      .describe("Due date in ISO 8601 format (e.g. 2026-03-01), optional"),
  },
  async execute(args, context) {
    const tasks = load(context.worktree)
    const task: Task = {
      id: randomUUID(),
      title: args.title,
      goal_id: args.goal_id ?? null,
      status: "open",
      created: new Date().toISOString(),
      due: args.due ?? null,
    }
    tasks.push(task)
    save(context.worktree, tasks)
    return `Task added (id: ${task.id}): "${task.title}"${task.due ? ` — due ${task.due}` : ""}`
  },
})

export const list = tool({
  description:
    "List tasks. Filter by status (open/done) or by goal_id. Returns all tasks if no filters given.",
  args: {
    status: tool.schema
      .enum(["open", "done"])
      .optional()
      .describe("Filter by task status"),
    goal_id: tool.schema
      .string()
      .uuid()
      .optional()
      .describe("Filter tasks belonging to a specific goal UUID"),
  },
  async execute(args, context) {
    let tasks = load(context.worktree)
    if (tasks.length === 0) return "No tasks found."
    if (args.status) tasks = tasks.filter((t) => t.status === args.status)
    if (args.goal_id) tasks = tasks.filter((t) => t.goal_id === args.goal_id)
    if (tasks.length === 0) return "No tasks match the given filters."
    return tasks
      .map((t) => {
        const due = t.due ? ` [due: ${t.due}]` : ""
        const goal = t.goal_id ? ` [goal: ${t.goal_id}]` : ""
        return `- [${t.status}] ${t.title}${due}${goal}  (id: ${t.id})`
      })
      .join("\n")
  },
})

export const update = tool({
  description: "Update a task's status or due date by its UUID.",
  args: {
    id: tool.schema.string().uuid().describe("UUID of the task to update"),
    status: tool.schema
      .enum(["open", "done"])
      .optional()
      .describe("New status for the task"),
    due: tool.schema
      .string()
      .optional()
      .describe("New due date in ISO 8601 format, or null to clear it"),
  },
  async execute(args, context) {
    const tasks = load(context.worktree)
    const idx = tasks.findIndex((t) => t.id === args.id)
    if (idx === -1) return `Task not found: ${args.id}`
    if (args.status !== undefined) tasks[idx].status = args.status
    if (args.due !== undefined) tasks[idx].due = args.due || null
    save(context.worktree, tasks)
    return `Task updated (id: ${args.id}): status=${tasks[idx].status}, due=${tasks[idx].due ?? "none"}`
  },
})
