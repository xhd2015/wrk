# Scenario

**Feature**: wrk skill is mutually exclusive with other wrk modes

```
wrk skill <action flags> + another mode flag -> non-zero, mutually exclusive
```

## Steps

- Descendants combine skill action flags with another wrk mode flag.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ensureSkillHelpersUsed()
	return nil
}
```
