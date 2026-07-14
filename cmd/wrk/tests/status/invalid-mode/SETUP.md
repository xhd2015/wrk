# Scenario

**Feature**: wrk --status is mutually exclusive with other wrk modes

```
# status and another mode both request control of the invocation
wrk --status + other mode -> error (mutually exclusive)
```

## Preconditions

- The effective cwd is inside a git work tree so mode validation is not masked by git discovery errors.

## Steps

- Descendant scenarios combine `--status` with another mode flag.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--status"}
	return nil
}
```
