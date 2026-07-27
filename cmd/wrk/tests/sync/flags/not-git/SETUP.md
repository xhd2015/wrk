# Scenario

**Feature**: wrk --sync requires a git work tree

```
# non-git cwd -> error before sync body
neutral dir -> wrk --sync -> not a git repository
```

## Steps

1. Set `req.RepoDir` to a neutral directory without `.git`.
2. Run `wrk --sync`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "not-git")
	req.Args = []string{"--sync"}
	return nil
}
```
