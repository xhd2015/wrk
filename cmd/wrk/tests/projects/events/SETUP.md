# Scenario

**Feature**: append-only event logging on every wrk invocation

```
every wrk run -> append one JSON line to events.jsonl (success or failure)
```

## Steps

- Descendants run commands that succeed or fail and assert the appended event line.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureProjectsHelpersUsed()
	return nil
}
```