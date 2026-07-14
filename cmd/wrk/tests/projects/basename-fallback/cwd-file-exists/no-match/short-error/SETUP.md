# Scenario

**Feature**: wrk foo emits short file-collision error when no saved project matches

```
workspace/foo (file), empty projects.json match for foo
wrk foo -> single stderr line only
```

## Steps

1. Create neutral cwd `{WorkRoot}/workspace` with regular file `foo`.
2. Leave `projects.json` without a `foo` basename entry.
3. Run `wrk foo` from workspace.

```go
func Setup(t *testing.T, req *Request) error {
	cwd := initNeutralCwd(t, req.WorkRoot, "workspace")
	initBasenameFile(t, cwd, "foo", "")
	req.RepoDir = cwd
	req.TargetDir = "foo"
	return nil
}
```