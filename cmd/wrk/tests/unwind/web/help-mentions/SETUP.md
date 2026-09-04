# Scenario

**Feature**: `wrk -h` documents `--unwind --web`

```
wrk -h
  -> --unwind --web and --port
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```
