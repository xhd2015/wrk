# Scenario

**Feature**: successful wrk --cd records events.jsonl command "cd"

```
WRK_FOLLOWUP_FILE set
wrk --cd /jumpto -> last event command=cd, exit_code=0, args include --cd
```

## Steps

1. Create abs non-git target; open channel.
2. Run `wrk --cd <abs>`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	target := cdAbsTarget(t, req, "jumpto")
	req.MainRepo = target
	setCDFlagThenPath(req, target)
	return nil
}
```
