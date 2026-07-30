# Scenario

**Feature**: wrk --where equals form is rejected (Bool flag; no String-compat)

```
wrk --where=spl -> non-zero; equals form fails (no treat-as-basename)
```

## Preconditions

- After Bool binding change, `--where=value` must not be accepted as basename lookup.
- Empty-arg exact string remains under `empty-arg/error/`.

## Steps

- Descendants pass `--where=…` without a separate positional.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureWhereHelpersUsed()
	return nil
}
```
