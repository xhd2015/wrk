# Scenario

**Feature**: successful wrk --main --where prints main abs path from various cwds

```
linked wt | main subdir | main root | flag order
  -> exit 0; stdout main\n; empty stderr; no shell
```

## Preconditions

- Install fake bash to detect accidental shell launch (log must stay empty of cwd=).

## Steps

- Descendants init layout, set RepoDir, installFakeBash, set compose Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Leaves set RepoDir / Args; install fake bash for accidental-launch detection.
	return nil
}
```
