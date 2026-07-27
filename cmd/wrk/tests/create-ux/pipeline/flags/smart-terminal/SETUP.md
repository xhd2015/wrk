# Scenario

**Feature**: `--smart-terminal` uses ModeSmart

```
wrk --smart-terminal -> iterm smart script; no space
```

## Steps

1. Run `wrk --smart-terminal`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--smart-terminal"}
	return nil
}
```
