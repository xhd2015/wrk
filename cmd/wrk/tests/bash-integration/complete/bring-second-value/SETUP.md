# Scenario

**Feature**: completion after a first `--bring` value still offers project basenames

```
# today only words[cword-1]=="--bring" completes; second value must too
wrk --bring already-typed be  -> beta
```

## Steps

1. Seed standard projects.json.
2. Complete the **second** value after `--bring` with prefix `be`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	seedStandardProjects(req)
	req.CompleteWords = []string{"wrk", "--bring", "already-typed", "be"}
	req.CompleteCWord = 3
	return nil
}
```
