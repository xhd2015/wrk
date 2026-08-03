# Scenario

**Feature**: -m alone requires --commit

```
workspace/ -> wrk -m "feat: alone"
  -> non-zero; message about requires --commit
```

## Preconditions

- No git repository required; validation fails at flag parse.

## Steps

1. Run `wrk -m "feat: alone"` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"-m", "feat: alone"}
	return nil
}
```
