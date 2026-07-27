# Scenario

**Feature**: wrk --propagate-tags --dry-run success plan paths

```
# valid git source cwd + registered consumers
source main -> ResolveSourceReleases -> match other projects
  -> stdout plan with source: and would: lines
  -> exit 0; no go.mod/tag/HEAD mutation
```

## Preconditions

- Leaves under this node pass `--propagate-tags --dry-run`.
- Source cwd is the source project's main repo root.
- Source has at least one numeric release tag unless a sibling error leaf.

## Steps

1. Grouping marks all descendants as dry-run plan scenarios.
2. Leaves construct source/consumer fixtures and set `req.Args`.

## Context

- Default args: `[]string{"--propagate-tags", "--dry-run"}`.
- Side-effect asserts use pre-run snapshots from `captureRepoSnapshots`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Descendants override Args/RepoDir after building fixtures.
	propTagsEnsureHelpersUsed()
	if len(req.Args) == 0 {
		req.Args = []string{"--propagate-tags", "--dry-run"}
	}
	return nil
}
```

