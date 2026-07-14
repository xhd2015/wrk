# Scenario

**Feature**: in-place wrk relative/path --cd resolves under process cwd

```
workspace/rel/target exists
cwd=workspace; wrk rel/target --cd + WRK_FOLLOWUP_FILE
  -> follow-up cd <abs workspace/rel/target>
```

## Steps

1. Create `{WorkRoot}/workspace/rel/target`.
2. Run `wrk rel/target --cd` from workspace with channel open.

```go
func Setup(t *testing.T, req *Request) error {
	// Parent in-place Setup already set RepoDir=workspace
	rel := filepath.Join("rel", "target")
	abs := filepath.Join(req.RepoDir, "rel", "target")
	mkdirAll(t, abs)
	req.MainRepo = resolvePath(t, abs)
	setCDPathThenFlag(req, rel)
	return nil
}
```
