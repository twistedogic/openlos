package main

import (
	"encoding/json"
	"io/ioutil"
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
	// call ideasLog directly
	if err := ideasLog(td, "unit test idea"); err != nil {
		t.Fatalf("ideasLog failed: %v", err)
	}
	// verify file
	p := filepath.Join(td, "data", "ideas.json")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		t.Fatalf("read ideas.json: %v", err)
	}
	var arr []Idea
	if err := json.Unmarshal(b, &arr); err != nil {
		t.Fatalf("unmarshal ideas: %v", err)
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
	// verify file
	gp := filepath.Join(td, "data", "goals.json")
	gb, err := ioutil.ReadFile(gp)
	if err != nil {
		t.Fatalf("read goals.json: %v", err)
	}
	var goals []Goal
	if err := json.Unmarshal(gb, &goals); err != nil {
		t.Fatalf("unmarshal goals: %v", err)
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
	// re-read and verify status
	gb2, _ := ioutil.ReadFile(gp)
	var goals2 []Goal
	if err := json.Unmarshal(gb2, &goals2); err != nil {
		t.Fatalf("unmarshal goals after update: %v", err)
	}
	if goals2[0].Status != "completed" {
		t.Fatalf("expected status completed, got %s", goals2[0].Status)
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

	// also inspect schedule.json
	sp := filepath.Join(td, "data", "schedule.json")
	sb, err := ioutil.ReadFile(sp)
	if err != nil {
		t.Fatalf("read schedule.json: %v", err)
	}
	var sched Schedule
	if err := json.Unmarshal(sb, &sched); err != nil {
		t.Fatalf("unmarshal schedule: %v", err)
	}
	if day, ok := sched["2026-03-01"]; !ok || day.Focus != "Sprint Planning" || len(day.Blocks) != 2 {
		t.Fatalf("unexpected schedule content: %+v", sched)
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
	// read tasks.json
	tp := filepath.Join(td, "data", "tasks.json")
	tb, err := ioutil.ReadFile(tp)
	if err != nil {
		t.Fatalf("read tasks.json: %v", err)
	}
	var tasks []Task
	if err := json.Unmarshal(tb, &tasks); err != nil {
		t.Fatalf("unmarshal tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Due == nil || *tasks[0].Due != "2026-03-05" {
		t.Fatalf("unexpected task after add: %+v", tasks)
	}

	// update to clear due
	id := tasks[0].ID
	// pause briefly to ensure timestamps differ if needed
	time.Sleep(10 * time.Millisecond)
	cmd = exec.Command(bin, "tasks", "update", "--id", id, "--due", "null", "--worktree", td)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tasks update failed: %v\n%s", err, string(out))
	}
	tb2, err := ioutil.ReadFile(tp)
	if err != nil {
		t.Fatalf("read tasks.json after update: %v", err)
	}
	var tasks2 []Task
	if err := json.Unmarshal(tb2, &tasks2); err != nil {
		t.Fatalf("unmarshal tasks after update: %v", err)
	}
	if tasks2[0].Due != nil {
		t.Fatalf("expected due cleared, got %+v", tasks2[0].Due)
	}
}
