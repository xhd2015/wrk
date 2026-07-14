# Scenario

**Feature**: wrk --all-deps --dry-run plans from registered projects but writes nothing

```
# registered deps match consumer requires -> would: lines in project-path order, no side effects
projects.json (mydep1, mydep2) + consumer -> wrk --all-deps --dry-run -> would: wrked N deps

# empty projects -> would: wrked 0 deps
no projects.json -> wrk --all-deps --dry-run -> would: wrked 0 deps
```

## Preconditions

- `--dry-run` is valid only with `--all-deps`.
- Planning uses the same registered-project discovery as the real run.

## Steps

- Descendants register deps (or leave projects empty) and run `wrk --all-deps --dry-run`.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()
	dryRunEnsureHelpersUsed()
	return nil
}

func dryRunEnsureHelpersUsed() {
	_ = allDepsGoModJSON{}
	_ = allDepsReadGoMod
	_ = allDepsHasReplaceForModule
	_ = allDepsReplacePathForModule
	_ = allDepsGitignoreContainsExternal
	_ = allDepsCountGitignoreExternalLines
	_ = allDepsExternalRelPath
	_ = allDepsExternalAbsPath
	_ = registerAllDepsProject
	_ = registerAllDepsProjects
	_ = initAllDepsRepo
	_ = initAllDepsConsumer
	_ = allDepsEnsureHelpersUsed
}
```