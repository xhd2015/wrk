# Scenario

**Feature**: multiple saved projects with same basename for --status

```
# two+ projects match basename
TTY + stdin -> numbered select -> status for chosen saved path
non-TTY -> error listing all candidate absolute paths
```

## Steps

- Descendants seed two saved repos sharing basename `myrepo` at different parent paths.
- Run `wrk myrepo --status` from neutral cwd without local `./myrepo`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureStatusBasenameFallbackHelpersUsed()
	return nil
}
```
