# Scenario

**Feature**: `--reuse-terminal` uses ModeReuseCurrent

```
wrk --reuse-terminal -> iterm reuse-current script; no space
```

## Steps

1. Run `wrk --reuse-terminal`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--reuse-terminal"}
	return nil
}
```
