# Scenario

**Feature**: multiple saved projects with same basename

```
# two+ projects match basename
TTY + stdin -> numbered select -> create from chosen path
non-TTY -> error listing all candidate absolute paths
```

## Steps

- Descendants seed two saved repos sharing basename `myrepo` at different parent paths.
- Run `wrk myrepo` from neutral cwd without local `./myrepo`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureBasenameFallbackHelpersUsed()
	return nil
}
```