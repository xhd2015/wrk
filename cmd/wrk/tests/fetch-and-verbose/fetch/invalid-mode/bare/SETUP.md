# Scenario

**Feature**: bare wrk --fetch is invalid

```
wrk --fetch -> exit 1
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--fetch"}
	return nil
}
```