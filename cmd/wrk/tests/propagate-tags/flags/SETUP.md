# Scenario

**Feature**: --dry-run host validation includes --propagate-tags once the flag lands

```
# bare --dry-run is still invalid without a host mode
wrk --dry-run -> non-zero; stderr lists valid hosts including --propagate-tags
```

## Preconditions

- Host list currently includes done / merge-back / all-deps / tag-next / sync;
 implementer extends it with `--propagate-tags`.

## Steps

1. Leaves exercise dry-run without `--propagate-tags` (and without other hosts).

## Context

- Assertion is deliberately loose on peer host ordering but requires
 `--propagate-tags` to appear in the error text.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Flag-validation subtree; leaves exercise bare --dry-run host list.
	propTagsEnsureHelpersUsed()
	return nil
}
```

