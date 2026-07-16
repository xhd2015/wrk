# Scenario

**Feature**: bare `wrk --push` from main advances origin/main and prints confirm line

```
# main checkout + bare origin (upstream set)
myrepo (main) + origin
  -> wrk --push
  -> pushed main → origin/main
  -> origin/main == local HEAD
```

## Steps

1. Seed main repo with bare `origin` and `push -u origin main`.
2. Run `wrk --push` from the main checkout.

```go
func Setup(t *testing.T, req *Request) error {
	setupPushMainWithOrigin(t, req)
	req.Args = []string{"--push"}
	return nil
}
```
