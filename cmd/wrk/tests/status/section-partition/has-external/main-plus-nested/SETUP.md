# Scenario

**Feature**: nested scan path not linked-to-main becomes the sole external entry

```
# scan=[main, nested], linked=[]
PartitionStatusPaths(main, [main, nested], [])
  -> Primary=[main], External=[nested]
```

## Steps

1. Scan includes main and one nested independent repo (`task-hub`).
2. Linked list empty.
3. Expect primary=`[main]`, external=`[nested]`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	main := pathMain()
	nested := pathNestedTaskHub()
	req.MainRoot = main
	req.ScanPaths = []string{main, nested}
	req.LinkedOrdered = []string{}
	req.WantPrimary = []string{main}
	req.WantExternal = []string{nested}
	return nil
}
```
