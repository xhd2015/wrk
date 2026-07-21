# Scenario

**Feature**: wrk --gen-commit-msg --add-all --dry-run plans git add -A on stderr

```
# bare path forwards --add-all; dry-run plans stage-all without mutating index
repo/ (1 staged) -> wrk --gen-commit-msg --add-all --dry-run
  -> stderr: would: git add -A
  -> stdout: mock B for N=1
  -> exit 0
  -> not: unrecognized flag: --add-all (wrk)
```

## Preconditions

- Isolated git repo with hooks disabled.
- Exactly one staged text file (`change.go`) so dry-run can succeed after the would-line.
- Agent is not required (pure plan).

## Steps

1. Init repo and stage `change.go`.
2. Run `wrk --gen-commit-msg --add-all --dry-run` from the repo cwd.

```go
func Setup(t *testing.T, req *Request) error {
	stageOneTextFile(t, req)
	req.Args = []string{"--gen-commit-msg", "--add-all", "--dry-run"}
	return nil
}
```
