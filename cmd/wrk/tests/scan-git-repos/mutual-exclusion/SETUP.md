# Scenario

**Feature**: --scan-git-repos is mutually exclusive with other standalone modes

```
wrk --scan-git-repos --projects
  -> non-zero exit
  -> stderr mentions mutual exclusion
  -> empty stdout
```

## Preconditions

- No git fixture required for mutex rejection (mode selection fails first).

## Steps

- Descendants combine `--scan-git-repos` with another exclusive mode.

```go
func Setup(t *testing.T, req *Request) error {
	// Error path: mode selection fails before scan; ensure helpers stay linked.
	ensureScanGitReposHelpersUsed()
	return nil
}
```
