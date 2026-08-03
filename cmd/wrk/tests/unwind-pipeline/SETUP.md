# Scenario

**Feature**: unwind composes repository-local pipeline stages without global process state

```
isolated app checkout -> nested linked dependency worktree -> wrkcli.Capture
wrkcli.Capture -> unwind planner -> per-peel stages or dry-run plan
```

## Preconditions

- Git is available. Each leaf receives a private temporary repository layout.
- The L2 harness uses `CaptureOpts.Dir` and `CaptureOpts.WrkHome`; it does not
  change the process working directory or environment.

## Steps

1. Create an app main repository and, when needed, a dependency main plus a
   linked worktree under `app/external/dep`.
2. Leaves select flags and snapshot refs before invoking `Run`.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	*req = Request{}
	req.DoctestRoot = d.DOCTEST_ROOT
	req.WorkRoot = t.TempDir()
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil { return err }
	return nil
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	// Doctest repositories must not inherit the developer machine's hooks.
	cmd := exec.Command("git", append([]string{"-c", "core.hooksPath=/dev/null", "-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=doctest", "GIT_AUTHOR_EMAIL=doctest@example.invalid", "GIT_COMMITTER_NAME=doctest", "GIT_COMMITTER_EMAIL=doctest@example.invalid")
	out, err := cmd.CombinedOutput()
	if err != nil { t.Fatalf("git %s: %v\\n%s", strings.Join(args, " "), err, out) }
	return strings.TrimSpace(string(out))
}

func setOrigin(t *testing.T, repo, remote string) {
	t.Helper()
	if strings.Contains("\n"+git(t, repo, "remote")+"\n", "\norigin\n") {
		git(t, repo, "remote", "set-url", "origin", remote)
		return
	}
	git(t, repo, "remote", "add", "origin", remote)
}

func seedMain(t *testing.T, req *Request) {
	t.Helper()
	req.MainRepo = filepath.Join(req.WorkRoot, "app")
	if err := os.MkdirAll(req.MainRepo, 0o755); err != nil { t.Fatal(err) }
	git(t, req.MainRepo, "init", "-b", "main")
	git(t, req.MainRepo, "config", "core.hooksPath", "/dev/null")
	os.WriteFile(filepath.Join(req.MainRepo, "go.mod"), []byte("module example.com/app\\n\\ngo 1.22\\n"), 0o644)
	os.WriteFile(filepath.Join(req.MainRepo, ".gitignore"), []byte("/external/\\n"), 0o644)
	git(t, req.MainRepo, "add", ".")
	git(t, req.MainRepo, "commit", "-m", "initial")
	mainRemote := filepath.Join(req.WorkRoot, "app-origin.git")
	git(t, req.WorkRoot, "init", "--bare", mainRemote)
	setOrigin(t, req.MainRepo, mainRemote)
	req.RepoDir = req.MainRepo
}

func seedLinkedDep(t *testing.T, req *Request) {
	t.Helper()
	seedMain(t, req)
	req.DepMain = filepath.Join(req.WorkRoot, "dep")
	os.MkdirAll(req.DepMain, 0o755)
	git(t, req.DepMain, "init", "-b", "main")
	git(t, req.DepMain, "config", "core.hooksPath", "/dev/null")
	os.WriteFile(filepath.Join(req.DepMain, "go.mod"), []byte("module example.com/dep\\n\\ngo 1.22\\n"), 0o644)
	git(t, req.DepMain, "add", ".")
	git(t, req.DepMain, "commit", "-m", "initial")
	depRemote := filepath.Join(req.WorkRoot, "dep-origin.git")
	git(t, req.WorkRoot, "init", "--bare", depRemote)
	setOrigin(t, req.DepMain, depRemote)
	req.DepWorktree = filepath.Join(req.MainRepo, "external", "dep")
	os.MkdirAll(filepath.Dir(req.DepWorktree), 0o755)
	git(t, req.DepMain, "worktree", "add", "-b", "feature-dep", req.DepWorktree)
	os.WriteFile(filepath.Join(req.DepWorktree, "change.txt"), []byte("generated commit input\\n"), 0o644)
	req.BeforeDep = git(t, req.DepMain, "rev-parse", "HEAD")
	req.BeforeMain = git(t, req.MainRepo, "rev-parse", "HEAD")
}

func assertExit0(t *testing.T, r *Response) { t.Helper(); if r.ExitCode != 0 { t.Fatalf("exit=%d stdout=%q stderr=%q", r.ExitCode, r.Stdout, r.Stderr) } }
func assertContainsInOrder(t *testing.T, s string, parts ...string) { t.Helper(); at := 0; for _, p := range parts { i := strings.Index(s[at:], p); if i < 0 { t.Fatalf("missing %q in %q", p, s) }; at += i + len(p) } }

func findUnwindModuleRoot(dir string) string {
	for dir != filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil { return dir }
		dir = filepath.Dir(dir)
	}
	return ""
}

