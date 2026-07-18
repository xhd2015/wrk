# Scenario

**Feature**: successful wrk --scan-git-repos discovers mains, always prints them, records new ones

```
# explicit root under WorkRoot
wrk --scan-git-repos <scan-root>
  -> scan_repo finds mains under root
  -> stdout always prints valid main abs paths (known or new)
  -> RecordProject(..., source="scan") only when not already in projects.json
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
