# Scenario

**Feature**: multi-module execute install×install collision fails before any go install

```
# Exit criteria: collision fails before install (execute path, no --dry-run)
mod/
  go.mod (parent scan root)
  mod-a/go.mod + cmd/samebin
  mod-b/go.mod + cmd/samebin
  GOBIN/samebin present
  -> wrk --reinstall-local
  -> non-zero; stderr names bin samebin (and identifies both modules)
  -> no go install; stub unchanged
```

## Steps

1. Write parent module `example.com/cli-exec-coll-parent` at ModuleRoot (scan root).
2. Write nested `mod-a` module `example.com/cli-exec-coll-a` with `./cmd/samebin`.
3. Write nested `mod-b` module `example.com/cli-exec-coll-b` with `./cmd/samebin`.
4. Touch `$GOBIN/samebin` so both nested modules would Action=install.
5. Run `wrk --reinstall-local` (no `--dry-run`) from ModuleRoot.
6. Expect non-zero exit at plan time; no install mutation of the stub.

```go
func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cli-exec-coll-parent")

	modA := filepath.Join(req.ModuleRoot, "mod-a")
	writeGoMod(t, modA, "example.com/cli-exec-coll-a")
	writePackageMain(t, filepath.Join(modA, "cmd", "samebin"))

	modB := filepath.Join(req.ModuleRoot, "mod-b")
	writeGoMod(t, modB, "example.com/cli-exec-coll-b")
	writePackageMain(t, filepath.Join(modB, "cmd", "samebin"))

	touchBin(t, req.BinDir, "samebin")

	req.Args = []string{"--reinstall-local"}
	return nil
}
```
