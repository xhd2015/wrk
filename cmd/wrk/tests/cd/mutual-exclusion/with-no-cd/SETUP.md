# Scenario

**Feature**: wrk --cd with --no-cd is rejected

```
wrk --cd /jumpto --no-cd -> non-zero
```

## Steps

1. Parent created abs target.
2. Combine `--cd` and `--no-cd`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--cd", req.MainRepo, "--no-cd"}
	return nil
}
```
