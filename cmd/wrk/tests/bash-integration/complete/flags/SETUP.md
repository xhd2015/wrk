# Scenario

**Feature**: completion returns wrk flags for `wrk -<tab>`

```
wrk --bash-integration --complete -- wrk - 1 -> matching flags
```

## Steps

1. Seed standard projects (flags should not depend on projects).
2. Complete flag prefix `-` at word index 1.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	seedStandardProjects(req)
	req.CompleteWords = []string{"wrk", "-"}
	req.CompleteCWord = 1
	return nil
}
```