# Scenario

**Feature**: wrk --sync flag validation and Phase 1 success path

```
# invalid flag combos / bare dry-run / positionals -> non-zero before sync body
# valid --sync / --sync --dry-run on main-only -> exit 0 summary (zeros)
wrk --sync [--dry-run] | invalid combos -> summary or flag error
```

## Preconditions

- `--sync` is a standalone mode; no positionals.
- `--dry-run` is valid with `--sync` (and existing modes such as `--tag-next` / `--propagate-tags`).
- Bare `wrk --dry-run` remains rejected; host list includes `--done`, `--merge-back`, `--tag-next`, `--propagate-tags`, and `--sync`.

## Steps

- Descendants set `req.RepoDir` / `req.Args` for the scenario under test.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	syncEnsureHelpersUsed()
	return nil
}
```
