# Scenario

**Feature**: wrk --where spl with no saved project errors

```
no saved spl
workspace/ -> wrk --where spl -> non-zero, stderr no-match, empty stdout
```

## Steps

1. Use neutral cwd with empty `projects.json` (no `--add` for `spl`).
2. Run `wrk --where spl`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = whereArgs(whereBasename)
	return nil
}```
