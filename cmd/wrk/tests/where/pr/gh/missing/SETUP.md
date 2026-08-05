# Scenario

**Feature**: missing `gh` on PATH fails clearly for `--where --pr`

```
recorded linked + PATH without gh
  -> wrk --where --pr URL
  -> non-zero
  -> stderr install prompt (cli.github.com and/or GitHub CLI / gh)
```

## Steps

1. Seed recorded linked fixture (github origin present).
2. Install PATH without gh (git only); do **not** install fake gh.
3. Run as **subprocess** (`InProcess=false`) so stripped PATH is child-only Env.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Subprocess: ExtraEnv PATH=<git-only> must not os.Setenv process PATH under Capture.
	req.InProcess = false
	// Seed git layout without fake gh install.
	mainRepo := wherePrSetupMainWithOrigin(t, req, "myrepo")
	linked := wherePrAddLinkedOnHead(t, req, mainRepo, "linked-wt")
	recordSavedProject(t, req, mainRepo)
	req.MainRepo = mainRepo
	req.WtDir = linked
	req.WtBranch = wherePrHeadBranch
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	installWherePrPathWithoutGh(t, req)
	req.Args = wherePrArgs(wherePrURL)
	return nil
}
```
