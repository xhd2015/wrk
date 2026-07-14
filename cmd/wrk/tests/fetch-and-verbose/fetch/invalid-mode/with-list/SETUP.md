# Scenario

**Feature**: wrk --fetch --list is invalid

```
wrk --fetch --list -> exit 1
```

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--fetch", "--list"}
	return nil
}
```