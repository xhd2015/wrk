# Scenario

**Feature**: --all reject forms (bare --all; --all with path args)

```
wrk --all
  | wrk --dep-update --all <path>
  -> non-zero
```

## Steps

- Leaves set invalid `--all` argument combinations.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```
