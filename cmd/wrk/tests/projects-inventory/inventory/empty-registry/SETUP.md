# Scenario

**Feature**: empty projects registry yields empty inventory

```
# no projects.json (or empty projects list)
BuildInventory(wrkHome)
  -> Projects=[], Modules=[], CrossEdges=[], IntraEdges=[], SkippedPaths=[]
```

## Steps

1. Leave WRK_HOME empty (no projects.json written).
2. Expect zero projects, modules, edges, and skipped paths.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// WRK_HOME exists from root Setup but projects.json is absent.
	req.WantProjectPaths = []string{}
	req.WantModules = []WantModule{}
	req.WantCrossEdges = []WantEdge{}
	req.WantIntraEdges = []WantEdge{}
	req.WantSkippedPaths = []string{}
	return nil
}
```
