# Scenario

**Feature**: wrk --where CLI forms `wrk --where BASE` and `wrk BASE --where` are equivalent

```
# Bool("--where") + exactly one basename positional
wrk --where spl  <->  wrk spl --where
# same projects.json lookup
```

## Preconditions

- Single saved match for basename `spl` (or leaf-chosen basename).
- Neutral cwd without a local `./spl` directory requirement (lookup is basename-only).

## Steps

1. Record one saved project named `spl`.
2. Leaf chooses flag-then-basename vs basename-then-flag.

## Context

- Existing `single-match/basic` covers flag-then-basename; this group owns the path-then-flag form and parity.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", whereBasename)
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	return nil
}
```
