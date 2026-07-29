# Scenario

**Feature**: wrk --bring without -v does not log go mod tidy pre-line

```
# matching bring (tidy still runs silently) -> no $ go … mod tidy on stderr
consumer (require dep) + mydep -> wrk --bring <dep> (no -v)
```

## Steps

1. Matching fixtures.
2. Run `wrk --bring <dep>` without `-v`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", dep}
	return nil
}
```
