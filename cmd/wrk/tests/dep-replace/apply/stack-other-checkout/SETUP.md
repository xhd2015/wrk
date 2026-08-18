# Scenario

**Feature**: apply writes absolute replace on primary and another stack checkout that already requires the dep

```
# primary requires dep + kool; replace kool => ./external/kool (own git)
# kool already requires dep
cwd=primary -> wrk --dep-replace <dep>
  -> replace on example.com/app (checkout .)
  -> replace on example.com/kool (checkout external/kool)
```

## Steps

1. Seed primary + independent `external/kool` git repo + filesystem replace.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackOtherCheckout(t, req)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```
