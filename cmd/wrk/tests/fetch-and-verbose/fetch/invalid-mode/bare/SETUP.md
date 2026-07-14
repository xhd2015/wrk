# Scenario

**Feature**: bare wrk --fetch is invalid

```
wrk --fetch -> exit 1
```

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--fetch"}
	return nil
}
```