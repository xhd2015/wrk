# Scenario

**Feature**: wrk --bring --no-dep from non-git cwd still creates external worktree without SKIP replace

```
# plain non-git + --no-dep: skip module analyse entirely
#   -> external under plain/external/; no SKIP local dep replacement
#   -> no .gitignore (non-git parent)
plain (no .git) + mydep
  -> wrk --bring <dep> --no-dep
  -> stdout abs path; no SKIP replace line
```

## Steps

1. Create plain non-git cwd and valid dep repo.
2. Run `wrk --bring <dep> --no-dep` from the plain cwd.

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
	req.Args = []string{"--bring", dep, "--no-dep"}
	return nil
}
```
