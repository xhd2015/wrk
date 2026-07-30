# Scenario

**Feature**: wrk --main --where prints the absolute main repository path of this checkout

```
# resolve main from cwd; print; never shell / follow-up
wrk --main --where
  -> stdout: abs(mainRepo)\n
  -> stderr empty on success
  -> always print (including when already at main root)

# event
command="where"; args include --main and --where
```

## Preconditions

- No positional basename; main is resolved from process cwd only.
- Fake bash not required for success leaves (must not launch shell; install only if detecting accidental launch).

## Steps

1. Descendants set cwd (linked wt / main subdir / main root / non-git).
2. Args default `--main --where` unless leaf overrides order or extras.

## Context

- Not projects.json lookup.
- Already-at-main: still print path; **no** bare-main stderr notice.
- `--exec` is rejected with `--where` (same as bare `--where`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setMainWhereArgs(req, "--main", "--where")
	return nil
}
```
