# Scenario

**Feature**: useMain=true → scan root is the main repository path

```
# workDir inside a checkout (often a linked worktree); useMain=true
# scanRoot = ResolveMainRepo(ShowToplevel(workDir)) = main repo abs path
workDir + useMain=true
  -> ResolveReinstallScanRoot -> main path
  -> Scan modules under main -> multi plan
```

## Preconditions

- Leaves under this branch set `UseMain=true`.
- Git fixtures include a main repo and a linked worktree.

## Steps

1. Group forces `UseMain=true` and success expectations.
2. Leaves build main + linked worktree and set WorkDir on the linked side.

## Context

- Assert locks path **identity** of ScanRoot to main (not the linked worktree
  path), even when module trees look similar.

```go
func Setup(t *testing.T, req *Request) error {
	req.UseMain = true
	req.WantError = false
	req.WantErrSubstrs = []string{}
	return nil
}
```
