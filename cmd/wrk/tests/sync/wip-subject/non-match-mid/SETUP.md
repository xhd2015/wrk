# Scenario

**Feature**: mid-string `wip:` (not a subject prefix) is not WIP

```
IsWipSubject("chore: wip: later") -> false
```

## Steps

1. Set subject to `chore: wip: later` (`wip:` appears after another prefix).

```go
func Setup(t *testing.T, req *Request) error {
	req.Subject = "chore: wip: later"
	return nil
}
```
