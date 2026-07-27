# Scenario

**Feature**: subject with `[wip]` prefix is WIP

```
IsWipSubject("[wip] experiment") -> true
```

## Steps

1. Set subject to `[wip] experiment`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subject = "[wip] experiment"
	return nil
}
```
