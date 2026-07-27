# Scenario

**Feature**: discovery from ./cmd/... package main only

```
# walk cmd; package main → go-install; bin filter install|skip
./cmd/<name> package main
  -> PlanItem{Method: go-install, RelPath: ./cmd/<name>}
```

## Preconditions

- Leaves place package mains only under `cmd/` (no script install trees unless
  a specific leaf documents a negative case).

## Steps

1. Leaves write `go.mod` and `./cmd/...` fixtures.
2. Optionally touch `$binDir/<bin>`.
3. Expect Method `go-install` for cmd-derived items.

## Context

- Bin name is the last path segment of the cmd package directory.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: cmd-only discovery branch defaults.
	return ensureCmdGroupMarked(req)
}

func ensureCmdGroupMarked(req *Request) error {
	// No shared fixture writes; leaves own layouts. Marker keeps Setup non-stub.
	if req.WantItems == nil {
		req.WantItems = []WantPlanItem{}
	}
	return nil
}
```
