# Scenario

**Feature**: wrk --where with zero projects.json matches

```
# no saved project matches basename
wrk --where spl -> non-zero exit, stderr no-match message, empty stdout
```

## Steps

- Descendants run `wrk --where spl` with no matching saved project.

```go
func Setup(t *testing.T, req *Request) error {
	ensureWhereHelpersUsed()
	return nil
}```
