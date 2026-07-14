# Scenario

**Feature**: bare create with empty config runs native worktree only

```
myrepo -> wrk
  -> wt path on stdout
  -> no space, no iterm, no agent-run
```

## Steps

1. Init myrepo; install mocks (to prove they stay silent).
2. Run bare `wrk`.

```go
func Setup(t *testing.T, req *Request) error {
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	req.Args = nil
	return nil
}
```
