# Scenario

**Feature**: wrk --version prints v0.0.1 on stdout

```
embedded VERSION.txt (0.0.1) in wrk binary
workspace/ -> wrk --version -> stdout v0.0.1\n, no events.jsonl
```

## Steps

1. Run `wrk --version` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--version"}
	return nil
}
```