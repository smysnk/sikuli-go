package grpcv1

import (
	"fmt"
	"path/filepath"
	"testing"

	pb "github.com/smysnk/sikuligo/internal/grpcv1/pb"
)

func TestBuildFindBenchFixtureFromRegionSpecUsesAllTargets(t *testing.T) {
	root := findBenchRepoRoot(".")
	if root == "" {
		root = filepath.Clean(filepath.Join("..", "..", ".."))
	}
	manifestPath := filepath.Join(root, "docs", "bench", "find-on-screen-scenarios.example.json")
	t.Setenv("FIND_BENCH_SCENARIO_MANIFEST", manifestPath)

	scenarios, _, err := loadFindBenchScenariosFromManifest(true, false)
	if err != nil {
		t.Fatalf("load manifest scenarios: %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatalf("expected at least one scenario from manifest")
	}

	source, queries, ok := buildFindBenchFixtureFromRegionSpec(t, scenarios[0])
	if !ok {
		t.Fatalf("expected region-spec fixture to be available")
	}
	if source == nil {
		t.Fatalf("expected source image")
	}
	if got, wantMin := len(queries), 3; got < wantMin {
		t.Fatalf("expected at least %d region queries, got %d", wantMin, got)
	}
	seenExpected := map[string]struct{}{}
	for i, query := range queries {
		if query.Pattern == nil {
			t.Fatalf("query[%d] missing pattern", i)
		}
		if query.Expected == nil || query.Expected.GetW() <= 0 || query.Expected.GetH() <= 0 {
			t.Fatalf("query[%d] missing expected rect", i)
		}
		key := rectKey(query.Expected)
		seenExpected[key] = struct{}{}
	}
	if got, wantMin := len(seenExpected), 3; got < wantMin {
		t.Fatalf("expected at least %d unique expected rects, got %d", wantMin, got)
	}
}

func rectKey(r *pb.Rect) string {
	if r == nil {
		return "nil"
	}
	return fmt.Sprintf("%d:%d:%d:%d", r.GetX(), r.GetY(), r.GetW(), r.GetH())
}
