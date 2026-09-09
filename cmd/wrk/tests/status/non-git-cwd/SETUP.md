# Scenario

**Feature**: wrk --status scans nested git repos under a plain (non-git) directory

```
# plain directory is not itself a git checkout; status still scans under it
plain cwd -> wrk --status -> discoverStatusRepos(abs(cwd)) -> scan-order blocks
# zero nested repos -> exit 0, empty stdout (no hard error)
```

## Preconditions

- The effective cwd is not inside a git work tree.
- Nested checkouts under cwd (if any) are independent git repositories.

## Steps

- Descendant scenarios run `wrk --status` from a plain directory.
- Nested-repo leaves create sibling git checkouts under that plain root.

## Context

- Presentation matches the non-main git path: scan-order blocks, no `Remote:`,
  no `---- external ----` partition (plain root has no main-repo identity).
- `scan_repo` sorts discovered paths lexicographically by absolute path.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--status"}
	return nil
}
```
