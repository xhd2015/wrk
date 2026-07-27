# Scenario

**Feature**: --color with --list is accepted; list output stays plain (color no-op outside --projects)

```
git repo cwd -> wrk --list --color -> same stdout as git worktree list, no ANSI
```

## Steps

1. Initialize git repo `{WorkRoot}/listcolor`.
2. Run `wrk --list --color` from repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureColorOutputHelpersUsed()
	repo := initProjectsRepo(t, req.WorkRoot, "listcolor")
	req.Args = []string{"--list", "--color"}
	req.RepoDir = repo
	return nil
}
```