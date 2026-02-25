import { tool } from "@opencode-ai/plugin"
import { readFileSync, writeFileSync, existsSync } from "fs"
import { join } from "path"
import { randomUUID } from "crypto"

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
  description: "Add a new goal with a title and optional description. Status defaults to 'active'.",
  args: {
    title: tool.schema.string().describe("Short goal title"),
    description: tool.schema
      .string()
      .optional()
      .describe("Longer description of what success looks like for this goal"),
  },
  async execute(args, context) {
    const goals = load(context.worktree)
    const goal: Goal = {
      id: randomUUID(),
      title: args.title,
      description: args.description ?? "",
      status: "active",
      created: new Date().toISOString(),
    }
    goals.push(goal)
    save(context.worktree, goals)
    return `Goal added (id: ${goal.id}): "${goal.title}"`
  },
})

export const list = tool({
  description:
    "List goals. Optionally filter by status (active/completed/paused). Returns all goals if no filter given.",
  args: {
    status: tool.schema
      .enum(["active", "completed", "paused"])
      .optional()
      .describe("Filter by goal status"),
  },
  async execute(args, context) {
    let goals = load(context.worktree)
    if (goals.length === 0) return "No goals found."
    if (args.status) goals = goals.filter((g) => g.status === args.status)
    if (goals.length === 0) return "No goals match the given filter."
    return goals
      .map((g) => {
        const desc = g.description ? `\n    ${g.description}` : ""
        return `- [${g.status}] ${g.title}  (id: ${g.id})${desc}`
      })
      .join("\n")
  },
})

export const update = tool({
  description: "Update a goal's status or description by its UUID.",
  args: {
    id: tool.schema.string().uuid().describe("UUID of the goal to update"),
    status: tool.schema
      .enum(["active", "completed", "paused"])
      .optional()
      .describe("New status for the goal"),
    description: tool.schema.string().optional().describe("Updated description for the goal"),
  },
  async execute(args, context) {
    const goals = load(context.worktree)
    const idx = goals.findIndex((g) => g.id === args.id)
    if (idx === -1) return `Goal not found: ${args.id}`
    if (args.status !== undefined) goals[idx].status = args.status
    if (args.description !== undefined) goals[idx].description = args.description
    save(context.worktree, goals)
    return `Goal updated (id: ${args.id}): status=${goals[idx].status}`
  },
})
