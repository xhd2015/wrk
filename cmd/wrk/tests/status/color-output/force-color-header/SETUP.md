# Scenario

**Feature**: --color wraps external section header in gray ANSI

```
# main + nested tools/child -> wrk --status --color
#   primary main (green clean)
#   gray ---- external ----
#   nested tools/child (green clean)
```
## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Commit `.gitignore` with `tools/` so parent porcelain stays clean.
3. Initialize nested independent `{WorkRoot}/myrepo/tools/child`.
4. Run `wrk --status --color` from the main repo root (pipe-safe force color).

```go
func Setup(t *testing.T, req *Request) error {
	ensureColorStatusHelpersUsed()
	withStatusColor(req)
	mainRepo, child := setupColorStatusMainPlusNested(t, req.WorkRoot)

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.DepPath = child
	return nil
}
```
