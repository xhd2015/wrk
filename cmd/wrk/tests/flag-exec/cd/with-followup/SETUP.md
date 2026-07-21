# Scenario

**Feature**: in-place `--cd` + `--exec pwd` writes follow-up and prints pwd

```
WRK_FOLLOWUP_FILE=tmp; workspace/
wrk --cd {WorkRoot}/jumpto --exec pwd
  -> follow-up: cd {WorkRoot}/jumpto\n
  -> stdout: {WorkRoot}/jumpto\n
  -> exit 0
```

## Steps

1. Create absolute target `{WorkRoot}/jumpto`.
2. Open follow-up channel.
3. Run `wrk --cd <abs> --exec pwd` from a neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	target := execCDTarget(t, req, "jumpto")
	req.MainRepo = target // reuse MainRepo as "resolved jump path" for asserts
	enableExecFollowup(t, req)
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--cd", target, "--exec", "pwd"}
	return nil
}
```
