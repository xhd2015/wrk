# Scenario

**Feature**: wrk --where rejects non-basename lookup arguments

```
# path separator or absolute path is not a basename
wrk --where sub/spl or wrk --where /abs/.../spl -> non-zero, basename-only rejection
```

## Steps

- Descendants seed a saved `spl` project when needed and pass a non-basename `--where` arg.

```go
func Setup(t *testing.T, req *Request) error {
	ensureWhereHelpersUsed()
	return nil
}```
