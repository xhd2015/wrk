# Scenario

**Feature**: wrk --bring from a plain non-git cwd materializes external worktree and soft-skips replace

```
# plain dir + abs dep path -> wrk --bring <dep>
#   -> exit 0; {plain}/external/mydep-main-{date}
#   -> SKIP … is not a git repository; no .gitignore
plain (no .git) + mydep (module example.com/dep)
  -> wrk --bring <abs-dep>
  -> stdout external path under plain/external/
```

## Steps

1. Create plain non-git directory as cwd.
2. Create dep git repo `mydep` with go.mod.
3. Run `wrk --bring <dep>` with cwd = plain dir.

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
	req.Args = []string{"--bring", dep}
	return nil
}
```
