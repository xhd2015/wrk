# Scenario

**Feature**: `--no-dep` with multi-bring materializes all worktrees without replace/tidy

```
# --no-dep applies to every brought dep
consumer requires dep1+dep2
  -> wrk --bring <d1> --bring <d2> --no-dep
  -> two external paths; go.mod byte-identical; no SKIP; exit 0
```

## Steps

1. Create consumer requiring both deps + both dep repos.
2. Snapshot consumer `go.mod`.
3. Run multi-bring with `--no-dep`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initMultiBringConsumerWithTwoRequires(t, req.WorkRoot)
	dep1 := initMultiBringDepRepo(t, req.WorkRoot, "mydep1", multiBringDep1Module)
	dep2 := initMultiBringDepRepo(t, req.WorkRoot, "mydep2", multiBringDep2Module)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.SecondRepo = dep2
	multiSnapshotBringGoMod(t, req, consumer)
	req.Args = []string{"--bring", dep1, "--bring", dep2, "--no-dep"}
	return nil
}
```
