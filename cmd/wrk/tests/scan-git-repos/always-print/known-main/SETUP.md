# Scenario

**Feature**: pre-seeded known main is still listed on stdout; projects stay single entry

```
seed projects.json with main (source=scan)
  -> wrk --scan-git-repos scan-root
  -> stdout contains main abs path once
  -> projects.json still exactly one entry source=scan
```

## Steps

1. Create `{WorkRoot}/scan-root/myrepo` as a main git repo.
2. Seed `{WRK_HOME}/projects.json` with that path and `source: "scan"`.
3. Run `wrk --scan-git-repos <scan-root>`.

```go
func Setup(t *testing.T, req *Request) error {
	scanRoot := makeScanRoot(t, req.WorkRoot)
	mainRepo := initScanMainRepo(t, scanRoot, "myrepo")
	seedScanProject(t, req.WrkHome, mainRepo)
	req.MainRepo = mainRepo
	req.Args = []string{"--scan-git-repos", scanRoot}
	return nil
}
```
