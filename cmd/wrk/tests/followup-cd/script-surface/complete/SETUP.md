# Scenario

**Feature**: bash-integration --complete lists flags including --no-cd and --force-cd

```
wrk --bash-integration --complete -- <words> <cword> -> flag candidates
  include --no-cd and --force-cd
```

## Steps

1. Set Mode to complete.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "complete"
	return nil
}
```
