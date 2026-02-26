import { tool } from "@opencode-ai/plugin"
import { execSync } from "child_process"

type GoalStatus = "active" | "completed" | "paused"

interface Goal {
  id: string
  title: string
  description: string
  status: GoalStatus
  created: string
}

function dataPath(worktree: string): string {
  return join(worktree, "data", "goals.json")
}

function load(worktree: string): Goal[] {
  const p = dataPath(worktree)
  if (!existsSync(p)) return []
  return JSON.parse(readFileSync(p, "utf-8")) as Goal[]
}

function save(worktree: string, goals: Goal[]): void {
  writeFileSync(dataPath(worktree), JSON.stringify(goals, null, 2) + "\n", "utf-8")
}

export const add = tool({
  description: "Add a new goal (backed by openlos).",
  args: {
    title: tool.schema.string().describe("Short goal title"),
    description: tool.schema
      .string()
      .optional()
      .describe("Longer description of what success looks like for this goal"),
  },
  async execute(args, context) {
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} goals add --title ${JSON.stringify(args.title)} --description ${JSON.stringify(
      args.description ?? "",
    )} --worktree ${JSON.stringify(context.worktree)}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})

export const list = tool({
  description:
    "List goals (backed by openlos). Optionally filter by status (active/completed/paused).",
  args: {
    status: tool.schema
      .enum(["active", "completed", "paused"])
      .optional()
      .describe("Filter by goal status"),
  },
  async execute(args, context) {
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} goals list --status ${JSON.stringify(args.status ?? "")} --worktree ${JSON.stringify(
      context.worktree,
    )}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})

export const update = tool({
  description: "Update a goal (backed by openlos).",
  args: {
    id: tool.schema.string().uuid().describe("UUID of the goal to update"),
    status: tool.schema
      .enum(["active", "completed", "paused"])
      .optional()
      .describe("New status for the goal"),
    description: tool.schema.string().optional().describe("Updated description for the goal"),
  },
  async execute(args, context) {
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} goals update --id ${JSON.stringify(args.id)} --status ${JSON.stringify(
      args.status ?? "",
    )} --description ${JSON.stringify(args.description ?? "")} --worktree ${JSON.stringify(
      context.worktree,
    )}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})
