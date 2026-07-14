# Scenario

**Feature**: wrk skill --install copies embedded SKILL.md to agent directories

```
embedded SKILL.md in wrk binary
wrk skill --install [flags] -> install.HandleInstall (SkillDirName wrk)
```

## Steps

- Descendants run `wrk skill --install` with flags.

```go
func Setup(t *testing.T, req *Request) error {
	ensureSkillHelpersUsed()
	return nil
}
```
