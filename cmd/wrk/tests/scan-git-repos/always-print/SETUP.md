# Scenario

**Feature**: --scan-git-repos always prints valid mains regardless of projects.json known-ness

```
# warm / pre-seeded projects does not gate stdout
optional: scan-root/myrepo pre-seeded in projects.json (legacy source=scan)
  -> wrk --scan-git-repos scan-root
  -> stdout includes main abs path exactly once
  -> projects count stays 1

# cold single-main: exactly one stdout line for the path (in-run dedup)
scan-root/myrepo (new)
  -> wrk --scan-git-repos scan-root
  -> count of main path lines on stdout == 1
```

## Preconditions

- Explicit scan root under `{WorkRoot}`; cwd remains non-git `{WorkRoot}`.
- Known-main leaves seed `projects.json` via `seedScanProject` (do not pre-call the feature).

## Steps

- Descendants init one main under scan-root and set Args; known-main also seeds projects.

```go
func Setup(t *testing.T, req *Request) error {
	ensureScanGitReposHelpersUsed()
	return nil
}
```
