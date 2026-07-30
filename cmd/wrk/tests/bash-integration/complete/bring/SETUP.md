# Scenario

**Feature**: completion after `--bring` returns basename candidates

```
wrk --bash-integration --complete -- wrk --bring be 2 -> beta
```

## Steps

1. Seed standard projects.json.
2. Complete value position after `--bring` with prefix `be`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	seedStandardProjects(req)
	req.CompleteWords = []string{"wrk", "--bring", "be"}
	req.CompleteCWord = 2
	return nil
}
```