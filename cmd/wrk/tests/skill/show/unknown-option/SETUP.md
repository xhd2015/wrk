# Scenario

**Feature**: wrk skill --show rejects unknown flags

```
embedded SKILL.md in wrk binary
workspace/ -> wrk skill --show --nope -> exit ≠0, stderr unknown option
```

## Steps

1. Run `wrk skill --show --nope` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "--show", "--nope"}
	return nil
}
```
