# Scenario

**Feature**: successful `wrk --new` records events.jsonl `command: "create"`

```
myrepo -> wrk --new
  -> events.jsonl: command=create, exit_code=0, args include --new
```

## Steps

1. Init main repo (parent).
2. Run `wrk --new`.
3. Assert last event is create with `--new` in args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--new"}
	return nil
}
```
