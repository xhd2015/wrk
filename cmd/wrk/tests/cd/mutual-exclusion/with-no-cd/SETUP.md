# Scenario

**Feature**: wrk --cd with --no-cd is rejected

```
wrk --cd /jumpto --no-cd -> non-zero
```

## Steps

1. Parent created abs target.
2. Combine `--cd` and `--no-cd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--cd", req.MainRepo, "--no-cd"}
	return nil
}
```
