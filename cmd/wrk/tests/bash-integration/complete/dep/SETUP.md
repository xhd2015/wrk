# Scenario

**Feature**: completion after `--dep` returns basename candidates

```
wrk --bash-integration --complete -- wrk --dep be 2 -> beta
```

## Steps

1. Seed standard projects.json.
2. Complete value position after `--dep` with prefix `be`.

```go
func Setup(t *testing.T, req *Request) error {
	seedStandardProjects(req)
	req.CompleteWords = []string{"wrk", "--dep", "be"}
	req.CompleteCWord = 2
	return nil
}
```