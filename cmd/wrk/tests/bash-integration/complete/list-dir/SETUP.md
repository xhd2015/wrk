# Scenario

**Feature**: completion after `-l` returns basename candidates

```
wrk --bash-integration --complete -- wrk -l al 2 -> alpha, alphalong
```

## Steps

1. Seed standard projects.json.
2. Complete value position after `-l` with prefix `al`.

```go
func Setup(t *testing.T, req *Request) error {
	seedStandardProjects(req)
	req.CompleteWords = []string{"wrk", "-l", "al"}
	req.CompleteCWord = 2
	return nil
}
```