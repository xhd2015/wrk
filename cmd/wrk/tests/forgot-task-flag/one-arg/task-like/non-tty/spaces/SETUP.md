# Scenario

**Feature**: one-arg multi-word, non-TTY from git repo cwd → default auto-promote

```
(cd myrepo && wrk "fix the login bug") non-TTY
  -> exit 0; create from cwd with task slug under WRK_HOME
```

## Steps

1. Cwd = mainRepo.
2. TargetDir (first positional) = multi-word task text.

```go
func Setup(t *testing.T, req *Request) error {
	setupOneArg(t, req, taskLikeSpaces)
	return nil
}
```
