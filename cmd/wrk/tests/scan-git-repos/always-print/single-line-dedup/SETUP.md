# Scenario

**Feature**: single main under root yields exactly one stdout line for that path (in-run dedup)

```
scan-root/myrepo (one main, cold)
  -> wrk --scan-git-repos scan-root
  -> stdout path-line count for main == 1
  -> projects.json remains empty (print-only)
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
	return nil
}
```
