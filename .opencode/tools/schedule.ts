import { tool } from "@opencode-ai/plugin"
import { readFileSync, writeFileSync, existsSync } from "fs"
import { join } from "path"

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
  description:
    "Read the schedule for a given date (YYYY-MM-DD). Defaults to today if no date is provided.",
  args: {
    date: tool.schema
      .string()
      .optional()
      .describe("Date in YYYY-MM-DD format. Defaults to today."),
  },
  async execute(args, context) {
    const date = args.date ?? todayISO()
    const schedule = load(context.worktree)
    const day = schedule[date]
    if (!day) return `No schedule found for ${date}.`
    const blocks =
      day.blocks.length > 0
        ? day.blocks.map((b) => `  ${b.time}  ${b.activity}`).join("\n")
        : "  (no time blocks set)"
    return `Schedule for ${date}\nFocus: ${day.focus}\n\n${blocks}`
  },
})

export const write = tool({
  description:
    "Write or update the schedule for a given date (YYYY-MM-DD). Provide a focus theme and optional time blocks.",
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
    const date = args.date ?? todayISO()
    const schedule = load(context.worktree)
    schedule[date] = {
      focus: args.focus,
      blocks: args.blocks ?? [],
    }
    save(context.worktree, schedule)
    const count = (args.blocks ?? []).length
    return `Schedule saved for ${date}: "${args.focus}" with ${count} time block(s).`
  },
})
