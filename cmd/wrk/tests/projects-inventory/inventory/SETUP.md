# Scenario

**Feature**: BuildInventory loads registry, scans modules, classifies require edges

```
# Op = inventory
WRK_HOME/projects.json
  -> wrkcli.BuildInventory(wrkHome)
  -> Inventory{Projects, SkippedPaths}
  -> CrossEdges() / IntraEdges()
```

## Preconditions

- Leaves under this branch set `req.Op = OpInventory`.
- `SourceMain` is unused.
- Fixtures register zero or more project paths and real go.mod trees.

## Steps

1. Set operation to inventory.
2. Leaves seed projects.json + repos, then set Want* expectations.
3. Run calls BuildInventory and edge methods.

## Context

- Soft-skip missing paths does not fail BuildInventory.
- Cross vs intra classification requires owned modules registered in the same inventory.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = OpInventory
	return nil
}
```
