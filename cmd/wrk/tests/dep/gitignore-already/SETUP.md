# Scenario

**Feature**: wrk --dep does not duplicate `/external` when already in .gitignore

```
# consumer .gitignore already has /external -> wrk --dep -> single /external line
```

## Steps

1. Create consumer with `/external` already in `.gitignore`.
2. Create dep repo and run `wrk --dep`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	writeFile(t, filepath.Join(consumer, ".gitignore"), "/external\n")
	dep := initDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.Args = []string{"--dep", dep}
	return nil
}
```