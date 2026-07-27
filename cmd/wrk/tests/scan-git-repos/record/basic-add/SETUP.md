# Scenario

**Feature**: cold first scan always prints main path; does not write projects.json

```
scan-root/myrepo (main, not in projects yet)
  -> wrk --scan-git-repos scan-root
  -> stdout absolute main path (always-print on cold find)
  -> projects.json unchanged (no write)
```

## Steps

1. Create `{WorkRoot}/scan-root/myrepo` as a main git repo.
2. Run `wrk --scan-git-repos <scan-root>` from non-git WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	scanRoot := makeScanRoot(t, req.WorkRoot)
	mainRepo := initScanMainRepo(t, scanRoot, "myrepo")
	req.MainRepo = mainRepo
	req.Args = []string{"--scan-git-repos", scanRoot}
	// RepoDir stays WorkRoot (parent Setup) so auto-record is a no-op.
	return nil
}
```
