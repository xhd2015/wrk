# Scenario

**Feature**: create with no window/terminal/agent axes via `wrk --new`

```
empty config + wrk --new -> native create only; no space/iterm/agent
```

## Steps

- Leaves use empty config and `--new` (no UX flags).
- P1: bare no-args is dashboard; create entry is `--new`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
