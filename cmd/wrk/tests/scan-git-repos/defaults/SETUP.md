# Scenario

**Feature**: bare `wrk --scan-git-repos` (no ROOT...) uses `$HOME` as the default scan root

```
# no positional roots → default root is UserHomeDir() (~), not ~/Projects
wrk --scan-git-repos
  -> roots = [$HOME] when home is a directory
  -> scan_repo finds mains under home
  -> RecordProject(..., source="scan") + stdout abs paths

# unusable home
wrk --scan-git-repos  (HOME missing or not a directory)
  -> non-zero; clear error about home / ~ (not require Projects)
```

## Preconditions

- Isolates `HOME` via `Request.FakeHome` (wrkEnv sets `HOME=`).
- Fixtures must **not** create `{FakeHome}/Projects` so a Projects-only default cannot pass by accident.
- Cwd remains non-git `{WorkRoot}` (parent Setup).

## Steps

- Descendants set `FakeHome`, optional fixtures under home, and `Args: --scan-git-repos` with no roots.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: bare-flag default-root leaves share scan helpers.
	ensureScanGitReposHelpersUsed()
	return nil
}
```
