# Scenario

**Feature**: fallback wrk --cd relative/path resolves under cwd

```
workspace/rel/target exists; channel closed; fake bash
cwd=workspace; wrk --cd rel/target
  -> stdout abs; shell cwd = abs
```

## Steps

1. Create relative target under workspace.
2. Install fake bash; run `wrk --cd rel/target`.

```go
func Setup(t *testing.T, req *Request) error {
	rel := filepath.Join("rel", "target")
	abs := filepath.Join(req.RepoDir, "rel", "target")
	mkdirAll(t, abs)
	req.MainRepo = resolvePath(t, abs)
	installFakeBash(t, req, 0)
	setCDFlagThenPath(req, rel)
	return nil
}
```
