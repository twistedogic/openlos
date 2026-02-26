import { tool } from "@opencode-ai/plugin"
import { execSync } from "child_process"

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
    "Add a new task. Optionally link it to a goal by providing goal_id. Optionally set a due date (ISO 8601). (backed by openlos)",
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
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} tasks add --title ${JSON.stringify(args.title)} --goal-id ${JSON.stringify(
      args.goal_id ?? "",
    )} --due ${JSON.stringify(args.due ?? "")} --worktree ${JSON.stringify(context.worktree)}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})

export const list = tool({
  description:
    "List tasks. Filter by status (open/done) or by goal_id. Returns all tasks if no filters given. (backed by openlos)",
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
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} tasks list --status ${JSON.stringify(args.status ?? "")} --goal-id ${JSON.stringify(
      args.goal_id ?? "",
    )} --worktree ${JSON.stringify(context.worktree)}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})

export const update = tool({
  description: "Update a task's status or due date by its UUID. (backed by openlos)",
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
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} tasks update --id ${JSON.stringify(args.id)} --status ${JSON.stringify(
      args.status ?? "",
    )} --due ${JSON.stringify(args.due ?? "")} --worktree ${JSON.stringify(context.worktree)}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})
