# Scenario

**Feature**: create help co-present with a UX flag still shows help and does not write config

```
workspace/ -> wrk --set-config --create --new-window --help
  -> dedicated create usage on stdout, exit 0
  -> config.json still absent (no merge)
```

## Steps

1. Start with empty WRK_HOME (no config.json).
2. Run `wrk --set-config --create --new-window --help`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = setConfigArgs("--create", "--new-window", "--help")
	return nil
}
```
