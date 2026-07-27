# Scenario

**Feature**: print bash-integration script

```
wrk --bash-integration -> stdout integration script
```

## Steps

1. Set Mode to print.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "print"
	return nil
}
```
