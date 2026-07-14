# Scenario

**Feature**: wrk --fetch --done is invalid

```
wrk --fetch --done -> exit 1
```

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--fetch", "--done"}
	return nil
}
```