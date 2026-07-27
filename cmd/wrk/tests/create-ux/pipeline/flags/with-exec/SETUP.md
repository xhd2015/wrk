# Scenario

**Feature**: create UX + `--exec` runs exec after UX setup in worktree

```
wrk --new-terminal --exec pwd
  -> create; iterm; stdout: wt\nwt\n (path then pwd)
  # order: create → terminal → exec
```

## Steps

1. Run `--new-terminal --exec pwd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--new-terminal", "--exec", "pwd"}
	return nil
}
```
