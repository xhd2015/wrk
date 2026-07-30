# Scenario

**Feature**: bare `wrk --dry-run` host list no longer includes `--all-deps`

```
wrk --dry-run -> non-zero; --dry-run is only valid with … (hosts without --all-deps)
```

## Steps

1. Create a trivial consumer git repo.
2. Run `wrk --dry-run` with no host mode.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, false)
	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--dry-run"}
	return nil
}
```
