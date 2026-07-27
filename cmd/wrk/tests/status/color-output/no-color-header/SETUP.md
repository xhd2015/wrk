# Scenario

**Feature**: piped --status without --color prints plain external section header

```
# main + nested tools/child -> wrk --status (pipe, no --color)
#   primary main (plain clean)
#   ---- external ----   (no ANSI)
#   nested tools/child
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Commit `.gitignore` with `tools/` so parent porcelain stays clean.
3. Initialize nested independent `{WorkRoot}/myrepo/tools/child`.
4. Run `wrk --status` (no `--color`) from the main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureColorStatusHelpersUsed()
	// Parent color-output SETUP already sets req.Args = ["--status"] (no --color).
	mainRepo, child := setupColorStatusMainPlusNested(t, req.WorkRoot)

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.DepPath = child
	return nil
}
```
