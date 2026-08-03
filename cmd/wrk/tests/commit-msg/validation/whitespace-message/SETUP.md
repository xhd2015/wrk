# Scenario

**Feature**: whitespace-only -m value is rejected

```
workspace/ -> wrk --commit -m "   "
  -> non-zero; empty / invalid message
```

## Preconditions

- Locked decision D7: whitespace-only message rejected (same class as empty).

## Steps

1. Run `wrk --commit -m "   "` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--commit", "-m", "   "}
	return nil
}
```
