# Scenario

**Feature**: successful wrk --scan-git-repos discovers mains and records them

```
# explicit root under WorkRoot
wrk --scan-git-repos <scan-root>
  -> scan_repo finds mains under root
  -> RecordProject(..., source="scan")
  -> stdout newly recorded abs paths
```

## Preconditions

- Fixtures use an explicit scan root (`{WorkRoot}/scan-root/...`), not `$HOME/Projects`.
- Cwd remains non-git `{WorkRoot}`.

## Steps

- Descendants place one or more git checkouts under the scan root and set Args.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: success-path record leaves share scan helpers.
	ensureScanGitReposHelpersUsed()
	return nil
}
```
