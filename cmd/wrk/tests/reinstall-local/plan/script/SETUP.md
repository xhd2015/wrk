# Scenario

**Feature**: discovery from ./script/.../install package main only

```
# directory named install that is package main → go-run-install
# nested: bin = parent name; bare ./script/install → ModuleName
./script/.../install package main
  -> PlanItem{Method: go-run-install, RelPath: ./script/.../install}
```

## Preconditions

- Leaves place package mains under `script/` install dirs only (no `cmd/` mains
  unless a leaf documents otherwise).

## Steps

1. Leaves write `go.mod` and script install package mains.
2. Touch matching bins when expecting install action.

## Context

- Method is always `go-run-install` for script-derived items.

```go
func Setup(t *testing.T, req *Request) error {
	if req.WantItems == nil {
		req.WantItems = []WantPlanItem{}
	}
	// Grouping marker: script-only discovery branch.
	return nil
}
```
