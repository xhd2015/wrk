# Scenario

**Feature**: wrk --cd without WRK_FOLLOWUP_FILE falls back to stdout path + interactive shell

```
# channel closed → Branch B
WRK_FOLLOWUP_FILE unset
wrk --cd <abs>
  -> stderr: install hint (wrk --bash-integration --install)
  -> stdout: <abs>\n
  -> shell/interactive.LoginInteractive(abs, Base(abs), ...)
  -> wrk exit = shell exit
```

## Preconditions

- Every successful fallback leaf **must** call `installFakeBash` so CI cannot hang.
- Fake bash is resolved via PATH (`bash`) after detect.Shell sees basename `bash` from `SHELL`.
- Follow-up env is **not** set (`UseFollowupEnv` false).

## Steps

1. Create target directory.
2. Install fake interactive shell (exit 0 unless overridden).
3. Descendants set path form / CLI args.

## Context

- Implementer wires `shell/interactive.LoginInteractive`; tests seal stdout/stderr/exit
  and that the child shell's cwd is the resolved absolute target.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Default: channel closed. Leaves may installFakeBash and set paths.
	if req.RepoDir == "" {
		req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	}
	return nil
}
```
