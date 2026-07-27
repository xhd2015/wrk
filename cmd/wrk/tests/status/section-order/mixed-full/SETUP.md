# Scenario

**Feature**: main + in-tree linked + out-of-tree linked + nested → primary three, header, nested

```
# full mixed fixture
myrepo
  + wrk out-of-tree wt
  + wt-linked (in-tree)
  + tools/child (nested independent)
wrk --status
  -> primary: main + ListLinked (two linked)
  -> ---- external ----
  -> tools/child
```

## Steps

1. Create main + WRK external worktree.
2. Add in-tree linked worktree `wt-linked` on branch `wt-side`.
3. Commit `.gitignore` with `tools/` and create nested `tools/child`.
4. Run `wrk --status` from main root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, _, _ := setupWrkWorktreeFromMain(t, req)
	inTree := addInTreeLinkedWorktree(t, mainRepo, "wt-linked", "wt-side")
	ensureToolsGitignore(t, mainRepo, "tools/")
	child := initNestedIndependentRepo(t, mainRepo, "tools/child", "nested child")

	req.RepoDir = mainRepo
	// Reuse Request fields: DepsLinkedWtDir = in-tree path; Wt2Branch = in-tree branch.
	req.DepsLinkedWtDir = inTree
	req.Wt2Branch = "wt-side"
	req.DepPath = child
	return nil
}
```
