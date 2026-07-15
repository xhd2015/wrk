# Scenario

**Feature**: wrk --tag-next creates lightweight tags at HEAD

```
# tagscope Plan + Apply -> git tag <next> HEAD; stdout tagged lines + N tag created
wrk --tag-next -> tag refs created locally
```

## Preconditions

- Git available; scopes with planned bumps get new lightweight tags.

## Steps

- Descendants seed repos and set `req.Args = []string{"--tag-next"}`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```