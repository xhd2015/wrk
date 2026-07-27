# Scenario

**Feature**: --projects-dep-graph is mutually exclusive with other mode flags

```
wrk --projects-dep-graph + --projects|--list
  -> non-zero exit
  -> stderr mentions exclusive / wrk:
  -> stdout empty
```

## Steps

1. Descendants combine `--projects-dep-graph` with one peer exclusive mode flag.

## Context

- Same exclusive-mode family as bare `--projects`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	depGraphEnsureHelpersUsed()
	return nil
}
```
