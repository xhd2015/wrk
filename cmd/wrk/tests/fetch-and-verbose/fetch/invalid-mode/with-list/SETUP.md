# Scenario

**Feature**: wrk --fetch --list is invalid

```
wrk --fetch --list -> exit 1
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--fetch", "--list"}
	return nil
}
```