# Scenario

**Feature**: wrk --projects-dep-graph success paths (registry topologies)

```
# exclusive success mode
workspace/ (non-git) + WRK_HOME projects.json
  -> wrk --projects-dep-graph
  -> stdout human graph + footer (exit 0)
  -> stderr only soft warnings when paths missing
```

## Steps

1. Grouping sets the mode flag; leaves seed fixtures and refine expectations.
2. All graph leaves share `req.Args = []string{"--projects-dep-graph"}`.

## Context

- Process cwd is neutral non-git `workspace/` (no git required).
- Implementation uses `BuildInventory` + `CrossEdges` under the hood.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--projects-dep-graph"}
	depGraphEnsureHelpersUsed()
	return nil
}
```
