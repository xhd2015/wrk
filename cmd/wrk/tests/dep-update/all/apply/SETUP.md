# Scenario

**Feature**: wrk --dep-update --all applies inventory pins and tidies affected modules

```
consumer git toplevel + inventory owners + optional file:// GOPROXY
  -> wrk --dep-update --all
  -> dep-update lines; go mod tidy ok; summary
  -> only consumer-tree go.mod/go.sum change; no commit
```

## Preconditions

- Leaves under this node pass `--dep-update --all` **without** `--dry-run`.
- Apply leaves seed modproxy when tidy must resolve synthetic versions.

## Steps

1. Grouping marks apply (mutate) scenarios.
2. Leaves construct fixtures + optional proxy and set Args.

## Context

- Default args: `[]string{"--dep-update", "--all"}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	if len(req.Args) == 0 {
		req.Args = []string{"--dep-update", "--all"}
	}
	ensureDepUpdateHelpersUsed()
	return nil
}
```
