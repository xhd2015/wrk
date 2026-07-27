# Scenario

**Feature**: `--push --json` alone is still invalid (`--json` only with `--tag-next`)

```
myrepo -> wrk --push --json
  -> non-zero
  -> stderr mentions --json (and typically only valid with --tag-next)
```

## Steps

1. Seed main repo with origin (json reject is flag-layer; origin optional).
2. Run `wrk --push --json`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushMainWithOrigin(t, req)
	req.Args = []string{"--push", "--json"}
	return nil
}
```
