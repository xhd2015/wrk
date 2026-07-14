# Scenario

**Feature**: `wrk --set-config --create -h` prints dedicated create usage

```
workspace/ -> wrk --set-config --create -h
  -> create UX usage on stdout, exit 0
```

## Steps

1. Run `wrk --set-config --create -h`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = setConfigArgs("--create", "-h")
	return nil
}
```
