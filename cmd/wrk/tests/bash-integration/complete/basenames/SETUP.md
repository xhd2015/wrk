# Scenario

**Feature**: completion returns prefix-filtered project basenames

```
projects alpha, alphalong, beta
wrk --bash-integration --complete -- wrk al 1 -> alpha, alphalong
```

## Steps

1. Seed standard projects.json.
2. Complete first positional with prefix `al`.

```go
func Setup(t *testing.T, req *Request) error {
	seedStandardProjects(req)
	req.CompleteWords = []string{"wrk", "al"}
	req.CompleteCWord = 1
	return nil
}
```