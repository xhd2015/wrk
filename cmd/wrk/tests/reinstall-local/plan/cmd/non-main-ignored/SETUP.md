# Scenario

**Feature**: non-main package under cmd is not a reinstall candidate

```
# S10: ./cmd/lib is package lib (not main) → no plan items
PlanLocalReinstalls
  -> Items=[]
```

## Steps

1. Write `go.mod` with module `example.com/cmd-non-main`.
2. Write `./cmd/lib` as `package lib` (not main).
3. Touch `$binDir/lib` (should not matter — not a candidate).
4. Expect empty Items.

```go
func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cmd-non-main")
	writePackageNamed(t, filepath.Join(req.ModuleRoot, "cmd", "lib"), "lib")
	touchBin(t, req.BinDir, "lib")

	req.WantModuleName = "cmd-non-main"
	req.WantItems = []WantPlanItem{}
	return nil
}
```
