# Scenario

**Feature**: wrk skill --show --header prints YAML frontmatter only

```
embedded SKILL.md in wrk binary
workspace/ -> wrk skill --show --header -> stdout ---\nname: wrk\n---\n
```

## Steps

1. Run `wrk skill --show --header` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show", "--header"}
	return nil
}
```
