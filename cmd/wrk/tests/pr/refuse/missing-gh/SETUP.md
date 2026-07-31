# Scenario

**Feature**: `--pr` fails clearly when `gh` is not on PATH

```
# linked wt + github origin; PATH has git but no gh
linked wt + github origin + PATH without gh
  -> wrk --pr --title T --comment C
  -> non-zero
  -> stderr install prompt (cli.github.com and/or "GitHub CLI" / gh)
```

## Steps

1. Seed linked feature (remote may exist — gh missing is the gate under test).
2. Install PATH without `gh` (git symlinked only); do **not** install fake gh.
3. Run default `--pr` as a **subprocess** (`InProcess=false`) so stripped PATH is
   child-only Env — avoids Capture's temporary process PATH racing parallel Setup (`cp`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Subprocess: ExtraEnv PATH=<git-only> must not os.Setenv process PATH under Capture.
	req.InProcess = false
	setupPrLinkedFeatureRemoteExists(t, req)
	installPathWithoutGh(t, req)
	req.Args = prDefaultArgs()
	return nil
}
```
