# Scenario

**Feature**: path-like token after `--done` yields empty custom completion

```
wrk --bash-integration --complete -- wrk --done <path-like> 2 -> empty stdout
```

## Steps

1. Complete the value position after `--done` (cword index 2).
2. Descendants vary the path-like prefix shape only.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Words filled by leaf: wrk --done <path-like>
	req.CompleteCWord = 2
	return nil
}
```
