# Scenario

**Feature**: completion after `--where` returns basename candidates

```
wrk --bash-integration --complete -- wrk --where al 2 -> alpha, alphalong
```

## Steps

1. Seed standard projects.json.
2. Complete value position after `--where` with prefix `al`.

```go
func Setup(t *testing.T, req *Request) error {
	seedStandardProjects(req)
	req.CompleteWords = []string{"wrk", "--where", "al"}
	req.CompleteCWord = 2
	return nil
}
```