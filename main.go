package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/twistedogic/openlos/assets"
	"github.com/twistedogic/openlos/db"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

func worktreeFromEnvOrFlag(wFlag string) string {
	if wFlag != "" && wFlag != "." {
		return wFlag
	}
	if w := os.Getenv("PICOCLAW_WORKSPACE"); w != "" {
		return w
	}
	if w := os.Getenv("OPENLOS_WORKTREE"); w != "" {
		return w
	}
	return "."
}

func openDB(worktree string) (*db.Queries, *sql.DB, error) {
	dir := filepath.Join(worktree, "data")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	dbPath := filepath.Join(dir, "openlos.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, nil, err
	}
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		sqlDB.Close()
		return nil, nil, err
	}
	if _, err := sqlDB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		sqlDB.Close()
		return nil, nil, err
	}
	if _, err := sqlDB.Exec(schema); err != nil {
		sqlDB.Close()
		return nil, nil, err
	}
	return db.New(sqlDB), sqlDB, nil
}

// TimeBlock is the value type for schedule blocks; stored as JSON in the DB.
type TimeBlock struct {
	Time     string `json:"time"`
	Activity string `json:"activity"`
}

func marshalBlocks(blocks []TimeBlock) (string, error) {
	b, err := json.Marshal(blocks)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalBlocks(s string) ([]TimeBlock, error) {
	var blocks []TimeBlock
	if s == "" {
		return blocks, nil
	}
	if err := json.Unmarshal([]byte(s), &blocks); err != nil {
		return nil, err
	}
	return blocks, nil
}

// ideas

func ideasLog(worktree, text string) error {
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	entry, err := q.CreateIdea(context.Background(), db.CreateIdeaParams{
		ID:      uuid.NewString(),
		Text:    text,
		Created: time.Now().Unix(),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Idea captured (id: %s): %q\n", entry.ID, entry.Text)
	return nil
}

func ideasList(worktree string, limit int) error {
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if limit <= 0 {
		limit = int(^uint(0) >> 1) // math.MaxInt
	}
	arr, err := q.ListIdeas(context.Background(), int64(limit))
	if err != nil {
		return err
	}
	if len(arr) == 0 {
		fmt.Println("No ideas captured yet.")
		return nil
	}
	for i, it := range arr {
		fmt.Printf("%d. [%s] %s  (id: %s)\n", i+1, time.Unix(it.Created, 0).Format(time.DateTime), it.Text, it.ID)
	}
	return nil
}

// goals

func goalsAdd(worktree, title, description string) error {
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	g, err := q.CreateGoal(context.Background(), db.CreateGoalParams{
		ID:          uuid.NewString(),
		Title:       title,
		Description: &description,
		Status:      "active",
		Created:     time.Now().Unix(),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Goal added (id: %s): \"%s\"\n", g.ID, g.Title)
	return nil
}

func goalsList(worktree, status string) error {
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	var arr []db.Goal
	if status == "" {
		arr, err = q.ListGoals(context.Background())
	} else {
		arr, err = q.ListGoalsByStatus(context.Background(), status)
	}
	if err != nil {
		return err
	}
	if len(arr) == 0 {
		if status == "" {
			fmt.Println("No goals found.")
		} else {
			fmt.Println("No goals match the given filter.")
		}
		return nil
	}
	for _, g := range arr {
		desc := ""
		if g.Description != nil && strings.TrimSpace(*g.Description) != "" {
			desc = "\n    " + *g.Description
		}
		fmt.Printf("- [%s] %s  (id: %s)%s\n", g.Status, g.Title, g.ID, desc)
	}
	return nil
}

func goalsUpdate(worktree, id, status, description string) error {
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if status != "" && status != "active" && status != "completed" && status != "paused" {
		return errors.New("Invalid status. Must be one of: active, completed, paused")
	}

	ctx := context.Background()
	current, err := q.GetGoal(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("Goal not found: " + id)
		}
		return err
	}
	newStatus := current.Status
	if status != "" {
		newStatus = status
	}
	newDesc := current.Description
	if description != "" {
		newDesc = &description
	}
	updated, err := q.UpdateGoal(ctx, db.UpdateGoalParams{
		ID:          id,
		Status:      newStatus,
		Description: newDesc,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Goal updated (id: %s): status=%s\n", updated.ID, updated.Status)
	return nil
}

// tasks

func tasksAdd(worktree, title, goalID, due string) error {
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

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
	t, err := q.CreateTask(context.Background(), db.CreateTaskParams{
		ID:      uuid.NewString(),
		Title:   title,
		GoalID:  goalPtr,
		Status:  "open",
		Created: time.Now().Unix(),
		Due:     duePtr,
	})
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
			return errors.New("Goal not found: " + *goalPtr)
		}
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
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	ctx := context.Background()
	var arr []db.Task

	hasStatus := status != ""
	hasGoal := goalID != ""

	switch {
	case hasStatus && hasGoal:
		g := goalID
		arr, err = q.ListTasksByStatusAndGoal(ctx, db.ListTasksByStatusAndGoalParams{
			Status: status,
			GoalID: &g,
		})
	case hasStatus:
		arr, err = q.ListTasksByStatus(ctx, status)
	case hasGoal:
		g := goalID
		arr, err = q.ListTasksByGoal(ctx, &g)
	default:
		arr, err = q.ListTasks(ctx)
	}
	if err != nil {
		return err
	}
	if len(arr) == 0 {
		if hasStatus || hasGoal {
			fmt.Println("No tasks match the given filters.")
		} else {
			fmt.Println("No tasks found.")
		}
		return nil
	}
	for _, t := range arr {
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
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	ctx := context.Background()
	current, err := q.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("Task not found: " + id)
		}
		return err
	}
	newStatus := current.Status
	if status != "" {
		newStatus = status
	}
	newDue := current.Due
	if due != "" {
		if due == "null" {
			newDue = nil
		} else {
			d := due
			newDue = &d
		}
	}
	updated, err := q.UpdateTask(ctx, db.UpdateTaskParams{
		ID:     id,
		Status: newStatus,
		Due:    newDue,
	})
	if err != nil {
		return err
	}
	dueVal := "none"
	if updated.Due != nil {
		dueVal = *updated.Due
	}
	fmt.Printf("Task updated (id: %s): status=%s, due=%s\n", updated.ID, updated.Status, dueVal)
	return nil
}

// schedule

func scheduleRead(worktree, date string) error {
	if date == "" {
		date = time.Now().UTC().Format(time.DateOnly)
	}
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	row, err := q.GetSchedule(context.Background(), date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("No schedule found for %s.\n", date)
			return nil
		}
		return err
	}
	blocks, err := unmarshalBlocks(*row.Blocks)
	if err != nil {
		return err
	}
	blockStr := "  (no time blocks set)"
	if len(blocks) > 0 {
		var sb strings.Builder
		for _, b := range blocks {
			fmt.Fprintf(&sb, "  %s  %s\n", b.Time, b.Activity)
		}
		blockStr = strings.TrimRight(sb.String(), "\n")
	}
	fmt.Printf("Schedule for %s\nFocus: %s\n\n%s\n", date, row.Focus, blockStr)
	return nil
}

func scheduleWrite(worktree, date, focus string, blocks []string) error {
	if date == "" {
		date = time.Now().UTC().Format(time.DateOnly)
	}
	q, sqlDB, err := openDB(worktree)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	tb := []TimeBlock{}
	for _, b := range blocks {
		parts := strings.SplitN(b, "|", 2)
		if len(parts) != 2 {
			return errors.New("invalid block format, expected HH:MM|Activity")
		}
		tb = append(tb, TimeBlock{Time: parts[0], Activity: parts[1]})
	}
	blocksJSON, err := marshalBlocks(tb)
	if err != nil {
		return err
	}
	_, err = q.UpsertSchedule(context.Background(), db.UpsertScheduleParams{
		Date:   date,
		Focus:  focus,
		Blocks: &blocksJSON,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Schedule saved for %s: \"%s\" with %d time block(s).\n", date, focus, len(tb))
	return nil
}

// install

func copyEmbeddedFile(embedded fs.FS, path, target string) error {
	src, err := embedded.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func install(dir string, force bool) error {
	embedded, err := assets.FS()
	if err != nil {
		return fmt.Errorf("load embedded assets: %w", err)
	}

	// Write embedded .picoclaw files.
	err = fs.WalkDir(embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dir, ".picoclaw", filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !force {
			if _, err := os.Stat(target); err == nil {
				fmt.Printf("  skip (exists): %s\n", target)
				return nil
			}
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := copyEmbeddedFile(embedded, path, target); err != nil {
			return err
		}
		fmt.Printf("  write: %s\n", target)
		return nil
	})
	if err != nil {
		return err
	}

	// Copy the running binary to .picoclaw/bin/openlos.
	binDst := filepath.Join(dir, ".picoclaw", "bin", "openlos")
	if !force {
		if _, err := os.Stat(binDst); err == nil {
			fmt.Printf("  skip (exists): %s\n", binDst)
			return nil
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	// Follow symlinks (e.g. go run produces a temp binary).
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("eval symlinks: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".picoclaw", "bin"), 0o755); err != nil {
		return err
	}
	in, err := os.Open(exe)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(binDst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	fmt.Printf("  write: %s\n", binDst)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: openlos <ideas|tasks|goals|schedule|install> <subcommand> [flags]")
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
	case "install":
		fs := flag.NewFlagSet("install", flag.ExitOnError)
		dir := fs.String("dir", ".", "target directory to install .opencode into")
		force := fs.Bool("force", false, "overwrite existing files")
		fs.Parse(os.Args[2:])
		d := *dir
		if d == "." {
			var err error
			d, err = os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		}
		fmt.Printf("Installing openlos into %s/.picoclaw/\n", d)
		if err := install(d, *force); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println("Done.")
	default:
		fmt.Println("unknown command:", cmd)
		os.Exit(2)
	}
}
