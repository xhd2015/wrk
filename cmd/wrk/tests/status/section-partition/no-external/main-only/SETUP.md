# Scenario

**Feature**: main-only status inputs → single primary entry, no external

```
# scan=[main], linked=[]
PartitionStatusPaths(main, [main], [])
  -> Primary=[main], External=[]
```

## Steps

1. Set `MainRoot` to the synthetic main path.
2. Set `ScanPaths` to `[main]` only.
3. Set `LinkedOrdered` empty.
4. Expect primary=`[main]`, external=`[]`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	main := pathMain()
	req.MainRoot = main
	req.ScanPaths = []string{main}
	req.LinkedOrdered = []string{}
	req.WantPrimary = []string{main}
	req.WantExternal = []string{}
	return nil
}
```
