# Scenario

**Feature**: second `wrk --bring` of the same required dep reuses the first external worktree

```
# first --bring creates external/mydep + replace
# second --bring same dep -> same stdout path; reuse warning; no -1 dir
consumer (require dep) + mydep -> wrk --bring mydep (1st)
  -> external/mydep
consumer -> wrk --bring mydep (2nd)
  -> reuse external/mydep
  -> stderr reuse warning
```

## Steps

1. Create consumer requiring `example.com/dep` and dep repo `mydep`.
2. Run `wrk --bring <dep>` once; record external path.
3. Run `wrk --bring <dep>` again via doctest `Run` (second invocation).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureBringReuseHelpersUsed()

	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	first := runWrkWithArgs(t, req, consumer, "--bring", dep)
	wantFirst := bringExternalWorktreePath(consumer, "mydep", "main", 0)
	if first != wantFirst {
		t.Fatalf("first --bring: expected %q, got %q", wantFirst, first)
	}

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.ExternalWtDir = wantFirst
	req.Args = []string{"--bring", dep}
	return nil
}
```
