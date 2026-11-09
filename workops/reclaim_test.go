package workops

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanReclaimMainCheckoutHidden(t *testing.T) {
	main := reclaimRepo(t)
	plan, err := PlanReclaim(context.Background(), main)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Visible {
		t.Fatalf("main checkout must hide reclaim: %+v", plan)
	}
}

func TestPlanReclaimIdenticalCleanEnabled(t *testing.T) {
	main := reclaimRepo(t)
	wt := reclaimWorktree(t, main, "topic-clean")
	plan, err := PlanReclaim(context.Background(), wt)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Visible || !plan.Enabled || len(plan.Reasons) != 0 {
		t.Fatalf("identical clean should enable: %+v", plan)
	}
}

func TestPlanReclaimDirtyDisabled(t *testing.T) {
	main := reclaimRepo(t)
	wt := reclaimWorktree(t, main, "topic-dirty")
	if err := os.WriteFile(filepath.Join(wt, "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanReclaim(context.Background(), wt)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Visible || plan.Enabled {
		t.Fatalf("dirty should disable: %+v", plan)
	}
	if !containsReason(plan.Reasons, "uncommitted changes") {
		t.Fatalf("reasons=%v", plan.Reasons)
	}
}

func TestPlanReclaimAheadDisabled(t *testing.T) {
	main := reclaimRepo(t)
	wt := reclaimWorktree(t, main, "topic-ahead")
	reclaimCommit(t, wt, "ahead.txt", "ahead")
	plan, err := PlanReclaim(context.Background(), wt)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Visible || plan.Enabled {
		t.Fatalf("ahead should disable: %+v", plan)
	}
	if !containsReason(plan.Reasons, "ahead of the main checkout") {
		t.Fatalf("reasons=%v", plan.Reasons)
	}
}

func TestPlanReclaimDivergedDisabled(t *testing.T) {
	main := reclaimRepo(t)
	wt := reclaimWorktree(t, main, "topic-div")
	reclaimCommit(t, wt, "wt.txt", "on wt")
	reclaimCommit(t, main, "main.txt", "on main")
	plan, err := PlanReclaim(context.Background(), wt)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Visible || plan.Enabled {
		t.Fatalf("diverged should disable: %+v", plan)
	}
	if !containsReason(plan.Reasons, "diverged") {
		t.Fatalf("reasons=%v", plan.Reasons)
	}
}

func TestReclaimRemovesIncludedWorktree(t *testing.T) {
	main := reclaimRepo(t)
	wt := reclaimWorktree(t, main, "topic-rm")
	if err := Reclaim(context.Background(), ReclaimOptions{WorktreeDir: wt}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists: %v", err)
	}
}

func reclaimRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runReclaimGit(t, dir, "init", "-b", "main")
	runReclaimGit(t, dir, "config", "user.email", "t@t")
	runReclaimGit(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runReclaimGit(t, dir, "add", ".")
	runReclaimGit(t, dir, "commit", "-m", "init")
	return dir
}

func reclaimWorktree(t *testing.T, main, branch string) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), branch)
	runReclaimGit(t, main, "worktree", "add", "-b", branch, wt)
	return wt
}

func reclaimCommit(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runReclaimGit(t, dir, "add", name)
	runReclaimGit(t, dir, "commit", "-m", body)
}

func runReclaimGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func containsReason(reasons []string, part string) bool {
	for _, reason := range reasons {
		if strings.Contains(reason, part) {
			return true
		}
	}
	return false
}
