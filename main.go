package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

func worktreeFromEnvOrFlag(wFlag string) string {
	if wFlag != "" && wFlag != "." {
		return wFlag
	}
	if w := os.Getenv("OPENLOS_WORKTREE"); w != "" {
		return w
	}
	return "."
}

func dataPath(worktree, file string) string {
	return filepath.Join(worktree, "data", file)
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// treat missing file as empty result
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeJSON(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

// models
type Idea struct {
	ID      string    `json:"id"`
	Text    string    `json:"text"`
	Created time.Time `json:"created"`
}

type Task struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	GoalID  *string   `json:"goal_id"`
	Status  string    `json:"status"`
	Created time.Time `json:"created"`
	Due     *string   `json:"due"`
}

type Goal struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Created     time.Time `json:"created"`
}

type TimeBlock struct {
	Time     string `json:"time"`
	Activity string `json:"activity"`
}

type DaySchedule struct {
	Date   time.Time   `json:"date"`
	Focus  string      `json:"focus"`
	Blocks []TimeBlock `json:"blocks"`
}

type Schedule map[string]DaySchedule

// ideas
func ideasLog(worktree, text string) error {
	var arr []Idea
	p := dataPath(worktree, "ideas.json")
	if err := readJSON(p, &arr); err != nil {
		return err
	}
	entry := Idea{ID: uuid.NewString(), Text: text, Created: time.Now()}
	arr = append(arr, entry)
	if err := writeJSON(p, arr); err != nil {
		return err
	}
	fmt.Printf("Idea captured (id: %s): %q\n", entry.ID, entry.Text)
	return nil
}

func ideasList(worktree string, limit int) error {
	var arr []Idea
	p := dataPath(worktree, "ideas.json")
	if err := readJSON(p, &arr); err != nil {
		return err
	}
	if len(arr) == 0 {
		fmt.Println("No ideas captured yet.")
		return nil
	}
	// most recent first
	slices.SortFunc(arr, func(a, b Idea) int { return b.Created.Compare(a.Created) })
	if limit <= 0 || limit > len(arr) {
		limit = len(arr)
	}
	for i := 0; i < limit; i++ {
		it := arr[i]
		fmt.Printf("%d. [%s] %s  (id: %s)\n", i+1, it.Created, it.Text, it.ID)
	}
	return nil
}

// goals
func goalsAdd(worktree, title, description string) error {
	var arr []Goal
	p := dataPath(worktree, "goals.json")
	if err := readJSON(p, &arr); err != nil {
		return err
	}
	g := Goal{ID: uuid.NewString(), Title: title, Description: description, Status: "active", Created: time.Now()}
	arr = append(arr, g)
	if err := writeJSON(p, arr); err != nil {
		return err
	}
	fmt.Printf("Goal added (id: %s): \"%s\"\n", g.ID, g.Title)
	return nil
}

func goalsList(worktree, status string) error {
	var arr []Goal
	p := dataPath(worktree, "goals.json")
	if err := readJSON(p, &arr); err != nil {
		return err
	}
	if len(arr) == 0 {
		fmt.Println("No goals found.")
		return nil
	}

	var out []Goal
	if status == "" {
		out = arr
	} else {
		for _, g := range arr {
			if g.Status == status {
				out = append(out, g)
			}
		}
	}
	if len(out) == 0 {
		fmt.Println("No goals match the given filter.")
		return nil
	}
	for _, g := range out {
		desc := ""
		if strings.TrimSpace(g.Description) != "" {
			desc = "\n    " + g.Description
		}
		fmt.Printf("- [%s] %s  (id: %s)%s\n", g.Status, g.Title, g.ID, desc)
	}
	return nil
}

func goalsUpdate(worktree, id, status, description string) error {
	var arr []Goal
	p := dataPath(worktree, "goals.json")
	if err := readJSON(p, &arr); err != nil {
		return err
	}
	idx := -1
	for i, g := range arr {
		if g.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("Goal not found: " + id)
	}
	if status != "" {
		arr[idx].Status = status
	}
	if description != "" {
		arr[idx].Description = description
	}
	if err := writeJSON(p, arr); err != nil {
		return err
	}
	fmt.Printf("Goal updated (id: %s): status=%s\n", id, arr[idx].Status)
	return nil
}

