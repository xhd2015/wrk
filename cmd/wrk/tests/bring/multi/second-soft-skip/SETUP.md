# Scenario

**Feature**: multi-bring continues when the second dep soft-SKIPs (not a dependency)

```
# dep1 required; dep2 is a go module but not required by consumer
# wrk --bring dep1 --bring dep2
#   -> both external worktrees + two stdout lines; exit 0
#   -> replace only for dep1; SKIP not a dependency for dep2 on stderr
consumer (require dep1 only) + mydep1 + mydep2
  -> multi-bring -> soft SKIP second; continue
```

## Steps

1. Create consumer requiring only `example.com/dep1` (import dep1 only).
2. Create `mydep1` (required) and `mydep2` (module `example.com/dep2`, not required).
3. Run `wrk --bring <dep1> --bring <dep2>`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true

	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runBringGo(t, consumer, "mod", "edit", "-require="+multiBringDep1Module+"@v0.0.0")
	var err error
	consumer, err = filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks consumer: %v", err)
	}

	dep1 := initMultiBringDepRepo(t, req.WorkRoot, "mydep1", multiBringDep1Module)
	dep2 := initMultiBringDepRepo(t, req.WorkRoot, "mydep2", multiBringDep2Module)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.SecondRepo = dep2
	req.Args = []string{"--bring", dep1, "--bring", dep2}
	return nil
}
```
