# Scenario

**Feature**: module root without go.mod yields an error

```
# empty ModuleRoot (dir exists, no go.mod)
PlanLocalReinstalls(moduleRoot, binDir) -> error
```

## Steps

1. Leave ModuleRoot empty (root Setup already mkdir'd it; no go.mod written).
2. Expect non-nil error from Run.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// ModuleRoot exists but has no go.mod — S9.
	req.WantError = true
	return nil
}
```