// tasks
func tasksAdd(worktree, title, goalID, due string) error {
	var arr []Task
	p := dataPath(worktree, "tasks.json")
	if err := readJSON(p, &arr); err != nil {
		return err
	}
	var goalPtr *string
	if strings.TrimSpace(goalID) != "" {
		g := goalID
		goalPtr = &g
	}
	var duePtr *string
	if strings.TrimSpace(due) != "" {
		d := due
		duePtr = &d
	}
	t := Task{ID: uuid.NewString(), Title: title, GoalID: goalPtr, Status: "open", Created: time.Now(), Due: duePtr}
	arr = append(arr, t)
	if err := writeJSON(p, arr); err != nil {
		return err
	}
	dueMsg := ""
	if t.Due != nil {
		dueMsg = fmt.Sprintf(" — due %s", *t.Due)
	}
	fmt.Printf("Task added (id: %s): \"%s\"%s\n", t.ID, t.Title, dueMsg)
	return nil
}

func tasksList(worktree, status, goalID string) error {
	var arr []Task
	p := dataPath(worktree, "tasks.json")
	if err := readJSON(p, &arr); err != nil {
		return err
	}
	if len(arr) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}
	var out []Task
	for _, t := range arr {
		if status != "" && t.Status != status {
			continue
		}
		if goalID != "" {
			if t.GoalID == nil || *t.GoalID != goalID {
				continue
			}
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		fmt.Println("No tasks match the given filters.")
		return nil
	}
	for _, t := range out {
		due := ""
		if t.Due != nil {
			due = fmt.Sprintf(" [due: %s]", *t.Due)
		}
		goal := ""
		if t.GoalID != nil {
			goal = fmt.Sprintf(" [goal: %s]", *t.GoalID)
		}
		fmt.Printf("- [%s] %s%s%s  (id: %s)\n", t.Status, t.Title, due, goal, t.ID)
	}
	return nil
}

func tasksUpdate(worktree, id, status, due string) error {
	var arr []Task
	p := dataPath(worktree, "tasks.json")
	if err := readJSON(p, &arr); err != nil {
		return err
	}
	idx := -1
	for i, t := range arr {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("Task not found: " + id)
	}
	if status != "" {
		arr[idx].Status = status
	}
	if due != "" {
		if due == "null" {
			arr[idx].Due = nil
		} else {
			d := due
			arr[idx].Due = &d
		}
	}
	if err := writeJSON(p, arr); err != nil {
		return err
	}
	dueVal := "none"
	if arr[idx].Due != nil {
		dueVal = *arr[idx].Due
	}
	fmt.Printf("Task updated (id: %s): status=%s, due=%s\n", id, arr[idx].Status, dueVal)
	return nil
}

// schedule
func scheduleRead(worktree, date string) error {
	if date == "" {
		date = time.Now().UTC().Format(time.DateOnly)
	}
	var s Schedule
	p := dataPath(worktree, "schedule.json")
	if err := readJSON(p, &s); err != nil {
		return err
	}
	day, ok := s[date]
	if !ok {
		fmt.Printf("No schedule found for %s.\n", date)
		return nil
	}
	blocks := "  (no time blocks set)"
	if len(day.Blocks) > 0 {
		var sb strings.Builder
		for _, b := range day.Blocks {
			fmt.Fprintf(&sb, "  %s  %s\n", b.Time, b.Activity)
		}
		blocks = strings.TrimRight(sb.String(), "\n")
	}
	fmt.Printf("Schedule for %s\nFocus: %s\n\n%s\n", date, day.Focus, blocks)
	return nil
}

