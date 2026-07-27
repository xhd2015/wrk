# Scenario

**Feature**: `--reuse-terminal` uses ModeReuseCurrent

```
wrk --reuse-terminal -> iterm reuse-current script; no space
```

## Steps

1. Run `wrk --reuse-terminal`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--reuse-terminal"}
	return nil
}
```
