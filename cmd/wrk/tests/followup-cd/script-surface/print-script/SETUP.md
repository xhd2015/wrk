# Scenario

**Feature**: print bash-integration script

```
wrk --bash-integration -> stdout integration script
```

## Steps

1. Set Mode to print.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "print"
	return nil
}
```
