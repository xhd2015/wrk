# Scenario

**Feature**: wrk skill --install --cursor --dry-run plans .cursor/skills/wrk

```
embedded SKILL.md in wrk binary
workspace/ -> wrk skill --install --cursor --dry-run -> dry-run lines, no writes
```

## Steps

1. Run `wrk skill --install --cursor --dry-run` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "--install", "--cursor", "--dry-run"}
	return nil
}
```
