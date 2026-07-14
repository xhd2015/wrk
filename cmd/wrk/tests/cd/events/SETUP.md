# Scenario

**Feature**: successful wrk --cd appends events.jsonl with command "cd"

```
WRK_FOLLOWUP_FILE set; wrk --cd /abs
  -> events.jsonl last event: command=cd, exit_code=0, args include --cd
```

## Preconditions

- Uses in-place success path (exit 0) so shell is not required.
- Auto-record may also run when target is git; event assert is on last event command.

## Steps

1. Open channel; create non-git abs target (avoid create mode confusion).
2. Run `wrk --cd <abs>`.

## Context

- Event `work_dir` is the resolved absolute cd target per requirement.

```go
func Setup(t *testing.T, req *Request) error {
	enableInPlaceChannel(t, req)
	if req.RepoDir == "" {
		req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	}
	return nil
}
```
