# Scenario

**Feature**: wrk --dry-run without --all-deps is rejected before any planning

```
# bare --dry-run (no host) -> non-zero; host list includes --all-deps and --propagate-tags
wrk --dry-run -> error (host list: done|merge-back|all-deps|tag-next|propagate-tags|sync)
```

## Steps

1. Create a consumer git repo requiring `example.com/dep1` (so cwd is a valid git repo with a go.mod, though the error fires before scanning regardless).
2. Run `wrk --dry-run` from the consumer (no `--all-deps`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--dry-run"}
	return nil
}
```
