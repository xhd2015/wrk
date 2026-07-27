# Scenario

**Feature**: `--new-window` at macOS max Desktops is best-effort (warn + continue)

```
WRK_SPACE_FAIL=max-desktops; wrk --new-window
  -> space CreateAndActivate attempted once
  -> stderr warning: maximum Desktops / continuing on current Desktop
  -> worktree still created; path on stdout; exit 0
  -> iterm ForceNew still runs (current Desktop)
```

## Steps

1. Install mocks with `WRK_SPACE_FAIL=max-desktops`.
2. Run `wrk --new-window`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--new-window"}
	req.ExtraEnv = append(req.ExtraEnv, "WRK_SPACE_FAIL=max-desktops")
	return nil
}
```
