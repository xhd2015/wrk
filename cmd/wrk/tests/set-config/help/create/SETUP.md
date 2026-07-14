# Scenario

**Feature**: help with `--create` (and not `--show`) prints dedicated create usage

```
wrk --set-config --create -h|--help
  -> dedicated create UX usage (flags, conflicts, examples)
  -> exit 0; short-circuits before requiring UX flags / write
```

## Steps

- Leaves set `--create` plus help form (any order); optional co-present UX flags still show help only.

```go
func Setup(t *testing.T, req *Request) error {
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	// Level: dedicated create help under --set-config --create.
	return nil
}
```
