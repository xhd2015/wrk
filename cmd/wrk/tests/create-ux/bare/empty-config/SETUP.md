# Scenario

**Feature**: `wrk --new` with empty config runs native worktree only

```
myrepo -> wrk --new
  -> wt path on stdout
  -> no space, no iterm, no agent-run
```

## Steps

1. Init myrepo; install mocks (to prove they stay silent).
2. Run `wrk --new` (P1 create entry; bare no-args is no longer create).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	req.Args = []string{"--new"}
	return nil
}
```
