# Scenario

**Feature**: non-composed mode pairs remain mutually exclusive after composition lands

```
# composition must not accidentally open other exclusives
wrk --tag-next --list -> still mutually exclusive
wrk --reinstall-local --sync -> still mutually exclusive (no primary)
wrk --reinstall-local --list -> still mutually exclusive
wrk --gen-commit-msg --sync -> still mutually exclusive (no primary; post-only with primary)
```

## Preconditions

- Standalone exclusives such as `--tag-next` + `--list` stay rejected.
- Bare `--reinstall-local` stays exclusive with `--sync` / `--list` (only composes after primary).
- Bare `--gen-commit-msg --sync` (no `--done`/`--merge-back`) stays exclusive; `--sync` is only a
  post modifier of primary, not of bare gen-commit.

## Steps

- Leaves set still-invalid mode pairs.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: regression exclusives need a valid git cwd for mode flags.
	skipIfNoGit(t)
	return nil
}
```
