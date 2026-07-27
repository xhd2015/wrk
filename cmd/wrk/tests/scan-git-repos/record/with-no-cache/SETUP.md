# Scenario

**Feature**: --scan-git-repos --no-cache still discovers and records mains

```
wrk --scan-git-repos --no-cache <scan-root>
  -> scan_repo with NoCache=true
  -> still prints main; projects.json empty
  -> stdout main path
```

## Steps

1. Create `{WorkRoot}/scan-root/myrepo` as a main git repo.
2. Run `wrk --scan-git-repos --no-cache <scan-root>`.

```go
func Setup(t *testing.T, req *Request) error {
	scanRoot := makeScanRoot(t, req.WorkRoot)
	mainRepo := initScanMainRepo(t, scanRoot, "myrepo")
	req.MainRepo = mainRepo
	req.Args = []string{"--scan-git-repos", "--no-cache", scanRoot}
	return nil
}
```
