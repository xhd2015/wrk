# Scenario

**Feature**: empty moduleRoots → empty Modules, nil error (M5)

```
# M5: no module roots provided
PlanLocalReinstallsMulti([], binDir)
  -> {BinDir, Modules:[]}, err=nil
```

## Steps

1. Do not create any module directories under WorkRoot.
2. Keep `ModuleRoots` empty.
3. Expect empty `WantModules`, `WantError=false`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ModuleRoots = []string{}
	req.WantModules = []WantModulePlan{}
	return nil
}
```
