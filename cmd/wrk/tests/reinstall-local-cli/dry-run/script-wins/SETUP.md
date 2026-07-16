# Scenario

**Feature**: dry-run prefers script install over cmd for the same bin name

```
# C3: ./cmd/foo + ./script/foo/install + GOBIN/foo
mod/ -> wrk --reinstall-local --dry-run
  -> would: go run ./script/foo/install
  -> would: reinstall 1 binaries (0 skipped)
  # must NOT print would: go install ./cmd/foo
```

## Steps

1. Write `go.mod` with module `example.com/cli-script-wins`.
2. Write `./cmd/foo` as `package main`.
3. Write `./script/foo/install` as `package main`.
4. Touch `$GOBIN/foo`.
5. Run `wrk --reinstall-local --dry-run` from module root.

```go
func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cli-script-wins")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	touchBin(t, req.BinDir, "foo")
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
