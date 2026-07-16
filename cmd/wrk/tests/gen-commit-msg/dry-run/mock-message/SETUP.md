# Scenario

**Feature**: wrk --gen-commit-msg --dry-run prints mock message B for 1 staged file

```
# stage 1 text file, pure plan through wrk binary
repo/ (1 staged) -> wrk --gen-commit-msg --dry-run
  -> stdout: dry-run: would generate commit message for 1 staged file(s)\n
  -> exit 0
```

## Preconditions

- Isolated git repo with hooks disabled.
- Exactly one staged text file (`change.go`).

## Steps

1. Init repo and stage `change.go`.
2. Run `wrk --gen-commit-msg --dry-run` from the repo cwd.

```go
func Setup(t *testing.T, req *Request) error {
	stageOneTextFile(t, req)
	req.Args = []string{"--gen-commit-msg", "--dry-run"}
	return nil
}
```
