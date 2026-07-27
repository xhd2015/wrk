# Scenario

**Feature**: wrk --cd with WRK_FOLLOWUP_FILE set writes in-place follow-up (no shell)

```
# channel open → Branch A
WRK_FOLLOWUP_FILE=tmp
wrk --cd <resolved-abs> -> empty stdout; file: cd /abs\n; exit 0; no LoginInteractive
```

## Preconditions

- Follow-up write is unconditional for `--cd` (no create home-gate / done cwd-gate).
- Descendants must call `enableInPlaceChannel` and set a valid target path.
- No fake shell required (shell must not run).

## Steps

1. Open in-place channel via `enableInPlaceChannel`.
2. Descendants set path form and CLI args.

## Context

- Stdout empty (wrapper prints `cd` to stderr in real life; binary file-only).
- Follow-up always uses expanded absolute path.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	enableInPlaceChannel(t, req)
	if req.RepoDir == "" {
		req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	}
	return nil
}
```
