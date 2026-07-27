# Scenario

**Feature**: wrk --tag-next requires a git work tree

```
# non-git cwd -> error before tagscope
neutral dir -> wrk --tag-next -> not a git repository
```

## Steps

1. Set `req.RepoDir` to a neutral directory without `.git`.
2. Run `wrk --tag-next`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "not-git")
	req.Args = []string{"--tag-next"}
	return nil
}
```