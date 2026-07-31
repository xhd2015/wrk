# Scenario

**Feature**: `--title` and `--comment` are always required and non-empty after trim

```
# missing or empty title/comment
wrk --pr [partial flags]
  -> non-zero
  -> stderr names missing/empty --title and/or --comment
```

## Steps

- Leaves set incomplete argv; git fixture optional (flag-layer may fail first).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
