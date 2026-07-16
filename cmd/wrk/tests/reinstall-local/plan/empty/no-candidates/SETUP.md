# Scenario

**Feature**: module with go.mod but no cmd/script mains → empty install plan

```
# S8: only go.mod (no ./cmd, no ./script)
PlanLocalReinstalls
  -> Items=[], ModuleName set, err=nil
```

## Steps

1. Write `go.mod` with module `example.com/empty-mod`.
2. Do not create `cmd/` or `script/` trees.
3. Expect empty Items, ModuleName `empty-mod`.

```go
func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/empty-mod")

	req.WantModuleName = "empty-mod"
	req.WantItems = []WantPlanItem{}
	return nil
}
```
