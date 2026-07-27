# Scenario

**Feature**: mid-string `wip:` (not a subject prefix) is not WIP

```
IsWipSubject("chore: wip: later") -> false
```

## Steps

1. Set subject to `chore: wip: later` (`wip:` appears after another prefix).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subject = "chore: wip: later"
	return nil
}
```
