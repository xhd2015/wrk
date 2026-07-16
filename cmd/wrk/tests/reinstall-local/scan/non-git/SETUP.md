# Scenario

**Feature**: non-git workDir → walk-up go.mod as scan root

```
# directory is not inside a git work tree; useMain=false
# scanRoot = nearest ancestor (or self) with go.mod
workDir (non-git) + useMain=false
  -> ResolveReinstallScanRoot -> walk-up go.mod dir
  -> Scan that root (typically one module) -> multi plan
```

## Preconditions

- Leaves under this branch do **not** run `git init` on the fixture tree.
- `UseMain` stays false (useMain without git is not this branch's success path).

## Steps

1. Group forces non-git success defaults.
2. Leaves create plain directories with go.mod and set WorkDir.

## Context

- Error when no go.mod is found lives under `error/no-go-mod`.

```go
func Setup(t *testing.T, req *Request) error {
	req.UseMain = false
	req.WantError = false
	req.WantErrSubstrs = []string{}
	return nil
}
```
