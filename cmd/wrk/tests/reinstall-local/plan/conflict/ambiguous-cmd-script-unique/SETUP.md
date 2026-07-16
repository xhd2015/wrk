# Scenario

**Feature**: ambiguous cmd + unique script → script item; cmd warning only (no prefer-script)

```
# S4e: ./cmd/foo + ./cmd/nested/foo + ./script/foo/install; bin present
PlanLocalReinstalls
  -> Items=[{BinName:foo, Method:go-run-install, RelPath:./script/foo/install, Action:install}]
  -> Diagnostics=[{Level:warning, Kind:ambiguous-cmd, BinName:foo,
       Paths:[./cmd/foo, ./cmd/nested/foo]}]
  # no prefer-script notice (cmd side not unique)
```

## Steps

1. Write `go.mod` with module `example.com/amb-cmd-script`.
2. Write two cmd mains sharing bin `foo` and one unique script install.
3. Touch `$binDir/foo`.
4. Expect script survivor + ambiguous-cmd warning only.

```go
func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/amb-cmd-script")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "nested", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	touchBin(t, req.BinDir, "foo")

	req.WantModuleName = "amb-cmd-script"
	req.WantItems = []WantPlanItem{
		{
			BinName: "foo",
			Method:  methodGoRunInstall,
			RelPath: "./script/foo/install",
			Action:  actionInstall,
		},
	}
	req.WantDiagnostics = []WantDiagnostic{
		{
			Level:   diagLevelWarning,
			Kind:    diagKindAmbCmd,
			BinName: "foo",
			Paths:   []string{"./cmd/foo", "./cmd/nested/foo"},
		},
	}
	return nil
}
```
