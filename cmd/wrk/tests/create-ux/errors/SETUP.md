# Scenario

**Feature**: create UX flag validation and platform errors

```
conflicting flags | non-darwin window -> non-zero; clear stderr
```

## Steps

- Leaves set invalid combinations or platform mocks.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	return nil
}
```
