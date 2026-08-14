# Scenario

**Feature**: apply fail after create keeps the worktree (no rollback)

```
# first dep valid git; second path exists but is not a git repo
src cwd -> wrk --new --no-config --bring <d1> <not-git>
  -> non-zero
  -> create path exists; first external may exist
  -> do not require rollback
```

## Steps

1. Create `src` requiring dep1 + valid `mydep1`.
2. Create a non-git directory `not-a-git` under WorkRoot.
3. Run `--new --bring d1 not-a-git`.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)
	notGit := filepath.Join(req.WorkRoot, "not-a-git")
	if err := os.MkdirAll(notGit, 0o755); err != nil {
		t.Fatalf("mkdir not-a-git: %v", err)
	}
	writeFile(t, filepath.Join(notGit, "readme.txt"), "not a git repo\n")

	req.RepoDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.SecondRepo = notGit
	req.ConsumerTop = createBringDefaultWT(req)
	req.Args = []string{"--new", "--no-config", "--bring", dep1, notGit}
	return nil
}
```
