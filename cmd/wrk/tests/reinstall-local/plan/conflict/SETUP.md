# Scenario

**Feature**: cmd×script merge, prefer-script notice, and ambiguity skip/fallback

```
# unique cmd + unique script → script item + notice prefer-script
# ambiguous cmd/script → drop that tree; unique other side falls back
# both ambiguous / ambiguous alone → omit bin from Items + warning(s)
./cmd/... + ./script/.../install
  -> PlanLocalReinstalls
  -> Items (survivors) + Diagnostics
```

## Preconditions

- Leaves under this branch create discovery fixtures that share BinNames across
  cmd and/or script trees (unique or ambiguous).
- Bin stubs present when Action=install is expected for a survivor.

## Steps

1. Leaves write cmd and/or script package mains and bin stubs as needed.
2. Assert Items (possibly empty for pure ambiguity) and Diagnostics.

## Context

- Prefer-script notice only when **both** sides are unique.
- Ambiguous drop is not a binDir skip row — the bin is omitted from Items.
- Diagnostic Paths are sorted slash-form `./…` relative paths.

```go
func Setup(t *testing.T, req *Request) error {
	if req.WantItems == nil {
		req.WantItems = []WantPlanItem{}
	}
	if req.WantDiagnostics == nil {
		req.WantDiagnostics = []WantDiagnostic{}
	}
	return nil
}
```
