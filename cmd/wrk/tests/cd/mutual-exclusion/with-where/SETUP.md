# Scenario

**Feature**: wrk --cd with --where is mutually exclusive

```
wrk --cd /jumpto --where myrepo -> non-zero; mutually exclusive
```

## Steps

1. Parent created abs target.
2. Combine `--cd` and `--where`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--cd", req.MainRepo, "--where", cdBasename}
	return nil
}
```
