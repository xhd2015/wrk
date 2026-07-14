# Scenario

**Feature**: wrk skill with no action or --help/-h prints skill-level usage

```
wrk skill | wrk skill --help | wrk skill -h
  -> skill usage mentioning --list, --show, --install
  -> exit 0, trailing newline on stdout
```

## Steps

- Descendants set empty skill args or help flags only (no action flag).

```go
func Setup(t *testing.T, req *Request) error {
	ensureSkillHelpersUsed()
	return nil
}
```
