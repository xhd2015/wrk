# Scenario

**Feature**: root and nested sub numeric tags map to correct require versions

```
# sourceMain: example.com/src (.) + example.com/src/sub (sub)
# tags: v1.2.3, sub/v0.1.0
ResolveSourceReleases(sourceMain)
  -> Releases: [
       {ModulePath=example.com/src, Tag=v1.2.3, Version=v1.2.3},
       {ModulePath=example.com/src/sub, Tag=sub/v0.1.0, Version=v0.1.0},
     ]
  -> Missing=[]
```

## Steps

1. Create source repo with root + nested `sub/` modules under `example.com/src`.
2. Tag HEAD with lightweight tags `v1.2.3` and `sub/v0.1.0`.
3. Set SourceMain to the repo path.
4. Expect both SourceRelease entries with Tag/Version mapping as above; Missing empty.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	src := filepath.Join(req.WorkRoot, "repos", "src")
	initRootAndSubModuleRepo(t, src, "example.com/src")
	tagRepo(t, src, "v1.2.3")
	tagRepo(t, src, "sub/v0.1.0")
	src = resolvePath(t, src)

	req.SourceMain = src
	req.WantReleases = []WantRelease{
		{ModulePath: "example.com/src", Tag: "v1.2.3", Version: "v1.2.3"},
		{ModulePath: "example.com/src/sub", Tag: "sub/v0.1.0", Version: "v0.1.0"},
	}
	req.WantMissing = []string{}
	return nil
}
```
