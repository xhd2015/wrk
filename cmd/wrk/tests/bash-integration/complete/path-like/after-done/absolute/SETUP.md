# Scenario

**Feature**: `--done` with absolute `/` path-like cur yields empty candidates

```
wrk --bash-integration --complete -- wrk --done /tmp/x 2 -> empty stdout
```

## Steps

1. Complete after `--done` with current word `/tmp/x`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	req.CompleteWords = []string{"wrk", "--done", "/tmp/x"}
	req.CompleteCWord = 2
	return nil
}
```