func installFakeOpencode(t *testing.T, req *Request) {
	t.Helper()
	modRoot := findUnwindModuleRoot(req.DoctestRoot)
	if modRoot == "" { t.Fatal("cannot find wrk module root from doctest root") }
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(modRoot), "agent-pro-*", "cmd", "fake-opencode"))
	if err != nil || len(matches) == 0 { t.Skip("agent-pro cmd/fake-opencode fixture is unavailable") }
	fakeRoot := matches[len(matches)-1]
	req.FakeOpencode = filepath.Join(req.WorkRoot, "bin", "fake-opencode")
	if err := os.MkdirAll(filepath.Dir(req.FakeOpencode), 0o755); err != nil { t.Fatal(err) }
	cmd := exec.Command("go", "build", "-mod=mod", "-o", req.FakeOpencode, ".")
	cmd.Dir = fakeRoot
	if out, err := cmd.CombinedOutput(); err != nil { t.Fatalf("build fake-opencode: %v\\n%s", err, out) }
	mock := filepath.Join(req.WorkRoot, "fake-opencode.json")
	body := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"unwind","llm_events":[{"type":"message","text":"{\"title\": \"feat: unwind\", \"description\": \"offline test commit\"}"}]}`
	if err := os.WriteFile(mock, []byte(body), 0o644); err != nil { t.Fatal(err) }
	configDir := filepath.Join(req.WorkRoot, "opencode-config")
	if err := os.MkdirAll(configDir, 0o755); err != nil { t.Fatal(err) }
	req.ExtraEnv = []string{"FAKE_OPENCODE_MOCK_CONFIG=" + mock, "OPENCODE_CONFIG_DIR=" + configDir}
}

func unwindGenCommitArgs(t *testing.T, req *Request, stages ...string) []string {
	t.Helper()
	installFakeOpencode(t, req)
	args := []string{"--unwind", "--gen-commit-msg", "--commit", "--agent-runner", "opencode", "--agent-runner-binary", req.FakeOpencode}
	return append(args, stages...)
}

// pathInHEAD reports whether path is present in the commit at ref (default HEAD).
func pathInHEAD(t *testing.T, repo, path string) bool {
	t.Helper()
	cmd := exec.Command("git", "-c", "core.hooksPath=/dev/null", "-C", repo, "ls-tree", "-r", "--name-only", "HEAD")
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=doctest", "GIT_AUTHOR_EMAIL=doctest@example.invalid", "GIT_COMMITTER_NAME=doctest", "GIT_COMMITTER_EMAIL=doctest@example.invalid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ls-tree HEAD: %v\n%s", err, out)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == path {
			return true
		}
	}
	return false
}

// isUntracked reports whether path is untracked in repo (git status --porcelain).
func isUntracked(t *testing.T, repo, path string) bool {
	t.Helper()
	out := git(t, repo, "status", "--porcelain", "--", path)
	// "?? path" for untracked
	return strings.HasPrefix(strings.TrimSpace(out), "??")
}

// isTrackedModified reports whether path has staged or unstaged tracked changes.
func porcelainFor(t *testing.T, repo, path string) string {
	t.Helper()
	return strings.TrimSpace(git(t, repo, "status", "--porcelain", "--", path))
}
```

