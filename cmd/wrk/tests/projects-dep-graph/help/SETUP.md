# Scenario

**Feature**: root help documents --projects-dep-graph

```
# user discovers the flag from usage
wrk -h / wrk --help -> usage lists --projects-dep-graph
```

## Steps

1. Descendants invoke help and assert the flag is mentioned.

## Context

- Help may print on stdout and/or stderr; leaves search both.

```go
func Setup(t *testing.T, req *Request) error {
	depGraphEnsureHelpersUsed()
	return nil
}
```
