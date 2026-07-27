# Scenario

**Feature**: dry-run still runs preflight; dirty cascade target → error, no false `would:` success (D6/D7)

```
# dirty external + clean own
  -> wrk --done --dry-run
  -> non-zero + Error: (dirty)
  -> must NOT print successful would: cascade merge-back as if OK
  -> external + consumer still present
```

## Steps

1. Same clean contained fixture as with-cascade, then dirtify external.
2. Run `wrk --done --dry-run`.

```go
import (
	"os/exec"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

const dryRunDirtyCascadeDepModule = "example.com/dep-dry-dirty"

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
	runGoModInDir(t, mainRepo, "edit", "-require="+dryRunDirtyCascadeDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+dryRunDirtyCascadeDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	externalPath = compositionResolvePath(t, externalPath)
	req.ExternalWtDir = externalPath

	runGoModInDir(t, wtDir, "edit", "-dropreplace="+dryRunDirtyCascadeDepModule)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", ".gitignore", "go.mod")
	runGitIsolated(t, wtDir, "commit", "-m", "drop replace; ignore external")

	writeFile(t, filepath.Join(externalPath, "dirty-ext"), "uncommitted")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--dry-run"}
	return nil
}
```
