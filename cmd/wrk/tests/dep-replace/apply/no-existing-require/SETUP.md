# Scenario

**Feature**: replace does not require an existing require (D7)

```
consumer go.mod has no require for example.com/dep
  -> wrk --dep-replace <dep>
  -> absolute replace written
  -> exit 0
```

## Steps

1. Seed consumer without require line.
2. Apply replace for dep.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithDep(t, req, false) // no require
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```
