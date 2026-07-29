# Scenario

**Feature**: wrk --bring -v logs timestamped go mod tidy pre-line on stderr

```
# consumer requires dep -> wrk --bring <dep> -v
#   -> replace present; exit 0
#   -> stderr: [YYYY-MM-DD HH:MM:SS] $ go -C <consumerModDir> mod tidy
consumer (require dep) + mydep -> wrk --bring <dep> -v
```

## Steps

1. Create matching consumer + dep (same as `bring/basic`).
2. Run `wrk --bring <dep> -v`.

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
	req.ConsumerModDir = consumer
	req.Args = []string{"--bring", dep, "-v"}
	return nil
}
```
