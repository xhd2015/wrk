# Scenario

**Feature**: same bin name from cmd and script → script wins (single item)

```
# dedup: script go-run-install replaces cmd go-install for that BinName
./cmd/foo + ./script/foo/install
  -> one PlanItem Method=go-run-install RelPath=./script/foo/install
```

## Preconditions

- Leaves under this branch create both a cmd main and a script install main
  that would share a bin name.

## Steps

1. Leaves write both discovery sources and bin stubs as needed.
2. Assert a single plan item for the contested name.

## Context

- No dual listing; order rules still apply if other bins exist.

```go
func Setup(t *testing.T, req *Request) error {
	if req.WantItems == nil {
		req.WantItems = []WantPlanItem{}
	}
	return nil
}
```
