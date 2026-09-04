# Scenario

**Feature**: `--dev` is not valid with `--unwind --web`

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--web", "--dev"}
	return nil
}
```
