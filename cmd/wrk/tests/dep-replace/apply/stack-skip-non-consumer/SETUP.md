# Scenario

**Feature**: other-checkout module with neither require nor replace is left unchanged

```
# primary requires dep + kool; replace kool => ./external/kool
# kool has neither require nor replace for dep
cwd=primary -> wrk --dep-replace <dep>
  -> replace on example.com/app only
  -> kool go.mod unchanged; no skip line
```

## Steps

1. Seed stack with non-consumer other checkout.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackSkipNonConsumer(t, req)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```
