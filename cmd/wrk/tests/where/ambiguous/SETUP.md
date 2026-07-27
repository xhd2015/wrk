# Scenario

**Feature**: multiple saved projects with same basename for wrk --where

```
# two+ projects match basename
wrk --where spl -> stdout lists all matching absolute paths sorted; exit 0
```

## Steps

- Descendants seed two saved repos sharing basename `spl` at different parent paths.
- Run `wrk --where spl` from neutral cwd without local `./spl`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureWhereHelpersUsed()
	return nil
}```
