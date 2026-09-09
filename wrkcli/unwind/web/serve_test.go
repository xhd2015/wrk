package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/wrk/wrkcli/unwind"
)

func TestIndexAndPlan(t *testing.T) {
	dir := initGitDir(t)
	h, err := NewHandler(Options{WorkDir: dir, Flags: unwind.UnwindFlags{MergeBack: true}})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("GET / status %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type %q", ct)
	}
	buf := make([]byte, 1<<20)
	n, _ := res.Body.Read(buf)
	body := string(buf[:n])
	for _, m := range []string{
		"<html", "wrk unwind", "Run", "--tag-next",
		"--gen-commit-msg", "--commit", "--add-all", "--no-verify",
		`data-flag="gen_commit_msg"`, `data-flag="commit"`, `data-flag="add_all"`,
		`data-action-graph`, `data-action-svg`, "renderActionGraph",
		"epochNodes", `data-action-project`, "aproj",
		"data-plan-loading", "READY_POLL_MS", "Building unwind plan",
		`[data-plan-loading="1"]`, // spinner DOM reused across polls
		"renderActionDAG", `data-flag="cleanup"`, "action_graph",
		"no actions in plan", "action-arr",
		"orderLanes", "LANE_GUTTER", `data-lane-gutter`, `data-lane-band`,
		`data-matrix-corner`, "project",
		"expandCommitActions", "assignTopoRanksClient",
		"awaitingPlan", "maxIters", "renderPhases", "data-phases",
		"phase-1 · cross-repo unwind", "phase-2 · intra-repo update", "repo-acc",
		"no nested module requires a Phase 1 tag", "det.open = true",
		"wrapLanes", "opts.wrapLanes",
		"renderModuleGraphInto", `data-phase-id`, "renderSnapshotPhase",
		`"snapshot"`, `"ship"`, "push + sync", "show replace edges",
		"dep depth", "medge", "MNODE_W = 320",
		"moduleShortLabel", `dir === "."`, "isRoot", "moduleIsUnchanged", "dimmed",
		"consumer → dep", `data-module-dimmed`,
		"arrows to unchanged deps hidden", "dimmedDep",

		`data-rank-matrix`,
		"gen-commit-msg", "add-all", "lane_levels", "laneLevels",
		"add-all fans out to pins and gen-commit-msg", "dep-update", "pinAfterAdd", "pinJoinLast",
		"Msg does not wait on pins",
		"bindEdgeHover", "bindNodeHover", "applyGraphHL", "clearGraphHL",
		"aedge-hit", "medge-hit", ".aedge.hl", ".anode.hl",
	} {
		if !strings.Contains(strings.ToLower(body), strings.ToLower(m)) && !strings.Contains(body, m) {
			t.Fatalf("GET / missing %q", m)
		}
	}
	// Shell embed: status + flags only (no full action graph).
	const marker = "/*__UNWIND_JSON__*/"
	mi := strings.Index(body, marker)
	if mi < 0 {
		t.Fatal("GET / missing INITIAL JSON marker")
	}
	rest := body[mi+len(marker):]
	end := strings.Index(rest, ";")
	if end < 0 {
		t.Fatal("GET / INITIAL JSON not terminated")
	}
	embed := rest[:end]
	if strings.Contains(embed, `"epochs":[{`) || strings.Contains(embed, `"epochs": [{`) {
		t.Fatal("GET / embed must not include populated epochs; use /plan")
	}
	if strings.Contains(embed, `"action_graph"`) {
		t.Fatal("GET / embed must not include action_graph; use /plan")
	}
	if !strings.Contains(embed, `"status":"ready"`) {
		t.Fatal("GET / sync NewHandler embed should be status ready")
	}
	if !strings.Contains(embed, `"flags"`) {
		t.Fatal("GET / shell embed missing flags")
	}

	pres, err := http.Get(ts.URL + "/plan?merge_back=1")
	if err != nil {
		t.Fatal(err)
	}
	defer pres.Body.Close()
	if pres.StatusCode != 200 {
		t.Fatalf("GET /plan status %d", pres.StatusCode)
	}
	var planOut struct {
		Status string `json:"status"`
		unwind.JobPlan
	}
	if err := json.NewDecoder(pres.Body).Decode(&planOut); err != nil {
		t.Fatal(err)
	}
	if planOut.Status != "ready" {
		t.Fatalf("status=%q want ready", planOut.Status)
	}
	if !planOut.Flags.MergeBack {
		t.Fatal("expected merge_back from query")
	}
}

