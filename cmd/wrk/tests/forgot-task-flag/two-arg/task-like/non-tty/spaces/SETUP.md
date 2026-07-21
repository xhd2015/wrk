# Scenario

**Feature**: two-arg multi-word second positional, non-TTY → default auto-promote

```
wrk <myrepo> "fix the login bug" (non-TTY)
  -> auto-promote to --task under WRK_HOME; no hard error
```

## Steps

1. Init `myrepo` on main.
2. Run `wrk <myrepo> "fix the login bug"` from WorkRoot without confirm flags.

```go
func Setup(t *testing.T, req *Request) error {
	setupTwoArg(t, req, taskLikeSpaces)
	return nil
}
```
