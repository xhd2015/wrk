# Scenario

**Feature**: wrk --bring --exec runs in external worktree after non-git soft SKIP

```
# non-git cwd + --exec pwd
#   -> exit 0; SKIP on stderr; worktree under plain/external/
#   -> stdout: <external-abs>\n<external-abs>\n  (mode path then child pwd)
plain (no .git) + mydep
  -> wrk --bring <dep> --exec pwd
  -> child cmd.Dir = external abs
```

## Steps

1. Create plain non-git cwd and valid dep repo.
2. Run `wrk --bring <dep> --exec pwd` from the plain cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	plain := initBringPlainCwd(t, req.WorkRoot, "plain")
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = plain
	req.ConsumerTop = plain
	req.DepPath = dep
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", dep, "--exec", "pwd"}
	return nil
}
```
