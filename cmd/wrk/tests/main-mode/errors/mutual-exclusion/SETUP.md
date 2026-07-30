# Scenario

**Feature**: wrk --main remains mutually exclusive with other standalone modes (not with where/cd partners)

```
wrk --main --list           -> non-zero, mutually exclusive
wrk --main --where --cd     -> non-zero, mutually exclusive (two partners)
# compose partners --where / --cd alone are allowed under compose/
```

## Preconditions

- Main repo exists so rejection is about mode combination, not missing git.

## Steps

- Descendants set Args combining `--main` with another exclusive mode flag (or both partners).

## Context

- Error stderr should mention mutually exclusive (and/or the conflicting flags).
- `--main --cd` and `--main --where` are **not** exclusives; they are compose partners.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	return nil
}
```
