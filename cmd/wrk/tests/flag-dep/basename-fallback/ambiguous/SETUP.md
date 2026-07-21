# Scenario

**Feature**: multiple saved deps with same basename for --dep resolution

```
# two+ saved dep projects match basename
TTY + stdin -> numbered select -> --dep succeeds for chosen dep
non-TTY -> error listing all candidate absolute paths
```

## Steps

- Descendants seed two saved dep repos sharing basename `mydep` at different parent paths.
- Run `wrk --dep mydep` from consumer cwd without local `./mydep`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureDepBasenameFallbackHelpersUsed()
	return nil
}
```