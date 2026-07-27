# Scenario

**Feature**: tokens after `--exec` are not parsed as wrk flags

```
myrepo -> wrk --exec echo --task
  -> create wt; child argv = [echo, --task]
  -> stdout: <wt>\n--task\n
  # --task is NOT a wrk --task flag
```

## Steps

1. Initialize git repo `myrepo` on `main`.
2. Run `wrk --exec echo --task` from the main checkout.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initGitRepoOnMain(t, req.RepoDir)
	req.Args = []string{"--exec", "echo", "--task"}
	return nil
}
```
