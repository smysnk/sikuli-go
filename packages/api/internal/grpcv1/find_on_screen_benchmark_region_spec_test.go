package grpcv1

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type findBenchRegionSpecDocument struct {
	SchemaVersion string                     `json:"schema_version"`
	Images        []findBenchRegionSpecImage `json:"images"`
}

type findBenchRegionSpecPreviewImage struct {
	Label     string `json:"label"`
	ImagePath string `json:"image_path"`
}

type findBenchRegionSpecImage struct {
	ID              string   `json:"id"`
	ScenarioTypeIDs []string `json:"scenario_type_ids"`
	// ImagePath is a legacy alias for SourceImagePath.
	ImagePath          string                            `json:"image_path"`
	SourceImagePath    string                            `json:"source_image_path"`
	BenchmarkImagePath string                            `json:"benchmark_image_path"`
	PreviewImages      []findBenchRegionSpecPreviewImage `json:"preview_images"`
	Targets            []findBenchRegionSpecTarget       `json:"targets"`
	BenchmarkTargets   []findBenchRegionSpecTarget       `json:"benchmark_targets,omitempty"`
}

type findBenchRegionSpecTarget struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	W     int    `json:"w"`
	H     int    `json:"h"`
}

type findBenchScenarioRegionAssignment struct {
	SourceImagePath    string
	BenchmarkImagePath string
	Targets            []findBenchTargetRegion
	BenchmarkTargets   []findBenchTargetRegion
}

func applyFindBenchScenarioRegions(scenarios []findBenchScenario, manifestSpecPath string, manifestPath string) ([]findBenchScenario, string, int, error) {
	raw := strings.TrimSpace(os.Getenv("FIND_BENCH_REGION_SPEC"))
	explicitEnv := raw != ""
	if raw == "" {
		raw = strings.TrimSpace(manifestSpecPath)
	}
	explicitManifest := raw != ""
	if raw == "" {
		raw = filepath.Join("packages", "api", "internal", "grpcv1", "testdata", "find-bench-assets", "regions.json")
	}

	doc, specPath, err := loadFindBenchRegionSpec(raw, filepath.Dir(strings.TrimSpace(manifestPath)))
	if err != nil {
		if !explicitEnv && !explicitManifest {
			return scenarios, "", 0, nil
		}
		return nil, "", 0, fmt.Errorf("load region spec %s: %w", raw, err)
	}

	assignments, err := buildFindBenchRegionAssignments(doc, filepath.Dir(specPath))
	if err != nil {
		return nil, "", 0, fmt.Errorf("build region assignments from %s: %w", specPath, err)
	}

	out := make([]findBenchScenario, len(scenarios))
	copy(out, scenarios)
	mapped := 0
	for i := range out {
		typeID := strings.TrimSpace(out[i].scenarioTypeID)
		if typeID == "" {
			continue
		}
		assignment, ok := assignments[typeID]
		if !ok || assignment.SourceImagePath == "" || len(assignment.Targets) == 0 {
			continue
		}
		out[i].sourceImagePath = assignment.SourceImagePath
		out[i].benchmarkImagePath = assignment.BenchmarkImagePath
		out[i].sourceTargets = append([]findBenchTargetRegion(nil), assignment.Targets...)
		out[i].benchmarkTargets = append([]findBenchTargetRegion(nil), assignment.BenchmarkTargets...)
		mapped++
	}

	return out, specPath, mapped, nil
}

func loadFindBenchRegionSpec(rawPath string, baseDir string) (*findBenchRegionSpecDocument, string, error) {
	resolved, ok := resolveFindBenchAnyPathIfExists(rawPath, baseDir)
	if !ok {
		return nil, "", fmt.Errorf("file not found: %s", rawPath)
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", resolved, err)
	}
	var doc findBenchRegionSpecDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", resolved, err)
	}
	if strings.TrimSpace(doc.SchemaVersion) == "" {
		return nil, "", fmt.Errorf("%s missing schema_version", resolved)
	}
	if len(doc.Images) == 0 {
		return nil, "", fmt.Errorf("%s has no images", resolved)
	}
	return &doc, resolved, nil
}

