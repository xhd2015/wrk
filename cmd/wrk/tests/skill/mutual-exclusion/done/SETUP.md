# Scenario

**Feature**: wrk skill --list --done is mutually exclusive

```
embedded SKILL.md (skill path still parses first)
workspace/ -> wrk skill --list --done -> non-zero, mutually exclusive
```

## Steps

1. Run `wrk skill --list --done` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "--list", "--done"}
	return nil
}
```
