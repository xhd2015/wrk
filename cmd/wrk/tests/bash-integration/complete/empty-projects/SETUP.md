# Scenario

**Feature**: empty projects.json yields no basename candidates; flags still work

```
no projects.json
wrk --complete -- wrk al 1 -> empty stdout
wrk --complete -- wrk - 1 -> flags still returned
```

## Steps

1. Do not seed projects.
2. Run two complete invocations: basename attempt and flag attempt.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CompleteCases = []CompleteCase{
		{Name: "basename", Words: []string{"wrk", "al"}, CWord: 1},
		{Name: "flags", Words: []string{"wrk", "-"}, CWord: 1},
	}
	return nil
}
```