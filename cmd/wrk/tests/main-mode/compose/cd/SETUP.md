# Scenario

**Feature**: wrk --main --cd jumps to the main repository via existing runCd semantics

```
# resolve main from cwd; then runCd(main, execArgs)
wrk --main --cd
  # Branch A — follow-up active
  WRK_FOLLOWUP_FILE set -> empty stdout; follow-up: cd <main>\n; no shell
  # Branch B — channel closed
  unset -> stderr install hint; stdout main\n; LoginInteractive(main)

# already at main root: still runCd (no bare-main notice-only short-circuit)
# --main --cd --exec allowed (same as bare --cd --exec)
# event command="cd"; args include --main and --cd
```

## Preconditions

- In-place leaves call `enableFollowupChannel`.
- Fallback leaves call `installFakeBash`.
- Zero positionals with `--main`; path positional → unexpected arguments.

## Steps

1. Descendants init layout and set compose Args.
2. Assert follow-up line or fake shell cwd = main root.

## Context

- Partner wins over bare `--main` nested-shell / already-at-root notice.
- Flag order free: `--cd --main` ≡ `--main --cd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setMainCdArgs(req, "--main", "--cd")
	return nil
}
```
