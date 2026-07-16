# Scenario

**Feature**: cmd and script both produce bin foo → only script remains

```
# S4: ./cmd/foo and ./script/foo/install, $binDir/foo present
PlanLocalReinstalls
  -> Items=[{BinName:foo, Method:go-run-install, RelPath:./script/foo/install, Action:install}]
```

## Steps

1. Write `go.mod` with module `example.com/conflict-foo`.
2. Write `./cmd/foo` as `package main`.
3. Write `./script/foo/install` as `package main`.
4. Touch `$binDir/foo`.
5. Expect exactly one item: script path, not cmd.

```go
func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/conflict-foo")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	touchBin(t, req.BinDir, "foo")

	req.WantModuleName = "conflict-foo"
	req.WantItems = []WantPlanItem{
		{
			BinName: "foo",
			Method:  methodGoRunInstall,
			RelPath: "./script/foo/install",
			Action:  actionInstall,
		},
	}
	return nil
}
```
