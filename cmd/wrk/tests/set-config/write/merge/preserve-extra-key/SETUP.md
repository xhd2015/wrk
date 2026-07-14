# Scenario

**Feature**: unknown top-level key `extra` is preserved across set-config

```
config.json has "extra": 1
wrk --set-config --create --new-terminal
  -> create.terminal written; extra still 1
```

## Steps

1. Seed config with version + `extra: 1`.
2. Run terminal-only set-config.

```go
func Setup(t *testing.T, req *Request) error {
	writeSetConfigRaw(t, req.WrkHome, `{
  "version": 1,
  "extra": 1
}
`)
	req.Args = setConfigArgs("--create", "--new-terminal")
	return nil
}
```
