# Scenario

**Feature**: `--done` with `../` parent-relative path-like cur yields empty candidates

```
wrk --bash-integration --complete -- wrk --done ../foo 2 -> empty stdout
```

## Steps

1. Complete after `--done` with current word `../foo`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	req.CompleteWords = []string{"wrk", "--done", "../foo"}
	req.CompleteCWord = 2
	return nil
}
```
