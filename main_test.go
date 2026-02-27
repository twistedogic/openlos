package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join("..", "bin", "openlos")
	// always rebuild to pick up source changes
	cmd := exec.Command("go", "build", "-o", bin)
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}
	return bin
}

func TestIdeasRoundtrip(t *testing.T) {
	td := t.TempDir()
	if err := ideasLog(td, "unit test idea"); err != nil {
		t.Fatalf("ideasLog failed: %v", err)
	}
	// verify via list
	q, sqlDB, err := openDB(td)
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}
	defer sqlDB.Close()
	arr, err := q.ListIdeas(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListIdeas: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 idea, got %d", len(arr))
	}
	if arr[0].Text != "unit test idea" {
		t.Fatalf("unexpected idea text: %s", arr[0].Text)
	}
}

func TestCLISmoke(t *testing.T) {
	td := t.TempDir()
	bin := buildBinary(t)

	// ideas log via CLI
	cmd := exec.Command(bin, "ideas", "log", "--text", "cli idea", "--worktree", td)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ideas log CLI failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "Idea captured") {
		t.Fatalf("unexpected output: %s", string(out))
	}

	// ideas list via CLI
	cmd = exec.Command(bin, "ideas", "list", "--limit", "10", "--worktree", td)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ideas list CLI failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "cli idea") {
		t.Fatalf("ideas list did not include idea: %s", string(out))
	}

	// tasks add and list
	cmd = exec.Command(bin, "tasks", "add", "--title", "cli task", "--worktree", td)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tasks add failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "Task added") {
		t.Fatalf("unexpected tasks add output: %s", string(out))
	}

	cmd = exec.Command(bin, "tasks", "list", "--worktree", td)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tasks list failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "cli task") {
		t.Fatalf("tasks list did not include task: %s", string(out))
	}
}

func TestGoalsAndSchedule(t *testing.T) {
	td := t.TempDir()
	bin := buildBinary(t)

	// goals add (direct helper)
	if err := goalsAdd(td, "goal one", "do important stuff"); err != nil {
		t.Fatalf("goalsAdd failed: %v", err)
	}
	// verify via DB
	q, sqlDB, err := openDB(td)
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}
	defer sqlDB.Close()
	goals, err := q.ListGoals(context.Background())
	if err != nil {
		t.Fatalf("ListGoals: %v", err)
	}
	if len(goals) != 1 || goals[0].Title != "goal one" {
		t.Fatalf("unexpected goals content: %+v", goals)
	}

	// goals list via CLI
	cmd := exec.Command(bin, "goals", "list", "--worktree", td)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goals list CLI failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "goal one") {
		t.Fatalf("goals list did not include goal: %s", string(out))
	}

	// update goal status via CLI
	gid := goals[0].ID
	cmd = exec.Command(bin, "goals", "update", "--id", gid, "--status", "completed", "--worktree", td)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goals update CLI failed: %v\n%s", err, string(out))
	}
	// re-verify via DB
	q2, sqlDB2, err := openDB(td)
	if err != nil {
		t.Fatalf("openDB after update: %v", err)
	}
	defer sqlDB2.Close()
	updated, err := q2.GetGoal(context.Background(), gid)
	if err != nil {
		t.Fatalf("GetGoal: %v", err)
	}
	if updated.Status != "completed" {
		t.Fatalf("expected status completed, got %s", updated.Status)
	}

	// schedule write via CLI
	cmd = exec.Command(bin, "schedule", "write", "--date", "2026-03-01", "--focus", "Sprint Planning", "--blocks", "09:00|Plan,13:00|Lunch", "--worktree", td)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("schedule write failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "Schedule saved for 2026-03-01") {
		t.Fatalf("unexpected schedule write output: %s", string(out))
	}

	// schedule read via CLI
	cmd = exec.Command(bin, "schedule", "read", "--date", "2026-03-01", "--worktree", td)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("schedule read failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "Sprint Planning") || !strings.Contains(string(out), "09:00") {
		t.Fatalf("schedule read missing content: %s", string(out))
	}

	// verify schedule via DB
	q3, sqlDB3, err := openDB(td)
	if err != nil {
		t.Fatalf("openDB for schedule: %v", err)
	}
	defer sqlDB3.Close()
	sched, err := q3.GetSchedule(context.Background(), "2026-03-01")
	if err != nil {
		t.Fatalf("GetSchedule: %v", err)
	}
	if sched.Focus != "Sprint Planning" {
		t.Fatalf("unexpected schedule focus: %s", sched.Focus)
	}
	blocks, err := unmarshalBlocks(*sched.Blocks)
	if err != nil {
		t.Fatalf("unmarshalBlocks: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
}

func TestTasksUpdateClearDue(t *testing.T) {
	td := t.TempDir()
	bin := buildBinary(t)

	// add task with due via CLI
	cmd := exec.Command(bin, "tasks", "add", "--title", "t1", "--due", "2026-03-05", "--worktree", td)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tasks add failed: %v\n%s", err, string(out))
	}

	// verify via DB
	q, sqlDB, err := openDB(td)
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}
	defer sqlDB.Close()
	tasks, err := q.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Due == nil || *tasks[0].Due != "2026-03-05" {
		t.Fatalf("unexpected task after add: %+v", tasks)
	}
	taskID := tasks[0].ID
	sqlDB.Close()

	// pause briefly so created timestamps differ if needed
	time.Sleep(10 * time.Millisecond)

	// update to clear due
	cmd = exec.Command(bin, "tasks", "update", "--id", taskID, "--due", "null", "--worktree", td)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tasks update failed: %v\n%s", err, string(out))
	}

	// verify due is cleared
	q2, sqlDB2, err := openDB(td)
	if err != nil {
		t.Fatalf("openDB after update: %v", err)
	}
	defer sqlDB2.Close()
	updated, err := q2.GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if updated.Due != nil {
		t.Fatalf("expected due cleared, got %+v", updated.Due)
	}
}
