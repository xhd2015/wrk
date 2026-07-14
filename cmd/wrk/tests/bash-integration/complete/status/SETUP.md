# Scenario

**Feature**: completion after `--status` returns basename candidates

```
wrk --bash-integration --complete -- wrk --status be 2 -> beta
```

## Steps

1. Seed standard projects.json.
2. Complete value position after `--status` with prefix `be`.

```go
func Setup(t *testing.T, req *Request) error {
	seedStandardProjects(req)
	req.CompleteWords = []string{"wrk", "--status", "be"}
	req.CompleteCWord = 2
	return nil
}
```