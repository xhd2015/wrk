# Scenario

**Feature**: create with positionals / task flags still works without requiring `--new`

```
# product lock: only pure no-args becomes dashboard
wrk <dir>          -> create (no --new required)
wrk -t <task>      -> create (no --new required)
```

## Steps

- Leaves exercise create-with-args regressions that must stay green after bare no-args changes.
- Grouping only skips if git is missing; leaves set TargetDir or TaskDesc.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	// Must not inject --new: these leaves prove create without --new still works.
	req.Args = nil
	return nil
}
```
