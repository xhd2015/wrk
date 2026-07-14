# Scenario

**Feature**: wrk skill --list / -l prints the single wrk skill name

```
embedded SKILL.md in wrk binary
wrk skill --list | -l -> stdout wrk\n
```

## Steps

- Descendants run `wrk skill --list` or `wrk skill -l`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureSkillHelpersUsed()
	return nil
}
```
