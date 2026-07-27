# Scenario

**Feature**: wrk --projects with no recorded projects

```
(no projects) -> wrk --projects -> exit 0, empty stdout
```

## Steps

1. Run `wrk --projects` with empty `projects.json` (file absent).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDetailedStatusHelpersUsed()
	assertNoProjectsFile(t, req.WrkHome)
	return nil
}
```