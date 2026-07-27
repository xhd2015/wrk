# Scenario

**Feature**: consumer require versions differ from source releases → would-update plan

```
# source releases newer than consumer require
app requires example.com/lib@older
  -> would: update example.com/app
  -> indented old -> new arrows
```

## Preconditions

- At least one consumer module has a require on a source module at a **different** version.
- Source has usable numeric release tags.

## Steps

1. Build source with release tags and a consumer requiring older versions.
2. Register both projects; run from source main.

## Context

- Footer `N modules` / `M projects` counts only planned updates.
- Replace overlay is a descendant leaf, not this grouping's only case.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Mark subtree: plan expects at least one version-differing require.
	// Leaf fixtures supply paths, go.mod requires, and snapshots.
	propTagsEnsureHelpersUsed()
	if len(req.Args) == 0 {
		req.Args = []string{"--propagate-tags", "--dry-run"}
	}
	return nil
}
```

