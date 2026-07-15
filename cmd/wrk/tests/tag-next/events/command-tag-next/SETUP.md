# Scenario

**Feature**: successful tag-next records events.jsonl command tag-next

```
# wrk --tag-next success -> events.jsonl last event command=tag-next
git repo -> wrk --tag-next -> event appended
```

## Steps

1. `setupRootBumpRepo`.
2. Run `wrk --tag-next`.

```go
func Setup(t *testing.T, req *Request) error {
	setupRootBumpRepo(t, req)
	req.Args = []string{"--tag-next"}
	return nil
}
```