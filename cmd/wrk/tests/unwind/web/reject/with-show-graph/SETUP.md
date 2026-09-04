# Scenario

**Feature**: `--unwind --web` cannot combine with `--show-graph`

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--web", "--show-graph"}
	return nil
}
```
