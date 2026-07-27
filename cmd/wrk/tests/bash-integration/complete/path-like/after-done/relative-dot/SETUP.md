# Scenario

**Feature**: `--done` with `./` relative path-like cur yields empty candidates

```
wrk --bash-integration --complete -- wrk --done ./ex 2 -> empty stdout
```

## Steps

1. Complete after `--done` with current word `./ex`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	req.CompleteWords = []string{"wrk", "--done", "./ex"}
	req.CompleteCWord = 2
	return nil
}
```
