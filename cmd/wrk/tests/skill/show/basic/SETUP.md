# Scenario

**Feature**: wrk skill --show prints embedded SKILL.md

```
embedded SKILL.md (go:embed) in wrk binary
workspace/ -> wrk skill --show -> stdout embedded SKILL.md bytes
```

## Steps

1. Run `wrk skill --show` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "--show"}
	return nil
}
```
