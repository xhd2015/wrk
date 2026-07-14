# Scenario

**Feature**: `--exec` after successful `--dep` runs in the external worktree

```
consumer requires dep -> wrk --dep <dep> --exec pwd
  -> external wt under consumer/external/
  -> stdout: <external-abs>\n<external-abs>\n
  -> child cmd.Dir = external abs
```

## Preconditions

- Git and Go available.
- Consumer has go.mod requiring dep; dep is a git Go module.

## Steps

- Build consumer + dep under WorkRoot; leave Args/RepoDir to leaves.

```go
import (
	"fmt"
	"os/exec"
	"path/filepath"
)

const execDepModulePath = "example.com/dep"

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	ensureExecDepHelpersUsed()
	return nil
}

func initExecConsumer(t *testing.T, workRoot string) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoModInDir(t, consumer, "edit", "-require="+execDepModulePath+"@v0.0.0")
	resolved, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	return resolved
}

func initExecDepRepo(t *testing.T, workRoot, name string) string {
	t.Helper()
	dep := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, dep)
	writeFile(t, filepath.Join(dep, "go.mod"), "module "+execDepModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dep, "dep.go"), "package dep\n")
	runGitIsolated(t, dep, "add", "go.mod", "dep.go")
	runGitIsolated(t, dep, "commit", "-m", "add go module")
	return dep
}

func execExternalWorktreePath(consumerTop, depBasename, token string, suffix int) string {
	name := fmt.Sprintf("%s-%s-%s", depBasename, token, wrkDate)
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return filepath.Join(consumerTop, "external", name)
}

func ensureExecDepHelpersUsed() {
	_ = initExecConsumer
	_ = initExecDepRepo
	_ = execExternalWorktreePath
}
```
