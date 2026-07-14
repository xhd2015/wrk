# Scenario

**Bug**: hard-error stderr from `main` must end with trailing `\n`

```
# ordinary error (not ExitCodeError) printed via main → stderr + exit 1
wrk --open-in-agen
  -> exit 1
  -> stderr = "unrecognized flag: --open-in-agen\n"
  # last byte of stderr is \n so the shell prompt stays on its own line
```

## Preconditions

- Uses root harness `Request` / `Response` / `Run` (built `wrk` binary under session fixture).
- No git fixture required — flag parse fails before checkout checks.

## Steps

1. Set process cwd to `WorkRoot` (isolated temp; no `.git` needed).
2. Run `wrk --open-in-agen` (typo of a real-looking long flag → unrecognized).

## Context

- Scope of production fix: `go-pkgs/cmd/wrk/main.go` prints `err.Error()` with a trailing newline for non-`ExitCodeError` failures.
- Today: `fmt.Fprint(os.Stderr, err.Error())` — no `\n`; shell prompt glues to the message.
- Desired: last byte of stderr is `\n` (prefer `Fprintln` / single append).

```go
func Setup(t *testing.T, req *Request) error {
	// Parse-time hard error; cwd need not be a git checkout.
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--open-in-agen"}
	req.TargetDir = ""
	return nil
}
```
