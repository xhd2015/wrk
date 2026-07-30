# Scenario

**Feature**: wrk --main composed with partner flags (`--where` / `--cd`)

```
# partner wins; resolve main from cwd; zero positionals
wrk --main --where -> print main abs path (no shell)
wrk --main --cd    -> runCd(main) (follow-up or fallback shell)

# flag order free for bool partners
wrk --where --main  ≡  wrk --main --where
wrk --cd --main     ≡  wrk --main --cd
```

## Preconditions

- Shared helpers from `main-mode/SETUP.md` (`initLinkedWorktree`, `installFakeBash`, …).
- Do **not** redefine root `Request` / `Response` / `Run`.
- Arity with `--main`: exactly **0** positionals; extra args → unexpected arguments.
- Bare `--main` nested-shell policy stays under `launch/` / `already-at-root/`.

## Steps

1. Descendants build git layout, set `RepoDir` (cwd), and partner Args.
2. `where` leaves assert stdout path; `cd` leaves assert follow-up or fake shell.

## Context

- Main path = `resolveMainRepoForWorkDir(cwd)` (ShowToplevel + ResolveMainRepo), **not** projects.json basename lookup.
- Event command is the partner name (`where` / `cd`); args include both `--main` and the partner flag.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureMainHelpersUsed()
	return nil
}
```
