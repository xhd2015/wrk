# Scenario

**Feature**: --force-cd and --no-cd are mutually exclusive (hard error)

```
wrk --force-cd --no-cd ... -> non-zero; no follow-up; no shell
```

## Preconditions

- Binary mode; git fixtures optional (parse should fail before mode work).

## Steps

1. Descendants combine both flags on a create invocation with channel prepared.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "binary")
	return nil
}
```
