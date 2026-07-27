# Scenario

**Feature**: wrk skill no longer accepts list|show|install subcommands

```
# breaking change from skill-cli Shape 1
wrk skill list|show|install (positional subcommand) -> non-zero, clear error
```

## Steps

- Descendants pass a former skill subcommand as a positional after `skill`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ensureSkillHelpersUsed()
	return nil
}
```
