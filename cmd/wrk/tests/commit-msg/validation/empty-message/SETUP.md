# Scenario

**Feature**: empty -m value is rejected

```
workspace/ -> wrk --commit -m ""
  -> non-zero; empty / invalid message
```

## Preconditions

- Locked decision D7: empty message rejected.

## Steps

1. Run `wrk --commit -m ""` from neutral cwd (empty string argv).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--commit", "-m", ""}
	return nil
}
```
