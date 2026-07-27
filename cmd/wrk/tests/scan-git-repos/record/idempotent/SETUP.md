# Scenario

**Feature**: second --scan-git-repos does not duplicate projects.json; still prints the valid main

```
seed projects.json source=scan; wrk --scan-git-repos ROOT -> registry unchanged
wrk --scan-git-repos ROOT (2nd under test)
  -> exit 0
  -> still one projects entry
  -> stdout contains the main abs path once (always-print; not gated on known)
```

## Steps

1. Create `{WorkRoot}/scan-root/myrepo` as a main git repo.
2. Seed `{WRK_HOME}/projects.json` with that main path and `source: "scan"` (simulates a prior successful scan without calling the feature under test).
3. Run `wrk --scan-git-repos <scan-root>` as the leaf under test.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	scanRoot := makeScanRoot(t, req.WorkRoot)
	mainRepo := initScanMainRepo(t, scanRoot, "myrepo")
	seedScanProject(t, req.WrkHome, mainRepo)
	req.MainRepo = mainRepo
	req.Args = []string{"--scan-git-repos", scanRoot}
	return nil
}
```
