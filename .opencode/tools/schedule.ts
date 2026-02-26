import { tool } from "@opencode-ai/plugin"
import { execSync } from "child_process"

interface TimeBlock {
  time: string
  activity: string
}

interface DaySchedule {
  focus: string
  blocks: TimeBlock[]
}

type Schedule = Record<string, DaySchedule>

function dataPath(worktree: string): string {
  return join(worktree, "data", "schedule.json")
}

function load(worktree: string): Schedule {
  const p = dataPath(worktree)
  if (!existsSync(p)) return {}
  return JSON.parse(readFileSync(p, "utf-8")) as Schedule
}

function save(worktree: string, schedule: Schedule): void {
  writeFileSync(dataPath(worktree), JSON.stringify(schedule, null, 2) + "\n", "utf-8")
}

function todayISO(): string {
  return new Date().toISOString().slice(0, 10)
}

export const read = tool({
  description: "Read the schedule for a given date (YYYY-MM-DD). (backed by openlos)",
  args: {
    date: tool.schema
      .string()
      .optional()
      .describe("Date in YYYY-MM-DD format. Defaults to today."),
  },
  async execute(args, context) {
    const bin = `${context.worktree}/.opencode/bin/openlos`
    const cmd = `${bin} schedule read --date ${JSON.stringify(args.date ?? "")} --worktree ${JSON.stringify(
      context.worktree,
    )}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})

export const write = tool({
  description: "Write or update the schedule for a given date (YYYY-MM-DD). (backed by openlos)",
  args: {
    date: tool.schema
      .string()
      .optional()
      .describe("Date in YYYY-MM-DD format. Defaults to today."),
    focus: tool.schema.string().describe("One-line theme or intention for the day"),
    blocks: tool.schema
      .array(
        tool.schema.object({
          time: tool.schema.string().describe("Time in HH:MM format (24h)"),
          activity: tool.schema.string().describe("Description of the activity"),
        }),
      )
      .optional()
      .describe("Ordered list of time blocks for the day"),
  },
  async execute(args, context) {
    const bin = `${context.worktree}/.opencode/bin/openlos`
    // convert blocks into comma-separated HH:MM|Activity strings
    const blocksArg = (args.blocks ?? []).map((b: any) => `${b.time}|${b.activity}`).join(",")
    const cmd = `${bin} schedule write --date ${JSON.stringify(args.date ?? "")} --focus ${JSON.stringify(
      args.focus,
    )} --blocks ${JSON.stringify(blocksArg)} --worktree ${JSON.stringify(context.worktree)}`
    return execSync(cmd, { encoding: "utf-8" }).toString()
  },
})
