# Scenario

**Feature**: wrk skill --show prints embedded wrk SKILL.md content

```
embedded SKILL.md (go:embed)
wrk skill --show [--header] -> stdout full file or YAML header only
```

## Steps

- Descendants set `req.Args` for `wrk skill --show` (and optional `--header`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ensureSkillHelpersUsed()
	return nil
}
```
