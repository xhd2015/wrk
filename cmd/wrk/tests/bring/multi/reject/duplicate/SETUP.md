# Scenario

**Feature**: exact duplicate `--bring` paths in one invocation are rejected

```
# same resolved path twice
wrk --bring <dep1> --bring <dep1>
  -> non-zero; error about duplicate / same path
  -> prefer no external/ created
```

## Steps

1. Create consumer requiring dep1 and a single valid dep repo.
2. Run `wrk --bring <dep> --bring <dep>` with the same path twice.

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

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.Args = []string{"--bring", dep1, "--bring", dep1}
	return nil
}
```
