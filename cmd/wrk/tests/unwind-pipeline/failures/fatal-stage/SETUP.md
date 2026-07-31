# Scenario

**Feature**: a hard post-land failure stops later peel work and reinstall tail

```
first linked peel without push remote -> push failure -> no following peel or reinstall
```

```go
import "github.com/xhd2015/doctest/session"
func Setup(t *testing.T,d *session.Doctest,req *Request) error { _=d; seedLinkedDep(t,req); git(t,req.DepMain,"remote","remove","origin"); req.Args=unwindGenCommitArgs(t,req,"--merge-back","--sync","--tag-next","--push","--reinstall-local"); return nil }
```
