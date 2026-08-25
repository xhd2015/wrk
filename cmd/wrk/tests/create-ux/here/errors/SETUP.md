# Scenario

**Feature**: `--here` conflicts with window/terminal-on flags

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	return nil
}
```
