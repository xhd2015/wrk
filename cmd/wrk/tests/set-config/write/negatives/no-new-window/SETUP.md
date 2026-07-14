# Scenario

**Feature**: `--no-new-window` clears/off window axis

```
seed window+terminal -> wrk --set-config --create --no-new-window
  -> window cleared/off; terminal may remain
```

## Steps

1. Seed window.mode=new + terminal.mode=new.
2. Run `--no-new-window`.

```go
func Setup(t *testing.T, req *Request) error {
	writeSetConfigRaw(t, req.WrkHome, `{
  "version": 1,
  "create": {
    "window": { "mode": "new" },
    "terminal": { "mode": "new" }
  }
}
`)
	req.Args = setConfigArgs("--create", "--no-new-window")
	return nil
}
```
