# Scenario

**Feature**: wrk --repos is mutually exclusive with other wrk modes

```
wrk --repos + other mode -> error (mutually exclusive)
```

## Steps

- Descendant scenarios combine `--repos` with another mode flag.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--repos"}
	return nil
}
```
