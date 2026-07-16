# Scenario

**Feature**: wrk --propagate-tags (no --dry-run) applies consumer go.mod updates, builds, and commits

```
# valid git source cwd + registered consumers
source main -> ResolveSourceReleases -> match other projects
  -> drop cross-project local replace (when present)
  -> go mod edit -require to release version
  -> go mod tidy per updated consumer module
  -> go build ./... per updated module (P5)
  -> if all modules in project build OK: one commit (go.mod/go.sum) with chore(deps):
  -> if build fails: warning: on stderr; dirty tree; no commit; exit 0
  -> stdout uses updated / dropped replace / go build ok / committed (not would:)
```

## Preconditions

- Leaves under this node pass `--propagate-tags` **without** `--dry-run`.
- Source cwd is the source project's main repo root.
- Source has at least one numeric release tag.
- Apply leaves may seed `{WorkRoot}/modproxy` and set `GOPROXY=file://…` so tidy
  resolves synthetic module versions offline.
- Classic TDD P5: production currently stops after tidy (no build/commit) → leaves
  that expect `go build ./... ok` / `committed` stay RED until implementer lands P5.

## Steps

1. Grouping marks all descendants as apply (mutate) scenarios.
2. Leaves construct source/consumer fixtures, optional module proxy, and set `req.Args`.

## Context

- Default args: `[]string{"--propagate-tags"}`.
- Side-effect asserts: consumer go.mod changes on bumps; source go.mod / tags / HEAD
  must not change; consumer HEAD advances only when build succeeds and commit runs;
  build-fail and already-current leave consumer HEAD unchanged.

```go
func Setup(t *testing.T, req *Request) error {
	propTagsEnsureHelpersUsed()
	if len(req.Args) == 0 {
		req.Args = []string{"--propagate-tags"}
	}
	return nil
}
```
