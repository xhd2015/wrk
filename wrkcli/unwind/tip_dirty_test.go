package unwind

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestScopeGitPathspec(t *testing.T) {
	t.Parallel()
	cases := []struct {
		dir  string
		want string
	}{
		{"", ""},
		{".", ""},
		{"go-pkgs", "go-pkgs"},
		{"go-pkgs/", "go-pkgs"},
		{"go-pkgs/cmd", "go-pkgs/cmd"},
	}
	for _, tc := range cases {
		if got := scopeGitPathspec(tc.dir); got != tc.want {
			t.Fatalf("scopeGitPathspec(%q)=%q want %q", tc.dir, got, tc.want)
		}
	}
}

func TestWillUseAddAllTip(t *testing.T) {
	t.Parallel()
	if willUseAddAllTip(UnwindFlags{AddAll: true, GenCommitMsg: true, GenCommitArgs: []string{"--commit"}}) != true {
		t.Fatal("want true for add-all+gen-commit+commit")
	}
	if willUseAddAllTip(UnwindFlags{AddAll: true, GenCommitMsg: true}) {
		t.Fatal("want false without --commit")
	}
	if willUseAddAllTip(UnwindFlags{GenCommitMsg: true, GenCommitArgs: []string{"--commit"}}) {
		t.Fatal("want false without AddAll")
	}
}

