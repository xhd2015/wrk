# Scenario

**Feature**: wrk --bash-integration --install writes script and dual profile markers

```
fake HOME + WRK_HOME
wrk --bash-integration --install -> integration/bash.sh + markers in .bash_profile and .bashrc
```

## Steps

1. Set `req.Mode = "install"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "install"
	req.PreInstall = false
	return nil
}
```