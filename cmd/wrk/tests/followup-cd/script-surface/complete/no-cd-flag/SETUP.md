# Scenario

**Feature**: completion for `wrk -<tab>` includes --no-cd

```
wrk --bash-integration --complete -- wrk - 1 -> candidates include --no-cd
```

## Steps

1. Complete flag prefix `-` at word index 1.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CompleteWords = []string{"wrk", "-"}
	req.CompleteCWord = 1
	return nil
}
```
