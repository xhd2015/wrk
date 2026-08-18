# Scenario

**Feature**: two dep args; one stack consumer gets both replaces; the other is gated only for the first

```
# primary requires dep + dep2 + kool; kool requires only dep
cwd=primary -> wrk --dep-replace <dep> <dep2>
  -> two dep headers (argv order)
  -> app: both replaces
  -> kool: only the first
```

## Steps

1. Seed two deps + stack other-checkout.
2. Run multi-arg apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMultiDirStack(t, req)
	req.Args = []string{"--dep-replace", req.DepDir, req.Dep2Dir}
	return nil
}
```
