# Scenario

**Feature**: module without numeric release tag is listed Missing; others still resolve

```
# sourceMain: root + sub modules
# tags: only v1.0.0 on root (sub has no numeric tag)
ResolveSourceReleases
  -> Releases: [{example.com/src, v1.0.0, v1.0.0}]
  -> Missing: [example.com/src/sub]
```

## Steps

1. Create source repo with root `example.com/src` and nested `example.com/src/sub`.
2. Tag only `v1.0.0` (root). Do **not** create `sub/v…` tags.
3. Expect root in Releases; sub module path in Missing; overall err nil.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	src := filepath.Join(req.WorkRoot, "repos", "src")
	initRootAndSubModuleRepo(t, src, "example.com/src")
	tagRepo(t, src, "v1.0.0")
	// Intentionally no sub/v* tag.
	src = resolvePath(t, src)

	req.SourceMain = src
	req.WantReleases = []WantRelease{
		{ModulePath: "example.com/src", Tag: "v1.0.0", Version: "v1.0.0"},
	}
	req.WantMissing = []string{"example.com/src/sub"}
	return nil
}
```
