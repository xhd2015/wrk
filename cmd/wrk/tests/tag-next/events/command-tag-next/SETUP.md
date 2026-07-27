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
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupRootBumpRepo(t, req)
	req.Args = []string{"--tag-next"}
	return nil
}
```