# Scenario

**Feature**: `--new-terminal` opens ForceNew iTerm without space

```
wrk --new-terminal -> create; iterm ForceNew; no space; no agent
```

## Steps

1. Run `wrk --new-terminal`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--new-terminal"}
	return nil
}
```
