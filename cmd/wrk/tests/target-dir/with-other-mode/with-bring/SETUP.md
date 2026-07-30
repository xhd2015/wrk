# Scenario

**Feature**: <target-dir> + --bring <dep> → wrk: unexpected arguments (target-dir is create-only)

```
# consumer requires dep; --bring is not the create path; second positional rejected
consumer (go.mod + git) + dep repo -> wrk <consumer> <target-dir> --bring <dep> -> non-zero
```

## Preconditions

- Git and Go must be available on PATH.

## Steps

1. Build a consumer git repo `consumer` with a `go.mod` requiring `example.com/dep`.
2. Build a dep git repo `mydep` with module `example.com/dep`.
3. Set `req.TargetDir = {WorkRoot}/consumer` (first positional, the consumer repo).
4. Set `req.SpawnDir = {WorkRoot}/wt`.
5. Set `req.Args = ["--bring", {WorkRoot}/mydep]`.
6. Run `wrk consumer {WorkRoot}/wt --bring {WorkRoot}/mydep` from process cwd `{WorkRoot}`.

```go
import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

const tdDepModulePath = "example.com/dep"

func tdRunGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}

	// consumer repo requiring the dep
	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	tdRunGo(t, consumer, "mod", "edit", "-require="+tdDepModulePath+"@v0.0.0")
	runGitIsolated(t, consumer, "add", "go.mod")
	runGitIsolated(t, consumer, "commit", "-m", "require dep")

	// dep repo with a go.mod
	dep := filepath.Join(req.WorkRoot, "mydep")
	initGitRepoOnMain(t, dep)
	writeFile(t, filepath.Join(dep, "go.mod"), "module "+tdDepModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dep, "dep.go"), "package dep\n")
	runGitIsolated(t, dep, "add", "go.mod", "dep.go")
	runGitIsolated(t, dep, "commit", "-m", "add go module")

	req.TargetDir = consumer
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"--bring", dep}
	return nil
}
```
