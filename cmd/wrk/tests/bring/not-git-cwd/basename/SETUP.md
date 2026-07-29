# Scenario

**Feature**: wrk --bring mydep from non-git cwd resolves basename via projects.json

```
# saved mydep in projects.json; plain cwd has no ./mydep
#   -> wrk --bring mydep -> external under plain/external/; SKIP non-git
plain (no .git) + saved/mydep (registered)
  -> wrk --bring mydep
  -> stdout external path; basename fallback to registered dep
```

## Steps

1. Create plain non-git directory as cwd.
2. Create dep git repo at `{WorkRoot}/mydep` with go.mod.
3. Register dep with `wrk --add`.
4. Run `wrk --bring mydep` from the plain cwd (no local `./mydep`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	plain := initBringPlainCwd(t, req.WorkRoot, "plain")
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)
	recordBringSavedProject(t, req, dep)

	req.RepoDir = plain
	req.ConsumerTop = plain
	req.DepPath = dep
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", "mydep"}
	return nil
}
```
