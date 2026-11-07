# Scenario

**Feature**: embedded edit-dependency-code SKILL.md

```
workspace/ -> wrk skill --show edit-dependency-code -> stdout that skill's SKILL.md
```

## Steps

1. Run `wrk skill --show edit-dependency-code` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show", "edit-dependency-code"}
	return nil
}
```
