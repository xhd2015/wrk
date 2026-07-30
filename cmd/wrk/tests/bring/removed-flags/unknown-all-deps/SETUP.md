# Scenario

**Feature**: `wrk --all-deps` is an unknown flag after hard removal

```
wrk --all-deps -> non-zero; stderr unknown/invalid flag naming --all-deps
```

## Steps

1. Create a consumer git repo with a require (live `--all-deps` might exit 0 with `wrked 0 deps` if projects empty).
2. Run `wrk --all-deps`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```
