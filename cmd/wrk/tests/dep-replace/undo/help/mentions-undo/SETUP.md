# Scenario

**Feature**: wrk -h documents --dep-replace --undo

```
wrk -h
  -> help contains --undo (with dep-replace)
```

## Steps

1. Run `wrk -h`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```
