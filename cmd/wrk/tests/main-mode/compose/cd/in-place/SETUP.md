# Scenario

**Feature**: wrk --main --cd with WRK_FOLLOWUP_FILE writes in-place follow-up (no shell)

```
WRK_FOLLOWUP_FILE set
wrk --main --cd -> empty stdout; follow-up cd <main>\n; exit 0
```

## Preconditions

- Channel open via `enableFollowupChannel`.
- No fake shell required for pure in-place (optional for accidental-launch detection on already-at-main).

## Steps

1. Open follow-up channel.
2. Descendants set cwd layout and Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	enableFollowupChannel(t, req)
	return nil
}
```
