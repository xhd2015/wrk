# Scenario

**Feature**: wrk --fetch --done is invalid

```
wrk --fetch --done -> exit 1
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--fetch", "--done"}
	return nil
}
```