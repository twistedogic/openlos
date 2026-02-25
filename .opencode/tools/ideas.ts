import { tool } from "@opencode-ai/plugin"
import { readFileSync, writeFileSync, existsSync } from "fs"
import { join } from "path"
import { randomUUID } from "crypto"

interface Idea {
  id: string
  text: string
  created: string
}

function dataPath(worktree: string): string {
  return join(worktree, "data", "ideas.json")
}

function load(worktree: string): Idea[] {
  const p = dataPath(worktree)
  if (!existsSync(p)) return []
  return JSON.parse(readFileSync(p, "utf-8")) as Idea[]
}

function save(worktree: string, ideas: Idea[]): void {
  writeFileSync(dataPath(worktree), JSON.stringify(ideas, null, 2) + "\n", "utf-8")
}

export const log = tool({
  description: "Log a new raw idea. Call this whenever the user wants to capture a thought, observation, or idea.",
  args: {
    text: tool.schema.string().describe("The full text of the idea to capture"),
  },
  async execute(args, context) {
    const ideas = load(context.worktree)
    const entry: Idea = {
      id: randomUUID(),
      text: args.text,
      created: new Date().toISOString(),
    }
    ideas.push(entry)
    save(context.worktree, ideas)
    return `Idea captured (id: ${entry.id}): "${entry.text}"`
  },
})

export const list = tool({
  description: "List all captured ideas, most recent first.",
  args: {
    limit: tool.schema
      .number()
      .int()
      .min(1)
      .max(200)
      .optional()
      .describe("Maximum number of ideas to return (default: 50)"),
  },
  async execute(args, context) {
    const ideas = load(context.worktree)
    if (ideas.length === 0) return "No ideas captured yet."
    const sorted = [...ideas].sort(
      (a, b) => new Date(b.created).getTime() - new Date(a.created).getTime(),
    )
    const limited = sorted.slice(0, args.limit ?? 50)
    return limited
      .map((i, idx) => `${idx + 1}. [${i.created.slice(0, 10)}] ${i.text}  (id: ${i.id})`)
      .join("\n")
  },
})
