# Scenario

**Feature**: nested package main under ./cmd/... uses last path segment as bin

```
# ./cmd/nested/tool package main → BinName=tool, RelPath=./cmd/nested/tool
./cmd/<...>/<leaf>
  -> go-install PlanItem
```

## Preconditions

- Leaves create multi-segment paths under `cmd/` with package main at the leaf.

## Steps

1. Leaves write nested cmd package mains and bin stubs.
2. Assert bin name is only the last segment.

## Context

- Intermediate directories need not be package main.

```go
func Setup(t *testing.T, req *Request) error {
	if req.WantItems == nil {
		req.WantItems = []WantPlanItem{}
	}
	return nil
}
```