func scheduleWrite(worktree, date, focus string, blocks []string) error {
	if date == "" {
		date = time.Now().UTC().Format(time.DateOnly)
	}
	var s Schedule
	p := dataPath(worktree, "schedule.json")
	if err := readJSON(p, &s); err != nil {
		return err
	}
	if s == nil {
		s = make(Schedule)
	}
	tb := []TimeBlock{}
	for _, b := range blocks {
		// expect format HH:MM|Activity
		parts := strings.SplitN(b, "|", 2)
		if len(parts) != 2 {
			return errors.New("invalid block format, expected HH:MM|Activity")
		}
		tb = append(tb, TimeBlock{Time: parts[0], Activity: parts[1]})
	}
	s[date] = DaySchedule{Focus: focus, Blocks: tb}
	if err := writeJSON(p, s); err != nil {
		return err
	}
	fmt.Printf("Schedule saved for %s: \"%s\" with %d time block(s).\n", date, focus, len(tb))
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: openlos <ideas|tasks|goals|schedule> <subcommand> [flags]")
		os.Exit(2)
	}
	cmd := os.Args[1]
	switch cmd {
	case "ideas":
		if len(os.Args) < 3 {
			fmt.Println("usage: openlos ideas <log|list> [flags]")
			os.Exit(2)
		}
		sub := os.Args[2]
		switch sub {
		case "log":
			fs := flag.NewFlagSet("ideas log", flag.ExitOnError)
			text := fs.String("text", "", "idea text")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if *text == "" {
				fmt.Fprintln(os.Stderr, "missing --text")
				os.Exit(2)
			}
			if err := ideasLog(worktree, *text); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		case "list":
			fs := flag.NewFlagSet("ideas list", flag.ExitOnError)
			limit := fs.Int("limit", 50, "max items")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if err := ideasList(worktree, *limit); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		default:
			fmt.Println("unknown ideas subcommand:", sub)
			os.Exit(2)
		}
	case "goals":
		if len(os.Args) < 3 {
			fmt.Println("usage: openlos goals <add|list|update> [flags]")
			os.Exit(2)
		}
		sub := os.Args[2]
		switch sub {
		case "add":
			fs := flag.NewFlagSet("goals add", flag.ExitOnError)
			title := fs.String("title", "", "goal title")
			desc := fs.String("description", "", "goal description")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if *title == "" {
				fmt.Fprintln(os.Stderr, "missing --title")
				os.Exit(2)
			}
			if err := goalsAdd(worktree, *title, *desc); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		case "list":
			fs := flag.NewFlagSet("goals list", flag.ExitOnError)
			status := fs.String("status", "", "filter by status")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if err := goalsList(worktree, *status); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		case "update":
			fs := flag.NewFlagSet("goals update", flag.ExitOnError)
			id := fs.String("id", "", "goal id")
			status := fs.String("status", "", "new status")
			desc := fs.String("description", "", "new description")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if *id == "" {
				fmt.Fprintln(os.Stderr, "missing --id")
				os.Exit(2)
			}
			if err := goalsUpdate(worktree, *id, *status, *desc); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		default:
			fmt.Println("unknown goals subcommand:", sub)
			os.Exit(2)
		}
	case "tasks":
		if len(os.Args) < 3 {
			fmt.Println("usage: openlos tasks <add|list|update> [flags]")
			os.Exit(2)
		}
		sub := os.Args[2]
		switch sub {
		case "add":
			fs := flag.NewFlagSet("tasks add", flag.ExitOnError)
			title := fs.String("title", "", "task title")
			goal := fs.String("goal-id", "", "goal id")
			due := fs.String("due", "", "due date")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if *title == "" {
				fmt.Fprintln(os.Stderr, "missing --title")
				os.Exit(2)
			}
			if err := tasksAdd(worktree, *title, *goal, *due); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		case "list":
			fs := flag.NewFlagSet("tasks list", flag.ExitOnError)
			status := fs.String("status", "", "filter by status")
			goal := fs.String("goal-id", "", "goal id")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if err := tasksList(worktree, *status, *goal); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		case "update":
			fs := flag.NewFlagSet("tasks update", flag.ExitOnError)
			id := fs.String("id", "", "task id")
			status := fs.String("status", "", "new status")
			due := fs.String("due", "", "new due date or 'null' to clear")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if *id == "" {
				fmt.Fprintln(os.Stderr, "missing --id")
				os.Exit(2)
			}
			if err := tasksUpdate(worktree, *id, *status, *due); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		default:
			fmt.Println("unknown tasks subcommand:", sub)
			os.Exit(2)
		}
	case "schedule":
		if len(os.Args) < 3 {
			fmt.Println("usage: openlos schedule <read|write> [flags]")
			os.Exit(2)
		}
		sub := os.Args[2]
		switch sub {
		case "read":
			fs := flag.NewFlagSet("schedule read", flag.ExitOnError)
			date := fs.String("date", "", "date YYYY-MM-DD")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if err := scheduleRead(worktree, *date); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		case "write":
			fs := flag.NewFlagSet("schedule write", flag.ExitOnError)
			date := fs.String("date", "", "date YYYY-MM-DD")
			focus := fs.String("focus", "", "focus theme")
			blocks := fs.String("blocks", "", "comma-separated HH:MM|Activity blocks")
			w := fs.String("worktree", ".", "worktree")
			fs.Parse(os.Args[3:])
			worktree := worktreeFromEnvOrFlag(*w)
			if *focus == "" {
				fmt.Fprintln(os.Stderr, "missing --focus")
				os.Exit(2)
			}
			blks := []string{}
			if *blocks != "" {
				for s := range strings.SplitSeq(*blocks, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						blks = append(blks, s)
					}
				}
			}
			if err := scheduleWrite(worktree, *date, *focus, blks); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		default:
			fmt.Println("unknown schedule subcommand:", sub)
			os.Exit(2)
		}
	default:
		fmt.Println("unknown command:", cmd)
		os.Exit(2)
	}
}
