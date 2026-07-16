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
func Setup(t *testing.T, req *Request) error {
	setupPushMainWithOrigin(t, req)
	req.Args = []string{"--push", "--json"}
	return nil
}
```
