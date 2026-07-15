# Scenario

**Feature**: scan root with one main repo records source=scan and prints path

```
scan-root/myrepo (main)
  -> wrk --scan-git-repos scan-root
  -> projects.json entry source=scan
  -> stdout absolute main path
```

## Steps

1. Create `{WorkRoot}/scan-root/myrepo` as a main git repo.
2. Run `wrk --scan-git-repos <scan-root>` from non-git WorkRoot.

```go
func Setup(t *testing.T, req *Request) error {
	scanRoot := makeScanRoot(t, req.WorkRoot)
	mainRepo := initScanMainRepo(t, scanRoot, "myrepo")
	req.MainRepo = mainRepo
	req.Args = []string{"--scan-git-repos", scanRoot}
	// RepoDir stays WorkRoot (parent Setup) so auto-record is a no-op.
	return nil
}
```
