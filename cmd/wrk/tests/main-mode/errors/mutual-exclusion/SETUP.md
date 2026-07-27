# Scenario

**Feature**: wrk --main is mutually exclusive with other modes (same family as --cd)

```
wrk --main --list  -> non-zero, mutually exclusive
wrk --main --cd …  -> non-zero, mutually exclusive
```

## Preconditions

- Main repo exists so rejection is about mode combination, not missing git.

## Steps

- Descendants set Args combining `--main` with another mode flag.

## Context

- Error stderr should mention mutually exclusive (and/or the conflicting flags).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	return nil
}
```
