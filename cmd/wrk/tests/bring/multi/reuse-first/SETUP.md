# Scenario

**Feature**: multi-bring reuses Policy A for a dep already under external/, creates the other

```
# dep1 already brought once; multi --bring dep1 --bring dep2
#   -> stdout: reused dep1 path + new dep2 path
#   -> stderr reuse warning for dep1; exactly 2 external dirs
consumer (require both)
  -> precondition: wrk --bring dep1
  -> wrk --bring dep1 --bring dep2
```

## Steps

1. Create consumer requiring both deps + both dep repos.
2. Pre-run single `--bring dep1` via `runWrkWithArgs`; record path.
3. Run multi-bring of dep1 then dep2 via doctest `Run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initMultiBringConsumerWithTwoRequires(t, req.WorkRoot)
	dep1 := initMultiBringDepRepo(t, req.WorkRoot, "mydep1", multiBringDep1Module)
	dep2 := initMultiBringDepRepo(t, req.WorkRoot, "mydep2", multiBringDep2Module)

	first := runWrkWithArgs(t, req, consumer, "--bring", dep1)
	want1 := bringExternalWorktreePath(consumer, "mydep1", "main", 0)
	if first != want1 {
		t.Fatalf("precondition --bring mydep1: expected %q, got %q", want1, first)
	}
	// Precondition tidy drops unused dep2 require; re-pin so multi-bring can match.
	runBringGo(t, consumer, "mod", "edit", "-require="+multiBringDep2Module+"@v0.0.0")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.SecondRepo = dep2
	req.ExternalWtDir = want1
	req.Args = []string{"--bring", dep1, "--bring", dep2}
	return nil
}
```
