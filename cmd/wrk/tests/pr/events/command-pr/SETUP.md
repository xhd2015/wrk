# Scenario

**Feature**: successful bare `wrk --pr` records events.jsonl `command: "pr"`

```
linked wt + origin + fake gh -> wrk --pr --title T --comment C
  -> events.jsonl last: command=pr, exit_code=0, args include --pr/--title/--comment
```

## Steps

1. Seed linked feature (remote missing OK — ensure-push path).
2. Install fake gh; run default `--pr` argv.
3. Assert last events.jsonl event (do not re-invoke wrk before read).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeature(t, req)
	installFakeGh(t, req)
	req.Args = prDefaultArgs()
	return nil
}
```
