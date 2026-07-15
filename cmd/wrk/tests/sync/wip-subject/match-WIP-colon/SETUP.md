# Scenario

**Feature**: subject with lowercase `wip:` prefix is WIP

```
IsWipSubject("wip: half done") -> true
```

## Steps

1. Set subject to `wip: half done`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Subject = "wip: half done"
	return nil
}
```
