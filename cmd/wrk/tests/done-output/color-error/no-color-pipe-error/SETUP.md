# Scenario

**Feature**: pipe default (no `--color`) → hard-error stderr has **no** ANSI (D9)

```
# dirty cascade target preflight
  -> wrk --done   (pipe, no --color)
  -> non-zero
  -> stderr has Error: / dirty language without CSI sequences
```

## Steps

1. Clean contained external + uncommitted file on external.
2. Run bare `wrk --done` (doctest pipe; no `--color`).

```go
import (
	"os/exec"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

const noColorErrDepModule = "example.com/dep-nocolor-err"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoModInDir(t, mainRepo, "edit", "-require="+noColorErrDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+noColorErrDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	externalPath = compositionResolvePath(t, externalPath)
	req.ExternalWtDir = externalPath

	runGoModInDir(t, wtDir, "edit", "-dropreplace="+noColorErrDepModule)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", ".gitignore", "go.mod")
	runGitIsolated(t, wtDir, "commit", "-m", "drop replace; ignore external")

	writeFile(t, filepath.Join(externalPath, "dirty-ext"), "uncommitted")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