func TestLoadingStatusBeforeSnapshot(t *testing.T) {
	dir := initGitDir(t)
	s := newServer(Options{WorkDir: dir, Flags: unwind.UnwindFlags{MergeBack: true}})
	ts := httptest.NewServer(s.mux())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	buf := make([]byte, 1<<20)
	n, _ := res.Body.Read(buf)
	body := string(buf[:n])
	if !strings.Contains(body, `"status":"loading"`) {
		t.Fatal("GET / before collect should embed status loading")
	}
	if !strings.Contains(body, `"merge_back":true`) {
		t.Fatal("loading embed should still expose CLI flags")
	}

	pres, err := http.Get(ts.URL + "/plan")
	if err != nil {
		t.Fatal(err)
	}
	defer pres.Body.Close()
	var out struct {
		Status string `json:"status"`
		CanRun bool   `json:"can_run"`
		Epochs []any  `json:"epochs"`
	}
	if err := json.NewDecoder(pres.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Status != "loading" {
		t.Fatalf("status=%q want loading", out.Status)
	}
	if out.CanRun {
		t.Fatal("loading plan must not can_run")
	}
	if len(out.Epochs) != 0 {
		t.Fatalf("loading epochs=%v want empty", out.Epochs)
	}

	runRes, err := http.Post(ts.URL+"/run", "application/json", strings.NewReader(`{"merge_back":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer runRes.Body.Close()
	if runRes.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("POST /run while loading status %d want 503", runRes.StatusCode)
	}

	if err := s.collectSnapshot(); err != nil {
		t.Fatal(err)
	}
	pres2, err := http.Get(ts.URL + "/plan?merge_back=1")
	if err != nil {
		t.Fatal(err)
	}
	defer pres2.Body.Close()
	var ready struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(pres2.Body).Decode(&ready); err != nil {
		t.Fatal(err)
	}
	if ready.Status != "ready" {
		t.Fatalf("after collect status=%q", ready.Status)
	}
}

func TestIndexReflectsGenCommitCLIFlags(t *testing.T) {
	dir := initGitDir(t)
	h, err := NewHandler(Options{
		WorkDir: dir,
		Flags: unwind.UnwindFlags{
			MergeBack:     true,
			GenCommitMsg:  true,
			GenCommitArgs: []string{"--commit", "--add-all", "--agent-runner=commandcode"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	buf := make([]byte, 1<<20)
	n, _ := res.Body.Read(buf)
	body := string(buf[:n])
	// Embedded INITIAL.plan.flags must carry peeled gen-commit toggles.
	for _, m := range []string{
		`"gen_commit_msg":true`,
		`"commit":true`,
		`"add_all":true`,
		`"agent_runner":"commandcode"`,
	} {
		if !strings.Contains(body, m) {
			t.Fatalf("GET / embedded flags missing %s", m)
		}
	}
}

func TestPostPlanFullFlagsAndHint(t *testing.T) {
	dir := initGitDir(t)
	h, err := NewHandler(Options{WorkDir: dir, Flags: unwind.UnwindFlags{MergeBack: true}})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	// Unchecking merge_back must clear it even when the server started with MergeBack.
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/plan", strings.NewReader(`{"merge_back":false,"tag_next":true,"push":false}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("POST /plan status %d", res.StatusCode)
	}
	var plan unwind.JobPlan
	if err := json.NewDecoder(res.Body).Decode(&plan); err != nil {
		t.Fatal(err)
	}
	if plan.Flags.MergeBack {
		t.Fatal("POST body must clear merge_back (not inherit server opts)")
	}
	if !plan.Flags.TagNext {
		t.Fatal("expected tag_next from POST body")
	}
	if plan.Graph == nil || plan.Graph.Summary.Repos < 1 {
		t.Fatalf("expected graph.summary.repos >= 1, graph=%#v", plan.Graph)
	}
	if plan.Graph.Repos.Nodes == nil {
		t.Fatal("expected graph.repos.nodes slice")
	}
	if plan.Graph.Modules.Nodes == nil {
		t.Fatal("expected graph.modules.nodes slice")
	}
	raw, _ := json.Marshal(plan)
	for _, key := range []string{`"summary"`, `"repos"`, `"modules"`, `"peel_order"`, `"nodes"`, `"edges"`} {
		if !strings.Contains(string(raw), key) {
			t.Fatalf("graph JSON missing %s in %s", key, raw)
		}
	}
}

func TestRunRejectsWhenBlocked(t *testing.T) {
	dir := initGitDir(t)
	// Dirty linked is hard; NeedsLand with no land flags: mark as linked via
	// JobFlags empty on a stack that needs land. Single main is not linked.
	// Force a blocker by posting empty flags against a peel that needs land:
	// create a dummy snapshot is internal. Use merge_back=false on linked-less
	// repo: CanRun is true. Instead POST with tag_next on a cycle-less clean
	// repo is allowed. Use Validate path: HasPendingEdges requires tag+push.
	h, err := NewHandler(Options{WorkDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	// Cross-repo edges aren't present; CanRun is true. 400 path: send
	// job flags that still fail gen-commit without --commit.
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/run", strings.NewReader(`{"gen_commit_msg":true}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST /run status %d want 400", res.StatusCode)
	}
}

func initGitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "README")
	run("commit", "-m", "init")
	return dir
}
