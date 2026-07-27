# Scenario

**Feature**: wrk --bash-integration prints bash completion script

```
wrk --bash-integration -> stdout script with:
  path-like _wrk yield (compopt -o default)
  complete -o default -F _wrk wrk
  --complete callback
```

## Steps

1. Set `req.Mode = "print"`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "print"
	req.DryRun = false
	return nil
}
```
