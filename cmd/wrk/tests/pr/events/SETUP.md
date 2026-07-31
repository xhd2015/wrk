# Scenario

**Feature**: successful bare `--pr` records events.jsonl `command: "pr"`

```
linked wt + github origin + fake gh
  -> wrk --pr --title T --comment C
  -> events.jsonl last: command=pr, exit_code=0, args include --pr
```

## Steps

- Leaf seeds apply fixture and asserts event without a second wrk invoke.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
