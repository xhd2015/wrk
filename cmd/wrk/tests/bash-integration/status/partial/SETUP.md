# Scenario

**Feature**: status reports partial when script and profile markers disagree

```
partial integration state
wrk --bash-integration --status -> partial, exit 1
```

## Steps

1. Descendants pre-seed script-only or marker-only state.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "status")
	return nil
}
```