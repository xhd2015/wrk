# Scenario

**Feature**: PlanLocalReinstallsMulti hard-errors on cross-module install claims

```
# same BinName Action=install from two modules → error, no multi plan
moduleA + moduleB both claim bin with install
  -> PlanLocalReinstallsMulti
  -> error (names bin + both modules)
```

## Preconditions

- Leaves under this branch set `WantError = true` and fill `WantErrSubstrs`.
- Both modules produce the same BinName with Action=install (bin file present).

## Steps

1. Leaves arrange two module roots that both install-claim the same bin.
2. Assert Run returns a non-nil error identifying the bin and both modules.

## Context

- Skip×install or skip×skip for the same bin must not take this branch
  (see plan/skip-dup).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.WantError = true
	req.WantModules = []WantModulePlan{}
	return nil
}
```
