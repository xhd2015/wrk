# Scenario

**Feature**: a single `--bring` still succeeds after the flag becomes a repeatable slice

```
# flag type change String → StringSlice must not break one-path form
consumer requires dep -> wrk --bring <dep>
  -> same as basic: external path + replace; exit 0
```

## Steps

1. Create consumer requiring `example.com/dep` and dep repo `mydep`.
2. Run `wrk --bring <dep>` (single occurrence) with InProcess.

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
