# Scenario

**Feature**: subject with `wip(` prefix is WIP

```
IsWipSubject("wip(login): sketch") -> true
```

## Steps

1. Set subject to `wip(login): sketch`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Subject = "wip(login): sketch"
	return nil
}
```
