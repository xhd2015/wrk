# Scenario

**Feature**: bare ./script/install uses module path basename as bin name

```
# S5: module example.com/demo + ./script/install + $binDir/demo
PlanLocalReinstalls
  -> Items=[{BinName:demo, Method:go-run-install, RelPath:./script/install, Action:install}]
```

## Steps

1. Write `go.mod` with module `example.com/demo` (basename `demo`).
2. Write bare `./script/install` as `package main`.
3. Touch `$binDir/demo` (not `install` / not `script`).
4. Expect one go-run-install item with BinName `demo`.

```go
func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/demo")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "install"))
	touchBin(t, req.BinDir, "demo")

	req.WantModuleName = "demo"
	req.WantItems = []WantPlanItem{
		{
			BinName: "demo",
			Method:  methodGoRunInstall,
			RelPath: "./script/install",
			Action:  actionInstall,
		},
	}
	return nil
}
```