func initTipRepo(t *testing.T) (repo, releaseTag string) {
	t.Helper()
	repo = t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	if err := os.MkdirAll(filepath.Join(repo, "go-pkgs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go-pkgs", "doc.go"), []byte("package gopkgs\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "go-pkgs/doc.go")
	run("commit", "-m", "init")
	releaseTag = "go-pkgs/v0.0.1"
	run("tag", releaseTag)
	return repo, releaseTag
}

func TestTipScopeDirtyModeAStagedOnly(t *testing.T) {
	repo, tag := initTipRepo(t)
	prefix := "go-pkgs"

	dirty, err := tipScopeDirty(repo, tag, prefix, false)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("clean index: Mode A want not dirty")
	}

	if err := os.WriteFile(filepath.Join(repo, "go-pkgs", "lib.go"), []byte("package gopkgs\nconst X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirty, err = tipScopeDirty(repo, tag, prefix, false)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("untracked only: Mode A must ignore ??")
	}

	cmd := exec.Command("git", "add", "go-pkgs/lib.go")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	dirty, err = tipScopeDirty(repo, tag, prefix, false)
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Fatal("staged lib.go: Mode A want dirty")
	}
}

func TestTipScopeDirtyModeBIncludesUntracked(t *testing.T) {
	repo, tag := initTipRepo(t)
	prefix := "go-pkgs"

	if err := os.WriteFile(filepath.Join(repo, "go-pkgs", "lib.go"), []byte("package gopkgs\nconst X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirty, err := tipScopeDirty(repo, tag, prefix, true)
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Fatal("untracked under go-pkgs: Mode B (add-all) want dirty")
	}

	// Outside scope: root file must not dirty go-pkgs prefix.
	if err := os.WriteFile(filepath.Join(repo, "ROOT.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(repo, "go-pkgs", "lib.go"))
	dirty, err = tipScopeDirty(repo, tag, prefix, true)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("root-only untracked: Mode B must not dirty go-pkgs prefix")
	}
}

func TestRefreshNextTagsAndJobPlanAddAll(t *testing.T) {
	repo, tag := initTipRepo(t)
	if err := os.WriteFile(filepath.Join(repo, "go-pkgs", "lib.go"), []byte("package gopkgs\nconst X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	members := []StackMember{{
		Path: repo, MainRepo: repo, Label: "dep", Dirty: true, Linked: true,
	}}
	nodes := []UnwindGraphModuleNode{{
		Path: "example.com/dep/go-pkgs", Dir: "go-pkgs", RepoLabel: "dep",
		LatestTag: tag, SkipReason: "same-commit",
	}}

	// Without add-all tip: ?? ignored → still no NextTag.
	n1 := append([]UnwindGraphModuleNode(nil), nodes...)
	refreshNextTagsFromWorktreeTip(n1, members, false)
	if n1[0].NextTag != "" {
		t.Fatalf("no add-all: NextTag=%q want empty", n1[0].NextTag)
	}

	// With add-all tip: ?? under go-pkgs → NextTag.
	n2 := append([]UnwindGraphModuleNode(nil), nodes...)
	refreshNextTagsFromWorktreeTip(n2, members, true)
	if !strings.HasPrefix(n2[0].NextTag, "go-pkgs/v0.0.") || n2[0].NextTag == tag {
		t.Fatalf("add-all: NextTag=%q want bumped go-pkgs/v0.0.*", n2[0].NextTag)
	}
	if !n2[0].OwnedChanged || n2[0].SkipReason != "" {
		t.Fatalf("add-all: owned=%v skip=%q", n2[0].OwnedChanged, n2[0].SkipReason)
	}

	snap := &Snapshot{
		WorkDir:     repo,
		Inv:         StackInventory{Members: members},
		Peel:        &UnwindPlan{PeelOrder: []string{"dep"}, NeedsLand: true},
		ModuleNodes: nodes,
		ModuleEdges: nil,
		Cascade:     &UnwindCascadePlan{},
	}
	flags := UnwindFlags{
		TagNext: true, MergeBack: true, GenCommitMsg: true, AddAll: true,
		GenCommitArgs: []string{"--commit"},
	}
	job := BuildJobPlan(snap, flags)
	if job.ActionGraph == nil {
		t.Fatal("expected action graph")
	}
	found := false
	for _, a := range job.ActionGraph.Actions {
		if a.Mode == ModeTagNext && strings.Contains(a.ID, "go-pkgs") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("BuildJobPlan with add-all want tag-next for go-pkgs, actions=%v", actionIDs(job.ActionGraph))
	}
}

func TestClearNextTagsOnCleanRepos(t *testing.T) {
	t.Parallel()
	nodes := []UnwindGraphModuleNode{
		{Path: "example.com/clean", RepoLabel: "clean", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
		{Path: "example.com/dirty", RepoLabel: "dirty", NextTag: "v1.0.1", LatestTag: "v1.0.0", OwnedChanged: true},
	}
	members := []StackMember{
		{Label: "clean", Dirty: false},
		{Label: "dirty", Dirty: true},
	}
	clearNextTagsOnCleanRepos(nodes, members)
	if nodes[0].NextTag != "" || nodes[0].OwnedChanged {
		t.Fatalf("clean: next=%q owned=%v", nodes[0].NextTag, nodes[0].OwnedChanged)
	}
	if nodes[0].SkipReason != "no-changes" {
		t.Fatalf("clean skip=%q want no-changes", nodes[0].SkipReason)
	}
	if nodes[1].NextTag != "v1.0.1" || !nodes[1].OwnedChanged {
		t.Fatalf("dirty must keep tag: next=%q owned=%v", nodes[1].NextTag, nodes[1].OwnedChanged)
	}
}

func TestRefreshNextTagsSkipsCleanMember(t *testing.T) {
	repo, tag := initTipRepo(t)
	if err := os.WriteFile(filepath.Join(repo, "go-pkgs", "lib.go"), []byte("package gopkgs\nconst X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	members := []StackMember{{
		Path: repo, MainRepo: repo, Label: "dep", Dirty: false, Linked: true,
	}}
	nodes := []UnwindGraphModuleNode{{
		Path: "example.com/dep/go-pkgs", Dir: "go-pkgs", RepoLabel: "dep",
		LatestTag: tag, SkipReason: "same-commit",
	}}
	refreshNextTagsFromWorktreeTip(nodes, members, true)
	if nodes[0].NextTag != "" {
		t.Fatalf("clean member: NextTag=%q want empty", nodes[0].NextTag)
	}
}

func TestBuildJobPlanSkipsTagNextOnCleanRepo(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/dep", MainRepo: "/tmp/dep-main", Label: "dep", Dirty: false, Linked: true},
			{Path: "/tmp/app", MainRepo: "/tmp/app-main", Label: "app", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"app"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/dep", RepoLabel: "dep", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
			{Path: "example.com/dep/cmd", RepoLabel: "dep", LatestTag: "v0.0.1"},
			{Path: "example.com/app", RepoLabel: "app", NextTag: "v1.0.1", LatestTag: "v1.0.0", OwnedChanged: true},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/app", To: "example.com/dep", Kind: "require", Version: "v0.0.1"},
			{From: "example.com/dep/cmd", To: "example.com/dep", Kind: "require", Version: "v0.0.1"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/dep/cmd", DepModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/app", DepModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadeTagNext, ModulePath: "example.com/app", TagOrVersion: "v1.0.1"},
		}},
	}
	job := BuildJobPlan(snap, UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true, Push: true, Sync: true})
	if findAction(job.ActionGraph, "tag-next:example.com/dep") != nil {
		t.Fatalf("clean dep must not tag-next, actions=%v", actionIDs(job.ActionGraph))
	}
	if findAction(job.ActionGraph, "pin:example.com/app<example.com/dep") != nil {
		t.Fatalf("no propagate pin onto a tag we will not create, actions=%v", actionIDs(job.ActionGraph))
	}
	if findAction(job.ActionGraph, "merge-back:dep") != nil || findAction(job.ActionGraph, "push:dep") != nil {
		t.Fatalf("clean dep must not land/ship, actions=%v", actionIDs(job.ActionGraph))
	}
	if findAction(job.ActionGraph, "tag-next:example.com/app") == nil {
		t.Fatalf("dirty app must still tag-next, actions=%v", actionIDs(job.ActionGraph))
	}
	if ph := phaseByID(job.Phases, "modules"); ph != nil && ph.ByRepo != nil {
		if findAction(ph.ByRepo["dep"], "pin:example.com/dep/cmd<example.com/dep") != nil {
			t.Fatal("phase-2 must not pin onto a skipped NextTag")
		}
	}
}

const sampleGoMod = `module example.com/m

go 1.19

require example.com/dep v1.0.0
`

func TestGoModBytesReleaseDirty(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		after string
		want  bool
	}{
		{"identical", sampleGoMod, false},
		{"comment only", sampleGoMod + "// local checkout note\n", false},
		{"added replace", sampleGoMod + "\nreplace example.com/dep => ../dep\n", false},
		{"added replace plus extra", sampleGoMod + "\nreplace example.com/other => ../other\n", false},
		{"indirect bit ignored", `module example.com/m

go 1.19

require example.com/dep v1.0.0 // indirect
`, false},
		{"require bump", `module example.com/m

go 1.19

require example.com/dep v1.0.1
`, true},
		{"require bump and added replace", `module example.com/m

go 1.19

require example.com/dep v1.0.1

replace example.com/dep => ../dep
`, true},
		{"go version", `module example.com/m

go 1.21

require example.com/dep v1.0.0
`, true},
		{"module path", `module example.com/other

go 1.19

require example.com/dep v1.0.0
`, true},
		{"extra require", sampleGoMod + "require example.com/extra v0.0.1\n", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := goModBytesReleaseDirty([]byte(sampleGoMod), []byte(tc.after))
			if got != tc.want {
				t.Fatalf("got %v want %v\nafter:\n%s", got, tc.want, tc.after)
			}
		})
	}

	beforeWithReplace := sampleGoMod + "\nreplace example.com/dep => ../dep\n"
	t.Run("removed replace", func(t *testing.T) {
		t.Parallel()
		if !goModBytesReleaseDirty([]byte(beforeWithReplace), []byte(sampleGoMod)) {
			t.Fatal("removing a prior replace must be dirty")
		}
	})
	t.Run("changed replace target", func(t *testing.T) {
		t.Parallel()
		after := sampleGoMod + "\nreplace example.com/dep => ../other\n"
		if !goModBytesReleaseDirty([]byte(beforeWithReplace), []byte(after)) {
			t.Fatal("changing replace target must be dirty")
		}
	})
	t.Run("unparseable fallback", func(t *testing.T) {
		t.Parallel()
		a := []byte{0x00, 0x01, 'a'}
		b := []byte{0x00, 0x01, 'b'}
		if !goModBytesReleaseDirty(a, b) {
			t.Fatal("unparseable unequal blobs must be dirty")
		}
		if goModBytesReleaseDirty(a, a) {
			t.Fatal("unparseable equal blobs must not be dirty")
		}
	})
}

