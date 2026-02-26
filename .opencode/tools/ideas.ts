import { tool } from "@opencode-ai/plugin"
import { execSync } from "child_process"

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
  description: "Log a new raw idea (backed by openlos).",
  args: {
    text: tool.schema.string().describe("The full text of the idea to capture"),
  },
  async execute(args, context) {
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} ideas log --text ${JSON.stringify(args.text)} --worktree ${JSON.stringify(context.worktree)}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})

export const list = tool({
  description: "List all captured ideas (backed by openlos).",
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
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} ideas list --limit ${args.limit ?? 50} --worktree ${JSON.stringify(context.worktree)}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})
