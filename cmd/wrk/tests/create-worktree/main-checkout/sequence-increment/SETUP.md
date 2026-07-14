# Scenario

**Feature**: second wrk increments collision suffix (always new branch; P0 C1)

```
# first wrk creates myrepo-main-{date} + branch main-{date}
# second wrk creates myrepo-main-{date}-1 + NEW branch main-{date}-1 (never reuses first branch)
myrepo (main) -> wrk -> myrepo-main-2026-06-30
myrepo (main) -> wrk -> myrepo-main-2026-06-30-1
```

## Steps

1. Initialize git repo `myrepo` on branch `main`.
2. Run `wrk` once (first invocation) from `myrepo`.
3. Run `wrk` again via the doctest `Run` function (second invocation).

```go
func Setup(t *testing.T, req *Request) error {
	initGitRepoOnMain(t, req.RepoDir)
	firstPath := runWrkFrom(t, req, req.RepoDir)
	wantFirst := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	if firstPath != wantFirst {
		t.Fatalf("first wrk: expected %q, got %q", wantFirst, firstPath)
	}
	return nil
}
```