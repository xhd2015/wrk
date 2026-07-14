# Scenario

**Feature**: wrk --cd with --list is mutually exclusive

```
wrk --cd /jumpto --list -> non-zero; mutually exclusive
```

## Steps

1. Parent created abs target.
2. Args = `--cd`, path, `--list`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--cd", req.MainRepo, "--list"}
	return nil
}
```
