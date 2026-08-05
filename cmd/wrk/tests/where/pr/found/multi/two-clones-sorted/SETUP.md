# Scenario

**Feature**: two clones of acme/app each with head worktree → two sorted paths

```
clone-aaa + clone-zzz recorded; each has linked wt on feature-pr
workspace/ -> wrk --where --pr URL
  -> stdout lex-sorted abs paths of both linked worktrees
```

## Steps

1. Seed two mains with same github origin owner/repo; each gets a linked wt on head.
2. Record both; install fake gh; run from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	wherePrSetupTwoClonesOnHead(t, req)
	req.Args = wherePrArgs(wherePrURL)
	return nil
}
```
