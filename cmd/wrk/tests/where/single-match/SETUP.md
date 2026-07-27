# Scenario

**Feature**: single saved project match for wrk --where

```
# one projects.json entry matches basename
neutral cwd -> wrk --where spl -> stdout = saved absolute path
```

## Steps

- Descendants seed exactly one saved project whose basename matches the lookup arg.
- Run `wrk --where <basename>` from a cwd without a local `./<basename>` entry.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureWhereHelpersUsed()
	return nil
}```