func buildFindBenchRegionAssignments(doc *findBenchRegionSpecDocument, specDir string) (map[string]findBenchScenarioRegionAssignment, error) {
	assignments := map[string]findBenchScenarioRegionAssignment{}
	if doc == nil {
		return assignments, nil
	}
	for i, entry := range doc.Images {
		sourceRaw := strings.TrimSpace(entry.SourceImagePath)
		if sourceRaw == "" {
			sourceRaw = strings.TrimSpace(entry.ImagePath)
		}
		sourcePath, err := resolveFindBenchRegionImagePath(specDir, sourceRaw)
		if err != nil {
			return nil, fmt.Errorf("images[%d] (%s) source image path %q: %w", i, entry.ID, sourceRaw, err)
		}
		benchmarkPath := sourcePath
		benchmarkRaw := strings.TrimSpace(entry.BenchmarkImagePath)
		if benchmarkRaw != "" {
			resolvedBenchmarkPath, resolveErr := resolveFindBenchRegionImagePath(specDir, benchmarkRaw)
			if resolveErr != nil {
				return nil, fmt.Errorf("images[%d] (%s) benchmark_image_path %q: %w", i, entry.ID, benchmarkRaw, resolveErr)
			}
			benchmarkPath = resolvedBenchmarkPath
		}

		targets := parseFindBenchRegionTargets(entry.Targets)
		if len(targets) == 0 {
			continue
		}
		benchmarkTargets := parseFindBenchRegionTargets(entry.BenchmarkTargets)
		if len(benchmarkTargets) == 0 {
			benchmarkTargets = append([]findBenchTargetRegion(nil), targets...)
		}

		typeIDs := make([]string, 0, len(entry.ScenarioTypeIDs)+1)
		for _, raw := range entry.ScenarioTypeIDs {
			id := strings.TrimSpace(raw)
			if id == "" {
				continue
			}
			typeIDs = append(typeIDs, id)
		}
		if len(typeIDs) == 0 {
			if id := strings.TrimSpace(entry.ID); id != "" {
				typeIDs = append(typeIDs, id)
			}
		}
		for _, typeID := range typeIDs {
			assignments[typeID] = findBenchScenarioRegionAssignment{
				SourceImagePath:    sourcePath,
				BenchmarkImagePath: benchmarkPath,
				Targets:            append([]findBenchTargetRegion(nil), targets...),
				BenchmarkTargets:   append([]findBenchTargetRegion(nil), benchmarkTargets...),
			}
		}
	}
	return assignments, nil
}

func parseFindBenchRegionTargets(in []findBenchRegionSpecTarget) []findBenchTargetRegion {
	targets := make([]findBenchTargetRegion, 0, len(in))
	for j, t := range in {
		if t.W <= 0 || t.H <= 0 {
			continue
		}
		targetID := strings.TrimSpace(t.ID)
		if targetID == "" {
			targetID = fmt.Sprintf("target-%02d", j+1)
		}
		targets = append(targets, findBenchTargetRegion{
			ID:    targetID,
			Label: strings.TrimSpace(t.Label),
			X:     t.X,
			Y:     t.Y,
			W:     t.W,
			H:     t.H,
		})
	}
	return targets
}

func resolveFindBenchRegionImagePath(specDir, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty path")
	}
	candidates := make([]string, 0, 8)
	seen := map[string]struct{}{}
	add := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		candidates = append(candidates, clean)
	}
	if filepath.IsAbs(raw) {
		add(raw)
	} else {
		add(filepath.Join(specDir, raw))
		for _, c := range findBenchPathCandidates(raw) {
			add(c)
		}
	}
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		return path, nil
	}
	return "", fmt.Errorf("file not found (candidates=%s)", strings.Join(candidates, ", "))
}

func resolveFindBenchAnyPathIfExists(raw string, baseDir string) (string, bool) {
	candidates := make([]string, 0, 8)
	seen := map[string]struct{}{}
	add := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		candidates = append(candidates, clean)
	}
	if baseDir != "" && !filepath.IsAbs(raw) {
		add(filepath.Join(baseDir, raw))
	}
	for _, candidate := range findBenchPathCandidates(raw) {
		add(candidate)
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		return candidate, true
	}
	return "", false
}
