# Scenario

**Feature**: wrk skill --header --show accepts either flag order

```
embedded SKILL.md in wrk binary
workspace/ -> wrk skill --header --show -> same YAML frontmatter as --show --header
```

## Steps

1. Run `wrk skill --header --show` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--header", "--show"}
	return nil
}
```
