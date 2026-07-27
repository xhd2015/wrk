# Scenario

**Feature**: wrk --bring is mutually exclusive with --dep

```
# wrk --bring <dep> --dep <dep> -> non-zero; stderr mentions mutually exclusive
wrk --bring <path> --dep <path> -> mode conflict (hard error)
```

## Steps

1. Create consumer + dep fixtures (same shape as basic) so paths are valid if parse reaches them.
2. Run `wrk --bring <dep> --dep <dep>` from consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.Args = []string{"--bring", dep, "--dep", dep}
	return nil
}
```
