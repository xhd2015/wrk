# Scenario

**Feature**: subject with uppercase `WIP:` prefix is WIP (case-insensitive)

```
IsWipSubject("WIP: foo") -> true
```

## Steps

1. Set subject to `WIP: foo`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Subject = "WIP: foo"
	return nil
}
```
