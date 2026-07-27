# Scenario

**Feature**: multiple live external worktrees of the same depMain → reuse lex-smallest + multi warnings

```
# legacy: two external WTs of mydep under consumer/external/
# --bring reuses lex-smallest abs path; stderr: count + also present
consumer/external/mydep-main-{date}
consumer/external/mydep-main-{date}-1
  -> wrk --bring mydep
  -> stdout = lex-smallest path
  -> stderr multi reuse + also present
```

## Steps

1. Create consumer requiring dep + dep repo `mydep`.
2. Run first `--bring` to create the preferred external path (suffix 0).
3. Manually add a second linked worktree of the dep under `external/…-1` with branch `main-{date}-1`.
4. Run `wrk --bring <dep>` via `Run`.

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

	// Second legacy external WT of the same dep main (path suffix -1).
	second := bringExternalWorktreePath(consumer, "mydep", "main", 1)
	runGitIsolated(t, dep, "worktree", "add", "-b", branchName("main", wrkDate, 1), second)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.ExternalWtDir = wantFirst // lex-smallest expected reuse target
	req.ExternalWtDir2 = second
	req.Args = []string{"--bring", dep}
	return nil
}
```