func TestTipPathsReleaseDirtyNameFilter(t *testing.T) {
	t.Parallel()
	dirty, err := tipPathsReleaseDirty("/no-repo", "v0.0.1", []string{"pkgs/shared/go.sum"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("go.sum only must not be release-dirty")
	}
	dirty, err = tipPathsReleaseDirty("/no-repo", "v0.0.1", nil, true)
	if err != nil || dirty {
		t.Fatalf("empty names dirty=%v err=%v", dirty, err)
	}
	dirty, err = tipPathsReleaseDirty("/no-repo", "v0.0.1", []string{"pkgs/shared/lib.go"}, true)
	if err != nil || !dirty {
		t.Fatalf("other file dirty=%v err=%v want true", dirty, err)
	}
	dirty, err = tipPathsReleaseDirty("/no-repo", "v0.0.1", []string{"a/go.mod", "b/go.mod"}, true)
	if err != nil || !dirty {
		t.Fatalf("two go.mod dirty=%v err=%v want true", dirty, err)
	}
}

func TestTipScopeDirtyIgnoresSumAndAddedReplace(t *testing.T) {
	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	modDir := filepath.Join(repo, "pkgs", "shared")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modDir, "go.mod"), []byte(sampleGoMod), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "pkgs/shared/go.mod")
	run("commit", "-m", "init")
	tag := "pkgs/shared/v0.0.1"
	run("tag", tag)
	prefix := "pkgs/shared"

	if err := os.WriteFile(filepath.Join(modDir, "go.sum"), []byte("example.com/dep v1.0.0 h1:abc=\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirty, err := tipScopeDirty(repo, tag, prefix, true)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("go.sum only: want not dirty")
	}

	after := sampleGoMod + "\nreplace example.com/dep => ../external/dep\n"
	if err := os.WriteFile(filepath.Join(modDir, "go.mod"), []byte(after), 0o644); err != nil {
		t.Fatal(err)
	}
	dirty, err = tipScopeDirty(repo, tag, prefix, true)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("added-only replace + go.sum: want not dirty")
	}

	// Mode A staged: same files unstaged must stay clean; staged replace-only still clean.
	dirty, err = tipScopeDirty(repo, tag, prefix, false)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("unstaged replace-only: Mode A want not dirty")
	}
	run("add", "pkgs/shared/go.mod", "pkgs/shared/go.sum")
	dirty, err = tipScopeDirty(repo, tag, prefix, false)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("staged added-only replace + go.sum: Mode A want not dirty")
	}

	bumped := `module example.com/m

go 1.19

require example.com/dep v1.0.1
`
	if err := os.WriteFile(filepath.Join(modDir, "go.mod"), []byte(bumped), 0o644); err != nil {
		t.Fatal(err)
	}
	dirty, err = tipScopeDirty(repo, tag, prefix, true)
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Fatal("require bump: want dirty")
	}
}
