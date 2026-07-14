# Scenario

**Feature**: wrk --dep errors when consumer repo has no Go modules at all

```
# consumer git repo with zero go.mod files -> wrk --dep -> non-zero error
consumer (git, no go.mod anywhere) + dep (valid go.mod) -> wrk --dep -> error
```

## Preconditions

- Git and Go must be available.
- Consumer repo root has NO go.mod, no subdirectory go.mod.
- Consumer cwd is the repo root.

## Steps

1. Create consumer git repo with no go.mod anywhere.
2. Create dep git repo with valid go.mod.
3. Run `wrk --dep <dep>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	// NO go.mod at repo root. No go.mod in any subdirectory.
	writeFile(t, filepath.Join(consumer, "README.md"), "# consumer\n")
	runGitIsolated(t, consumer, "add", "README.md")
	runGitIsolated(t, consumer, "commit", "-m", "init consumer")

	dep := initDepRepo(t, req.WorkRoot, "mydep", true)

	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.Args = []string{"--dep", dep}
	return nil
}
```