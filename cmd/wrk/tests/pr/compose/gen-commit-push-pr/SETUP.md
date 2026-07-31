# Scenario

**Feature**: fixed order stages — gen-commit-msg → push → pr

```
# staged change on linked feature; hermetic commandcode mock
linked wt + staged compose-stage.go + github origin + fake gh
  -> wrk --gen-commit-msg --commit --agent-runner commandcode --agent-runner-binary <mock>
         --push --pr --title T --comment C
  -> 1) commit with mock title "feat: compose pr"
  -> 2) full push of branch tip
  -> 3) PR create + comment
```

## Steps

- Leaf stages file, installs commandcode mock + fake gh, sets compose argv.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
