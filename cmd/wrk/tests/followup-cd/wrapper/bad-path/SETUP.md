# Scenario

**Feature**: wrapper fails when follow-up cd target is invalid

```
fake wrk writes cd /missing; real wrapper executes cd
  -> stderr prints cd line; wrapper non-zero
```

## Steps

1. Descendants install wrapper and use fake wrk on PATH.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "wrapper")
	return nil
}
```
