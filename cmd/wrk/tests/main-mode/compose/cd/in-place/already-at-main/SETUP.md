# Scenario

**Feature**: already at main + wrk --main --cd still runCd (follow-up), not bare-main notice-only

```
WRK_FOLLOWUP_FILE set
mainRepo (cwd) -> wrk --main --cd
  -> follow-up: cd <mainRepo>\n
  -> empty stdout
  -> NOT bare-main "already at main repository root" short-circuit
```

## Steps

1. Initialize main repo; cwd = main root.
2. Open follow-up channel; install fake bash to detect accidental shell.
3. Run `wrk --main --cd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	enableFollowupChannel(t, req)
	installFakeBash(t, req, 0)
	setMainCdArgs(req, "--main", "--cd")
	return nil
}
```
