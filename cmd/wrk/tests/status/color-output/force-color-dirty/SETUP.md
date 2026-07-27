# Scenario

**Feature**: --color highlights dirty status with granular red/grey segments

```
dirty main repo (1 added, 1 changed) -> wrk --status --color -> red dirty + non-zero counts, grey zeros
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Stage one new file and modify one tracked file.
3. Run `wrk --status --color` from the main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureColorStatusHelpersUsed()
	withStatusColor(req)
	mainRepo := setupColorStatusMainRepo(t, req.WorkRoot, "myrepo", "dirty status base")
	dirtyColorStatusRepo(t, mainRepo)

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	return nil
}
```