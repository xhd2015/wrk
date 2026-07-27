# Scenario

**Feature**: ordinary non-WIP subject is not WIP

```
IsWipSubject("feat: done") -> false
```

## Steps

1. Set subject to `feat: done`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subject = "feat: done"
	return nil
}
```
