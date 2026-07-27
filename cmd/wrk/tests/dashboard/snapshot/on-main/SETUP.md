# Scenario

**Feature**: non-TTY bare `wrk` from main checkout still shows dashboard snapshot

```
main-repo
  -> wrk (non-TTY, no args)
  -> dashboard snapshot (not create)
  -> DONE / MERGE BACK may be [-] or show a disabled reason (soft)
  -> no create-hint; Batch would-run present
```

## Steps

1. Main repo cwd.
2. Bare `wrk`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	main := setupDashboardMainRepo(t, req)
	req.RepoDir = main
	req.Args = nil
	return nil
}
```
