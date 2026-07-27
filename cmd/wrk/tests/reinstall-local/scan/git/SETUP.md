# Scenario

**Feature**: git checkout, useMain=false → scan root is ShowToplevel

```
# workDir inside a git work tree; useMain=false
# scanRoot = worktree.ShowToplevel(workDir) (this checkout, not necessarily main)
workDir (in git) + useMain=false
  -> ResolveReinstallScanRoot -> ShowToplevel
  -> Scan full checkout modules -> multi plan
```

## Preconditions

- Leaves under this branch keep `UseMain=false`.
- Fixtures are real git repos (init + commit) so ShowToplevel succeeds.

## Steps

1. Group forces `UseMain=false` and success expectations.
2. Leaves create multi-module git trees and set WorkDir (often a subdir).

## Context

- Linked-worktree **main** identity is covered under `use-main/`, not here.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseMain = false
	req.WantError = false
	req.WantErrSubstrs = []string{}
	return nil
}
```
