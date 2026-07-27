# Scenario

**Feature**: print-script contains completion registration and WRK_HOME resolution

```
wrk --bash-integration -> script references:
  complete -o default -F _wrk wrk
  path-like yield via compopt -o default
  WRK_HOME + --complete callback
```

## Steps

1. Run `wrk --bash-integration` with default isolated env.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	requireMode(t, req, "print")
	requireNoPreseed(t, req)
	return nil
}
```
