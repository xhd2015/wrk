# Scenario

**Feature**: events for wrk --main --cd

```
wrk --main --cd -> events.jsonl command="cd" (partner wins)
```

## Steps

- Descendants run successful compose and assert last event.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	return nil
}
```
