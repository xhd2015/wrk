# Scenario

**Feature**: completion after `--add` and `--rm` returns basename candidates

```
wrk --complete for --add and --rm value positions -> prefix-filtered basenames
```

## Steps

1. Seed standard projects.json.
2. Run two complete invocations via `req.CompleteCases`.

```go
func Setup(t *testing.T, req *Request) error {
	seedStandardProjects(req)
	req.CompleteCases = []CompleteCase{
		{Name: "add", Words: []string{"wrk", "--add", "al"}, CWord: 2},
		{Name: "rm", Words: []string{"wrk", "--rm", "be"}, CWord: 2},
	}
	return nil
}
```