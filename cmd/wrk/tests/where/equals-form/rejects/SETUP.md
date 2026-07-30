# Scenario

**Feature**: equals form `--where=spl` fails (no treat-as-basename compat)

```
# even if "spl" would match a saved project, equals form is not a basename binding
saved/spl recorded
workspace/ -> wrk --where=spl -> non-zero
```

## Steps

1. Record `saved/spl` so a false-GREEN String binding would succeed.
2. Run `wrk --where=spl` from neutral cwd (no separate positional).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Seed a match so old String("--where") + equals form would GREEN incorrectly.
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", whereBasename)
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--where=" + whereBasename}
	req.TargetDir = ""
	return nil
}
```
