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

func TestClassifyFindBenchPositiveMatch_RegionAware(t *testing.T) {
	primary := &pb.Rect{X: 100, Y: 120, W: 80, H: 60}
	secondary := &pb.Rect{X: 340, Y: 120, W: 80, H: 60}
	all := []*pb.Rect{primary, secondary}
	pattern := &pb.GrayImage{Width: 80, Height: 60}

	if got := classifyFindBenchPositiveMatch(
		&pb.Rect{X: 102, Y: 122, W: 80, H: 60},
		primary,
		all,
		pattern,
		0.6,
		1.5,
		false,
	); got != findBenchMatchClassOK {
		t.Fatalf("expected ok classification, got %q", got)
	}

	if got := classifyFindBenchPositiveMatch(
		&pb.Rect{X: 340, Y: 120, W: 80, H: 60},
		primary,
		all,
		pattern,
		0.6,
		1.5,
		false,
	); got != findBenchMatchClassWrongRegion {
		t.Fatalf("expected wrong_region classification, got %q", got)
	}

	if got := classifyFindBenchPositiveMatch(
		&pb.Rect{X: 10, Y: 10, W: 80, H: 60},
		primary,
		all,
		pattern,
		0.6,
		1.5,
		false,
	); got != findBenchMatchClassOverlapMiss {
		t.Fatalf("expected overlap_miss classification, got %q", got)
	}
}

func TestClassifyFindBenchPositiveMatch_ZoneExpected(t *testing.T) {
	zone := &pb.Rect{X: 500, Y: 320, W: 460, H: 278}
	peer := &pb.Rect{X: 1000, Y: 320, W: 460, H: 278}
	all := []*pb.Rect{zone, peer}
	pattern := &pb.GrayImage{Width: 160, Height: 98}

	if got := classifyFindBenchPositiveMatch(
		&pb.Rect{X: 602, Y: 410, W: 170, H: 110},
		zone,
		all,
		pattern,
		0.22,
		2.0,
		false,
	); got != findBenchMatchClassOK {
		t.Fatalf("expected zone match to classify as ok, got %q", got)
	}

	if got := classifyFindBenchPositiveMatch(
		&pb.Rect{X: 560, Y: 360, W: 520, H: 350},
		zone,
		all,
		pattern,
		0.22,
		2.0,
		false,
	); got != findBenchMatchClassOverlapMiss {
		t.Fatalf("expected oversized in-zone match to classify as overlap_miss, got %q", got)
	}
}

func rectKey(r *pb.Rect) string {
	if r == nil {
		return "nil"
	}
	return fmt.Sprintf("%d:%d:%d:%d", r.GetX(), r.GetY(), r.GetW(), r.GetH())
}
