# Scenario

**Feature**: successful wrk --scan-git-repos discovers mains and always prints them (no registry write)

```
# explicit root under WorkRoot
wrk --scan-git-repos <scan-root>
  -> scan_repo finds mains under root
  -> stdout always prints valid main abs paths (known or new)
  -> projects.json is never written
```

## Preconditions

- Fixtures use an explicit scan root (`{WorkRoot}/scan-root/...`), not `$HOME/Projects`.
- Cwd remains non-git `{WorkRoot}`.

## Steps

- Descendants place one or more git checkouts under the scan root and set Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: success-path record leaves share scan helpers.
	ensureScanGitReposHelpersUsed()
	return nil
}
```
