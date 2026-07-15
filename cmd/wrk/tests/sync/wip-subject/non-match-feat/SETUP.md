# Scenario

**Feature**: ordinary non-WIP subject is not WIP

```
IsWipSubject("feat: done") -> false
```

## Steps

1. Set subject to `feat: done`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Subject = "feat: done"
	return nil
}
```
