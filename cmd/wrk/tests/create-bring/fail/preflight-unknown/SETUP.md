# Scenario

**Feature**: unresolvable `--bring` basename fails before create

```
src cwd -> wrk --new --no-config --bring no-such-basename
  -> non-zero before create
  -> no new WT under WRK_HOME/worktrees
```

## Steps

1. Create `src` (git Go module).
2. Run `--new --bring no-such-basename` from `src`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	req.RepoDir = src
	req.MainRepo = src
	req.Args = []string{"--new", "--no-config", "--bring", "no-such-basename"}
	return nil
}
```
