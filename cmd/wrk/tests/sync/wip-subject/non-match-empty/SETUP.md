# Scenario

**Feature**: empty subject is not WIP

```
IsWipSubject("") -> false
```

## Steps

1. Set subject to empty string.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subject = ""
	return nil
}
```
