# Scenario

**Feature**: wrk --tag-next flag validation and mutual exclusion

```
# invalid flag combos -> non-zero exit before tagscope runs
wrk --dry-run (alone) -> error
# primary+tag-next composition is covered under cmd/wrk/tests/done-compose/
```

## Preconditions

- Bare `--dry-run` requires a host: `--done`, `--merge-back`, `--all-deps`, `--tag-next`, `--propagate-tags`, or `--sync`.
- Composition of `--tag-next` with `--done` / `--merge-back` is allowed at flag layer (see `done-compose/`); still exclusive with other standalone modes (`--list`, etc.).

## Steps

- Descendants set `req.Args` for the invalid combination under test.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```
