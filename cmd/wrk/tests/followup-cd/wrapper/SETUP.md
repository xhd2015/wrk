# Scenario

**Feature**: bash wrk() wrapper auto-cd after successful modes

```
source bash.sh; wrk ...
  -> binary writes follow-up; wrapper prints cd to stderr; builtin cd
```

## Preconditions

- bash available on PATH.
- Git required for create/done/set-task fixtures.

## Steps

1. Set Mode to wrapper; install/print integration script before Run.
2. Descendants set start dir, CLI args, AutoCD, and fixtures.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Mode = "wrapper"
	return nil
}
```
