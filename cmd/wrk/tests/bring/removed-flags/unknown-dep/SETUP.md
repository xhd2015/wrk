# Scenario

**Feature**: `wrk --dep` is an unknown flag after hard removal

```
# end-state: --dep is not a registered flag
consumer requires dep + mydep -> wrk --dep <dep>
  -> non-zero; stderr unknown/invalid flag naming --dep
  -> no external worktree
```

## Steps

1. Create consumer requiring `example.com/dep` and dep repo `mydep` (so a live `--dep` would succeed).
2. Run `wrk --dep <dep>` — must be rejected as unknown flag after removal.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)
	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.Args = []string{"--dep", dep}
	return nil
}
```
