# Scenario

**Feature**: wrk myrepo --status with no match reports does not exist

```
workspace/ (cwd) -> wrk myrepo --status -> wrk: <abs>/myrepo does not exist
```

## Steps

1. Use neutral cwd `{WorkRoot}/workspace` (no `./myrepo`).
2. Leave `projects.json` empty.
3. Run `wrk myrepo --status`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "myrepo"
	return nil
}
```
