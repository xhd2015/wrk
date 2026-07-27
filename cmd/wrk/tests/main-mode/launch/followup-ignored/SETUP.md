# Scenario

**Feature**: wrk --main always nests a shell even when WRK_FOLLOWUP_FILE is set

```
# channel open would make --cd write in-place cd; --main ignores it
WRK_FOLLOWUP_FILE set
cwd = main subdir
wrk --main
  -> still LoginInteractive(mainRepo, ...)
  -> follow-up file remains empty (never writeFollowupCD)
  -> minimal UX
```

## Steps

1. Create main repo + nested subdir.
2. Enable follow-up channel (`UseFollowupEnv` + empty follow-up file).
3. Install fake bash (exit 0).
4. Run `wrk --main` from the subdir.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, sub := initMainRepoSubdir(t, req, "pkg", "tool")
	req.MainRepo = mainRepo
	req.RepoDir = sub
	enableFollowupChannel(t, req)
	installFakeBash(t, req, 0)
	setMainArgs(req)
	return nil
}
```
