package grpcv1

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

type findBenchManifest struct {
	SchemaVersion      string                          `json:"schema_version"`
	Name               string                          `json:"name"`
	Seed               int                             `json:"seed"`
	ScenarioRegionSpec string                          `json:"scenario_region_spec,omitempty"`
	Metadata           findBenchManifestMetadata       `json:"metadata"`
	Defaults           findBenchManifestDefaults       `json:"defaults"`
	ResolutionGroups   []findBenchManifestResolution   `json:"resolution_groups"`
	MonitorProfiles    []findBenchManifestMonitor      `json:"monitor_profiles"`
	ScenarioTypes      []findBenchManifestScenarioType `json:"scenario_types"`
	Matrix             findBenchManifestMatrix         `json:"matrix"`
	Outputs            findBenchManifestOutputs        `json:"outputs"`
}

type findBenchManifestMetadata struct {
	Owner      string   `json:"owner"`
	CreatedUTC string   `json:"created_utc"`
	Notes      string   `json:"notes"`
	Tags       []string `json:"tags"`
}

type findBenchManifestDefaults struct {
	TargetSizePx        int     `json:"target_size_px"`
	Iterations          int     `json:"iterations"`
	RetryAttempts       int     `json:"retry_attempts"`
	RPCTimeoutMs        int     `json:"rpc_timeout_ms"`
	QueryFromBase       bool    `json:"query_from_base"`
	ToleranceOverlapMin float64 `json:"tolerance_overlap_min"`
	MaxAreaRatio        float64 `json:"max_area_ratio"`
	StrictBBox          bool    `json:"strict_bbox"`
}

type findBenchManifestResolution struct {
	ID       string  `json:"id"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	DPIScale float64 `json:"dpi_scale"`
	Enabled  *bool   `json:"enabled,omitempty"`
}

type findBenchManifestMonitor struct {
	ID                    string  `json:"id"`
	Label                 string  `json:"label"`
	Width                 int     `json:"width"`
	Height                int     `json:"height"`
	DPIScale              float64 `json:"dpi_scale"`
	Gamma                 float64 `json:"gamma"`
	ColorTemperatureShift int     `json:"color_temperature_shift"`
	Sharpness             float64 `json:"sharpness"`
}

type findBenchManifestIntRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type findBenchManifestFloatRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type findBenchManifestScenarioType struct {
	ID                 string                                        `json:"id"`
	Kind               string                                        `json:"kind"`
	Enabled            *bool                                         `json:"enabled,omitempty"`
	Style              string                                        `json:"style"`
	Target             findBenchManifestTarget                       `json:"target"`
	Background         findBenchManifestBackground                   `json:"background"`
	Transforms         findBenchManifestTransforms                   `json:"transforms"`
	Photometric        findBenchManifestPhotometric                  `json:"photometric"`
	Occlusion          findBenchManifestOcclusion                    `json:"occlusion"`
	Decoys             findBenchManifestDecoys                       `json:"decoys"`
	MonitorSelector    findBenchManifestMonitorSel                   `json:"monitor_selector"`
	HybridPolicy       findBenchManifestHybridPolicy                 `json:"hybrid_policy"`
	EngineExpectations map[string]findBenchManifestEngineExpectation `json:"engine_expectations"`
	Expected           findBenchManifestExpected                     `json:"expected"`
}

type findBenchManifestTarget struct {
	Source          string                      `json:"source"`
	AssetPool       []string                    `json:"asset_pool"`
	SizePx          findBenchManifestIntRange   `json:"size_px"`
	RotationDegrees findBenchManifestFloatRange `json:"rotation_degrees"`
	AspectJitter    float64                     `json:"aspect_jitter"`
	SubpixelOffset  bool                        `json:"subpixel_offset"`
}

type findBenchManifestBackground struct {
	ContinuousCanvas bool    `json:"continuous_canvas"`
	ClutterDensity   float64 `json:"clutter_density"`
	Palette          string  `json:"palette"`
	TextureSeed      int     `json:"texture_seed"`
}

type findBenchManifestTransforms struct {
	Scale       findBenchManifestFloatRange         `json:"scale"`
	Rotate      findBenchManifestFloatRange         `json:"rotate"`
	SkewX       findBenchManifestFloatRange         `json:"skew_x"`
	SkewY       findBenchManifestFloatRange         `json:"skew_y"`
	Perspective findBenchManifestPerspectiveOptions `json:"perspective"`
}

type findBenchManifestPerspectiveOptions struct {
	Enabled        bool                        `json:"enabled"`
	CornerShiftPct findBenchManifestFloatRange `json:"corner_shift_pct"`
}

type findBenchManifestExpected struct {
	Positive     bool    `json:"positive"`
	IoUMin       float64 `json:"iou_min"`
	AreaRatioMax float64 `json:"area_ratio_max"`
	AllowPartial bool    `json:"allow_partial"`
}

type findBenchManifestPhotometric struct {
	Brightness  findBenchManifestFloatRange  `json:"brightness"`
	Contrast    findBenchManifestFloatRange  `json:"contrast"`
	Gamma       findBenchManifestFloatRange  `json:"gamma"`
	BlurSigma   findBenchManifestFloatRange  `json:"blur_sigma"`
	JPEGQuality findBenchManifestIntRange    `json:"jpeg_quality"`
	Noise       []findBenchManifestNoiseSpec `json:"noise"`
}

type findBenchManifestNoiseSpec struct {
	Type   string                      `json:"type"`
	Amount findBenchManifestFloatRange `json:"amount"`
}

type findBenchManifestOcclusion struct {
	Enabled           bool                        `json:"enabled"`
	TargetCoveragePct findBenchManifestFloatRange `json:"target_coverage_pct"`
}

type findBenchManifestDecoys struct {
	Enabled    bool                        `json:"enabled"`
	Count      findBenchManifestIntRange   `json:"count"`
	Similarity findBenchManifestFloatRange `json:"similarity"`
	Placement  string                      `json:"placement"`
}

type findBenchManifestMonitorSel struct {
	Mode       string   `json:"mode"`
	MonitorIDs []string `json:"monitor_ids"`
}

type findBenchManifestHybridPolicy struct {
	MustConsiderAllEngines bool     `json:"must_consider_all_engines"`
	SelectBy               string   `json:"select_by"`
	FallbackOrder          []string `json:"fallback_order"`
}

type findBenchManifestEngineExpectation struct {
	MinSuccessRate       float64 `json:"min_success_rate"`
	MaxFalsePositiveRate float64 `json:"max_false_positive_rate"`
	MaxAvgLatencyMs      float64 `json:"max_avg_latency_ms"`
}

type findBenchManifestOutputs struct {
	WriteJSON     bool                            `json:"write_json"`
	WriteMarkdown bool                            `json:"write_markdown"`
	WriteImages   bool                            `json:"write_images"`
	PatchReadmes  bool                            `json:"patch_readmes"`
	SummaryScale  int                             `json:"summary_scale"`
	MegaSummary   findBenchManifestMegaSummaryOut `json:"mega_summary"`
}

type findBenchManifestMegaSummaryOut struct {
	Format      string `json:"format"`
	JPEGQuality int    `json:"jpeg_quality"`
}

type findBenchManifestMatrix struct {
	ResolutionGroupIDs     []string                                `json:"resolution_group_ids"`
	ScenarioTypeIDs        []string                                `json:"scenario_type_ids"`
	ScenariosPerResolution int                                     `json:"scenarios_per_resolution"`
	PerResolutionOverrides map[string]findBenchManifestResOverride `json:"per_resolution_overrides"`
}

type findBenchManifestResOverride struct {
	ScenarioTypeIDs     []string                   `json:"scenario_type_ids"`
	TargetSizePx        *findBenchManifestIntRange `json:"target_size_px,omitempty"`
	ToleranceOverlapMin *float64                   `json:"tolerance_overlap_min,omitempty"`
	MaxAreaRatio        *float64                   `json:"max_area_ratio,omitempty"`
}

func loadFindBenchScenariosFromManifest(highResEnabled, ultraResEnabled bool) ([]findBenchScenario, string, error) {
	raw := strings.TrimSpace(os.Getenv("FIND_BENCH_SCENARIO_MANIFEST"))
	if raw == "" {
		return nil, "", fmt.Errorf("FIND_BENCH_SCENARIO_MANIFEST is not set")
	}
	resolved, err := resolveFindBenchManifestPath(raw)
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, "", fmt.Errorf("read manifest %s: %w", resolved, err)
	}
	if err := validateFindBenchManifestStrict(data, resolved); err != nil {
		return nil, "", err
	}
	var manifest findBenchManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, "", fmt.Errorf("parse manifest %s: %w", resolved, err)
	}
	if err := validateFindBenchSchemaVersion(manifest.SchemaVersion); err != nil {
		return nil, "", fmt.Errorf("schema compatibility check failed for %s: %w", resolved, err)
	}
	if override := parseEnvIntStrict("FIND_BENCH_SEED"); override >= 0 {
		manifest.Seed = override
	}
	scenarios, err := materializeFindBenchScenarios(&manifest, highResEnabled, ultraResEnabled)
	if err != nil {
		return nil, "", err
	}
	scenarios, regionSpecPath, regionMapped, err := applyFindBenchScenarioRegions(scenarios, manifest.ScenarioRegionSpec, resolved)
	if err != nil {
		return nil, "", err
	}
	if regionSpecPath != "" {
		_, _ = fmt.Fprintf(os.Stderr, "[find-bench] scenario region spec=%s mapped=%d/%d\n", regionSpecPath, regionMapped, len(scenarios))
	}
	return scenarios, resolved, nil
}

func validateFindBenchSchemaVersion(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return fmt.Errorf("schema_version is empty")
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return fmt.Errorf("schema_version %q must be semver-like MAJOR.MINOR.PATCH", v)
	}
	major := parseIntDefault(parts[0], -1)
	if major < 0 {
		return fmt.Errorf("schema_version %q has invalid major", v)
	}
	const supportedMajor = 1
	if major > supportedMajor {
		_, _ = fmt.Fprintf(os.Stderr, "[find-bench] warning: manifest schema major=%d is newer than supported=%d\n", major, supportedMajor)
		return fmt.Errorf("unsupported future schema major %d (supported=%d)", major, supportedMajor)
	}
	if major < supportedMajor {
		return fmt.Errorf("unsupported old schema major %d (supported=%d)", major, supportedMajor)
	}
	return nil
}

func validateFindBenchManifestStrict(data []byte, manifestPath string) error {
	schemaPath, err := resolveFindBenchSchemaPath()
	if err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	schemaFile, err := os.Open(schemaPath)
	if err != nil {
		return fmt.Errorf("open manifest schema %s: %w", schemaPath, err)
	}
	defer func() { _ = schemaFile.Close() }()

	const schemaResource = "find-on-screen-scenario.schema.json"
	if err := compiler.AddResource(schemaResource, schemaFile); err != nil {
		return fmt.Errorf("load manifest schema %s: %w", schemaPath, err)
	}
	schema, err := compiler.Compile(schemaResource)
	if err != nil {
		return fmt.Errorf("compile manifest schema %s: %w", schemaPath, err)
	}

	var rawDoc any
	if err := json.Unmarshal(data, &rawDoc); err != nil {
		return fmt.Errorf("parse manifest JSON %s: %w", manifestPath, err)
	}
	if err := schema.Validate(rawDoc); err != nil {
		return fmt.Errorf("manifest %s failed strict schema validation (%s): %w", manifestPath, schemaPath, err)
	}
	return nil
}

func resolveFindBenchSchemaPath() (string, error) {
	raw := strings.TrimSpace(os.Getenv("FIND_BENCH_SCENARIO_SCHEMA"))
	if raw == "" {
		raw = filepath.Join("docs", "bench", "find-on-screen-scenario.schema.json")
	}
	resolved, err := resolveFindBenchAnyPath(raw)
	if err != nil {
		return "", fmt.Errorf("resolve manifest schema path: %w", err)
	}
	return resolved, nil
}

func resolveFindBenchManifestPath(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "example" || raw == "default" {
		raw = filepath.Join("docs", "bench", "find-on-screen-scenarios.example.json")
	}
	resolved, err := resolveFindBenchAnyPath(raw)
	if err != nil {
		return "", err
	}
	return resolved, nil
}

func resolveFindBenchAnyPath(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("path is empty")
	}

	candidates := findBenchPathCandidates(raw)

	for _, path := range candidates {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		return path, nil
	}
	return "", fmt.Errorf("file not found: %s (candidates=%s)", raw, strings.Join(candidates, ", "))
}

func findBenchPathCandidates(raw string) []string {
	candidates := make([]string, 0, 6)
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

	add(raw)
	if !filepath.IsAbs(raw) {
		wd, _ := os.Getwd()
		add(filepath.Join(wd, raw))
		add(filepath.Join(wd, "..", "..", raw))
		if root := findBenchRepoRoot(wd); root != "" {
			add(filepath.Join(root, raw))
		}
	}
	return candidates
}

func findBenchRepoRoot(start string) string {
	if start == "" {
		start = "."
	}
	if !filepath.IsAbs(start) {
		abs, err := filepath.Abs(start)
		if err == nil {
			start = abs
		}
	}
	dir := filepath.Clean(start)
	for {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func materializeFindBenchScenarios(manifest *findBenchManifest, highResEnabled, ultraResEnabled bool) ([]findBenchScenario, error) {
	if manifest == nil {
		return nil, fmt.Errorf("manifest is nil")
	}

	typeByID := map[string]findBenchManifestScenarioType{}
	typeOrder := make([]string, 0, len(manifest.ScenarioTypes))
	for _, st := range manifest.ScenarioTypes {
		id := strings.TrimSpace(st.ID)
		if id == "" || !boolOrDefault(st.Enabled, true) {
			continue
		}
		typeByID[id] = st
		typeOrder = append(typeOrder, id)
	}
	if len(typeByID) == 0 {
		return nil, fmt.Errorf("manifest has no enabled scenario_types")
	}

	resByID := map[string]findBenchManifestResolution{}
	resOrderDefault := make([]string, 0, len(manifest.ResolutionGroups))
	for _, rg := range manifest.ResolutionGroups {
		id := strings.TrimSpace(rg.ID)
		if id == "" {
			continue
		}
		resByID[id] = rg
		resOrderDefault = append(resOrderDefault, id)
	}
	if len(resByID) == 0 {
		return nil, fmt.Errorf("manifest has no resolution_groups")
	}
	monitorByID := map[string]findBenchManifestMonitor{}
	for _, mp := range manifest.MonitorProfiles {
		id := strings.TrimSpace(mp.ID)
		if id == "" {
			continue
		}
		monitorByID[id] = mp
	}

	matrixTypeIDs := filterKnownIDs(manifest.Matrix.ScenarioTypeIDs, typeByID)
	if len(matrixTypeIDs) == 0 {
		matrixTypeIDs = append(matrixTypeIDs, typeOrder...)
	}
	if len(matrixTypeIDs) == 0 {
		return nil, fmt.Errorf("manifest matrix has no usable scenario_type_ids")
	}

	matrixResolutionIDs := manifest.Matrix.ResolutionGroupIDs
	if len(matrixResolutionIDs) == 0 {
		matrixResolutionIDs = append(matrixResolutionIDs, resOrderDefault...)
	}
	scenariosPerResolution := manifest.Matrix.ScenariosPerResolution
	if scenariosPerResolution <= 0 {
		scenariosPerResolution = maxInt(1, len(matrixTypeIDs))
	}

	scenarios := make([]findBenchScenario, 0, len(matrixResolutionIDs)*scenariosPerResolution)
	for _, resID := range matrixResolutionIDs {
		rg, ok := resByID[resID]
		if !ok {
			continue
		}
		if !boolOrDefault(rg.Enabled, true) {
			continue
		}
		if !manifestResolutionAllowed(rg.Width, rg.Height, highResEnabled, ultraResEnabled) {
			continue
		}

		override := manifest.Matrix.PerResolutionOverrides[resID]
		candidateTypeIDs := matrixTypeIDs
		if len(override.ScenarioTypeIDs) > 0 {
			candidateTypeIDs = filterKnownIDs(override.ScenarioTypeIDs, typeByID)
		}
		if len(candidateTypeIDs) == 0 {
			continue
		}

		for i := 0; i < scenariosPerResolution; i++ {
			typeID := candidateTypeIDs[i%len(candidateTypeIDs)]
			st := typeByID[typeID]
			scenarios = append(scenarios, materializeFindBenchScenario(manifest, rg, st, override, monitorByID, i))
		}
	}
	if len(scenarios) == 0 {
		return nil, fmt.Errorf("manifest produced zero scenarios")
	}
	return scenarios, nil
}

func materializeFindBenchScenario(
	manifest *findBenchManifest,
	rg findBenchManifestResolution,
	st findBenchManifestScenarioType,
	override findBenchManifestResOverride,
	monitorByID map[string]findBenchManifestMonitor,
	index int,
) findBenchScenario {
	baseSize := manifest.Defaults.TargetSizePx
	if baseSize < 16 {
		baseSize = 96
	}
	sizeRange := st.Target.SizePx
	if override.TargetSizePx != nil {
		sizeRange = *override.TargetSizePx
	}
	seedKey := fmt.Sprintf("%d:%s:%d:%dx%d", manifest.Seed, st.ID, index, rg.Width, rg.Height)
	size := pickIntInRange(sizeRange, baseSize, seedKey+":size")
	if size < 16 {
		size = 16
	}

	rotation := pickQuarterRotation(st.Target.RotationDegrees, seedKey+":quarter")
	tolerance := manifest.Defaults.ToleranceOverlapMin
	if tolerance <= 0 {
		tolerance = 0.20
	}
	if st.Expected.IoUMin > 0 {
		tolerance = st.Expected.IoUMin
	}
	if override.ToleranceOverlapMin != nil {
		tolerance = *override.ToleranceOverlapMin
	}
	tolerance = clampFloat(tolerance, 0.0, 1.0)

	maxAreaRatio := manifest.Defaults.MaxAreaRatio
	if maxAreaRatio < 1.0 {
		maxAreaRatio = 1.50
	}
	if st.Expected.AreaRatioMax >= 1.0 {
		maxAreaRatio = st.Expected.AreaRatioMax
	}
	if override.MaxAreaRatio != nil && *override.MaxAreaRatio >= 1.0 {
		maxAreaRatio = *override.MaxAreaRatio
	}

	transformKind, transformA, transformB, transformC := chooseManifestTransform(st, index, seedKey)
	variant := manifestStyleToVariant(st.Style, st.Kind)
	if st.Target.SubpixelOffset && index%2 == 1 {
		rotation = (rotation + 90) % 360
	}

	monitorMode, monitorID, monitorGamma, monitorSharpness, monitorColorShift := chooseMonitorProfile(st.MonitorSelector, monitorByID, seedKey, index)
	noiseGaussian, noisePoisson, noiseSaltPepper, noiseBanding, noiseCompression := parseNoiseAmounts(st.Photometric.Noise, seedKey)
	jpegQuality := pickIntInRange(st.Photometric.JPEGQuality, 90, seedKey+":jpeg_quality")
	if jpegQuality <= 0 {
		jpegQuality = 90
	}

	retryAttempts := manifest.Defaults.RetryAttempts
	if retryAttempts < 0 {
		retryAttempts = 0
	}
	rpcTimeoutMs := manifest.Defaults.RPCTimeoutMs
	if rpcTimeoutMs <= 0 {
		rpcTimeoutMs = 5000
	}
	decoyCount := pickIntInRange(st.Decoys.Count, 0, seedKey+":decoy_count")
	decoySimilarity := pickFloatInRange(st.Decoys.Similarity, 0.90, seedKey+":decoy_similarity")
	occlusionCoverage := pickFloatInRange(st.Occlusion.TargetCoveragePct, 0.0, seedKey+":occlusion")
	clutter := st.Background.ClutterDensity
	if clutter <= 0 {
		clutter = 0.75
	}

	name := fmt.Sprintf("%s_%dx%d_i%02d", sanitizeBenchScenarioToken(st.ID), rg.Width, rg.Height, index+1)
	if transformKind != "" {
		name = fmt.Sprintf("%s_%s", name, sanitizeBenchScenarioToken(transformKind))
	}
	name = fmt.Sprintf("%s_s%08x", name, shortSeedToken(seedKey))
	targetSource := strings.ToLower(strings.TrimSpace(st.Target.Source))
	targetAsset := ""
	if strings.EqualFold(st.Kind, "photographic") || targetSource == "asset" || targetSource == "mixed" {
		targetAsset = chooseFindBenchTargetAssetPath(st.Target.AssetPool)
	}

	return findBenchScenario{
		name:           name,
		scenarioTypeID: st.ID,
		kind:           st.Kind,
		variant:        variant,
		targetSource:   targetSource,
		targetAsset:    targetAsset,
		size:           size,
		rotation:       rotation,
		screenW:        rg.Width,
		screenH:        rg.Height,
		tolerance:      tolerance,
		maxAreaRatio:   maxAreaRatio,
		transformKind:  transformKind,
		transformA:     transformA,
		transformB:     transformB,
		transformC:     transformC,
		// Manifest-backed benchmark mode always uses source-image regions as query patterns.
		queryFromBase: true,
		seed:          hashKey(seedKey),

		expectedPositive: st.Expected.Positive,
		allowPartial:     st.Expected.AllowPartial,
		strictBBox:       manifest.Defaults.StrictBBox,
		retryAttempts:    retryAttempts,
		rpcTimeout:       time.Duration(rpcTimeoutMs) * time.Millisecond,

		backgroundContinuousCanvas: st.Background.ContinuousCanvas,
		backgroundClutterDensity:   clampFloat(clutter, 0.0, 1.0),
		backgroundPalette:          st.Background.Palette,
		backgroundTextureSeed:      st.Background.TextureSeed,

		brightnessDelta:  pickFloatInRange(st.Photometric.Brightness, 0.0, seedKey+":brightness"),
		contrastFactor:   pickFloatInRange(st.Photometric.Contrast, 1.0, seedKey+":contrast"),
		gammaFactor:      pickFloatInRange(st.Photometric.Gamma, 1.0, seedKey+":gamma"),
		blurSigma:        pickFloatInRange(st.Photometric.BlurSigma, 0.0, seedKey+":blur_sigma"),
		jpegQuality:      jpegQuality,
		noiseGaussian:    noiseGaussian,
		noisePoisson:     noisePoisson,
		noiseSaltPepper:  noiseSaltPepper,
		noiseBanding:     noiseBanding,
		noiseCompression: noiseCompression,

		occlusionEnabled:  st.Occlusion.Enabled,
		occlusionCoverage: clampFloat(occlusionCoverage, 0.0, 0.95),
		decoyEnabled:      st.Decoys.Enabled,
		decoyCount:        maxInt(0, decoyCount),
		decoySimilarity:   clampFloat(decoySimilarity, 0.50, 0.999),
		decoyPlacement:    st.Decoys.Placement,

		monitorID:         monitorID,
		monitorMode:       monitorMode,
		monitorGamma:      monitorGamma,
		monitorSharpness:  monitorSharpness,
		monitorColorShift: monitorColorShift,

		hybridMustConsiderAll: st.HybridPolicy.MustConsiderAllEngines,
		hybridSelectBy:        st.HybridPolicy.SelectBy,
		hybridFallbackOrder:   append([]string{}, st.HybridPolicy.FallbackOrder...),
	}
}

func chooseManifestTransform(st findBenchManifestScenarioType, index int, key string) (string, float64, float64, float64) {
	rotate := pickFloatInRange(st.Transforms.Rotate, 0.0, key+":rotate")
	scale := pickFloatInRange(st.Transforms.Scale, 1.0, key+":scale")
	skewX := pickFloatInRange(st.Transforms.SkewX, 0.0, key+":skewx") / 100.0
	cornerShift := math.Abs(pickFloatInRange(st.Transforms.Perspective.CornerShiftPct, 0.0, key+":perspective"))
	perspectiveTop := maxFloat(0.20, 1.0-cornerShift)
	perspectiveBottom := 1.0 + cornerShift

	switch st.Kind {
	case "repetitive_grid":
		rotate = chooseDeterministicRotationLevel(
			st.Transforms.Rotate,
			key,
			index,
			[]float64{-15, -12, -9, -6, -3, 3, 6, 9, 12, 15},
			rotate,
		)
		if math.Abs(rotate) < 2.0 {
			rotate = chooseDeterministicRotationLevel(
				st.Transforms.Rotate,
				key+":force",
				index,
				[]float64{-15, -12, -9, -6, 6, 9, 12, 15},
				rotate,
			)
		}
		if math.Abs(rotate) > 0.5 {
			return "rotate", rotate, 0, 0
		}
		return "", 0, 0, 0
	case "scale_rotate":
		if index%2 == 0 && math.Abs(scale-1.0) > 0.02 {
			return "scale", maxFloat(0.05, scale), 0, 0
		}
		if math.Abs(rotate) < 2.0 {
			rotate = 12.0
		}
		return "rotate", rotate, 0, 0
	case "perspective_skew":
		if st.Transforms.Perspective.Enabled {
			return "perspective", perspectiveTop, perspectiveBottom, skewX
		}
		if math.Abs(skewX) > 0.004 {
			return "skewx", skewX, 0, 0
		}
		return "rotate", rotate, 0, 0
	case "orb_feature_rich":
		if st.Transforms.Perspective.Enabled && index%2 == 1 {
			return "perspective", perspectiveTop, perspectiveBottom, skewX
		}
		if math.Abs(rotate) < 5.0 {
			rotate = 20.0
		}
		return "rotate", rotate, 0, 0
	case "hybrid_gate":
		if strings.EqualFold(st.HybridPolicy.SelectBy, "iou_then_area") {
			if st.Transforms.Perspective.Enabled {
				return "perspective", perspectiveTop, perspectiveBottom, skewX
			}
		}
		switch index % 3 {
		case 0:
			if math.Abs(scale-1.0) > 0.02 {
				return "scale", maxFloat(0.05, scale), 0, 0
			}
		case 1:
			return "rotate", rotate, 0, 0
		default:
			if st.Transforms.Perspective.Enabled {
				return "perspective", perspectiveTop, perspectiveBottom, skewX
			}
		}
		return "rotate", rotate, 0, 0
	case "multi_monitor_dpi":
		if math.Abs(scale-1.0) > 0.03 {
			return "scale", maxFloat(0.05, scale), 0, 0
		}
		if math.Abs(rotate) > 2.0 {
			return "rotate", rotate, 0, 0
		}
		return "", 0, 0, 0
	default:
		if math.Abs(rotate) > 4.0 {
			return "rotate", rotate, 0, 0
		}
		if math.Abs(scale-1.0) > 0.03 {
			return "scale", maxFloat(0.05, scale), 0, 0
		}
		if st.Transforms.Perspective.Enabled && cornerShift > 0.01 {
			return "perspective", perspectiveTop, perspectiveBottom, skewX
		}
		if math.Abs(skewX) > 0.005 {
			return "skewx", skewX, 0, 0
		}
		return "", 0, 0, 0
	}
}

func chooseMonitorProfile(sel findBenchManifestMonitorSel, profiles map[string]findBenchManifestMonitor, key string, index int) (mode, id string, gamma float64, sharpness float64, colorShift int) {
	mode = strings.ToLower(strings.TrimSpace(sel.Mode))
	if mode == "" {
		mode = "single"
	}
	ids := make([]string, 0, len(sel.MonitorIDs))
	for _, rawID := range sel.MonitorIDs {
		id := strings.TrimSpace(rawID)
		if id == "" {
			continue
		}
		if _, ok := profiles[id]; ok {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		for id := range profiles {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return mode, "", 1.0, 1.0, 0
	}

	chosen := ids[0]
	switch mode {
	case "all", "round_robin":
		chosen = ids[(index+int(hashKey(key)%uint64(len(ids))))%len(ids)]
	case "single":
		chosen = ids[0]
	}
	mp := profiles[chosen]
	gamma = mp.Gamma
	if gamma <= 0 {
		gamma = 1.0
	}
	sharpness = mp.Sharpness
	if sharpness <= 0 {
		sharpness = 1.0
	}
	colorShift = mp.ColorTemperatureShift
	return mode, chosen, gamma, sharpness, colorShift
}

func parseNoiseAmounts(noise []findBenchManifestNoiseSpec, key string) (gaussian, poisson, saltPepper, banding, compression float64) {
	for i, n := range noise {
		amount := pickFloatInRange(n.Amount, 0.0, fmt.Sprintf("%s:noise:%d:%s", key, i, n.Type))
		switch strings.ToLower(strings.TrimSpace(n.Type)) {
		case "gaussian":
			gaussian = maxFloat(gaussian, amount)
		case "poisson":
			poisson = maxFloat(poisson, amount)
		case "salt_pepper":
			saltPepper = maxFloat(saltPepper, amount)
		case "banding":
			banding = maxFloat(banding, amount)
		case "compression_blocks":
			compression = maxFloat(compression, amount)
		}
	}
	return
}

func shortSeedToken(key string) uint32 {
	return uint32(hashKey(key) & 0xffffffff)
}

func chooseFindBenchTargetAssetPath(assetPool []string) string {
	if override := strings.TrimSpace(os.Getenv("FIND_BENCH_PHOTO_ASSET")); override != "" {
		if resolved, err := resolveFindBenchAnyPath(override); err == nil {
			return resolved
		}
	}

	for _, raw := range assetPool {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if strings.ContainsAny(raw, "*?[") {
			matches := resolveFindBenchGlobPaths(raw)
			if len(matches) > 0 {
				return matches[0]
			}
			continue
		}
		if resolved, err := resolveFindBenchAnyPath(raw); err == nil {
			return resolved
		}
	}
	return ""
}

func resolveFindBenchGlobPaths(raw string) []string {
	matches := make([]string, 0, 8)
	seen := map[string]struct{}{}
	for _, pattern := range findBenchPathCandidates(raw) {
		expanded, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, path := range expanded {
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				continue
			}
			clean := filepath.Clean(path)
			if _, ok := seen[clean]; ok {
				continue
			}
			seen[clean] = struct{}{}
			matches = append(matches, clean)
		}
	}
	sort.Strings(matches)
	return matches
}

func manifestStyleToVariant(style string, kind string) string {
	s := strings.ToLower(strings.TrimSpace(style))
	switch s {
	case "vector", "ui", "photo", "noise", "orbtex":
		return s
	case "grid":
		return "glyph"
	case "mixed":
		switch strings.ToLower(strings.TrimSpace(kind)) {
		case "orb_feature_rich":
			return "orbtex"
		case "template_control", "vector_ui":
			return "ui"
		default:
			return "photo"
		}
	default:
		return "photo"
	}
}

func manifestResolutionAllowed(width, height int, highResEnabled, ultraResEnabled bool) bool {
	if width >= 2560 || height >= 1440 {
		return ultraResEnabled
	}
	if width >= 1280 || height >= 720 {
		return highResEnabled
	}
	return true
}

func filterKnownIDs[V any](ids []string, known map[string]V) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := known[id]; !ok {
			continue
		}
		out = append(out, id)
	}
	return out
}

func boolOrDefault(v *bool, defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}
	return *v
}

func pickIntInRange(r findBenchManifestIntRange, fallback int, key string) int {
	minV, maxV := r.Min, r.Max
	if minV == 0 && maxV == 0 {
		return fallback
	}
	if maxV < minV {
		minV, maxV = maxV, minV
	}
	if minV == maxV {
		return minV
	}
	u := hashUnitInterval(key)
	return minV + int(math.Round(float64(maxV-minV)*u))
}

func pickFloatInRange(r findBenchManifestFloatRange, fallback float64, key string) float64 {
	minV, maxV := r.Min, r.Max
	if minV == 0 && maxV == 0 {
		return fallback
	}
	if maxV < minV {
		minV, maxV = maxV, minV
	}
	if minV == maxV {
		return minV
	}
	u := hashUnitInterval(key)
	return minV + (maxV-minV)*u
}

func chooseDeterministicRotationLevel(r findBenchManifestFloatRange, key string, index int, levels []float64, fallback float64) float64 {
	minV, maxV := r.Min, r.Max
	if maxV < minV {
		minV, maxV = maxV, minV
	}
	if minV == 0 && maxV == 0 {
		return fallback
	}

	allowed := make([]float64, 0, len(levels))
	for _, level := range levels {
		if level >= minV && level <= maxV {
			allowed = append(allowed, level)
		}
	}
	if len(allowed) == 0 {
		return clampFloat(pickFloatInRange(r, fallback, key+":sample"), minV, maxV)
	}
	pos := int(hashKey(fmt.Sprintf("%s:rotate-level:%d", key, index)) % uint64(len(allowed)))
	return allowed[pos]
}

func pickQuarterRotation(r findBenchManifestFloatRange, key string) int {
	candidates := []int{}
	for _, d := range []int{0, 90, 180, 270} {
		if float64(d) >= r.Min && float64(d) <= r.Max {
			candidates = append(candidates, d)
		}
	}
	if len(candidates) == 0 {
		return 0
	}
	idx := int(hashKey(key) % uint64(len(candidates)))
	return candidates[idx]
}

func hashUnitInterval(key string) float64 {
	return float64(hashKey(key)%10000) / 9999.0
}

func hashKey(key string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return h.Sum64()
}

func sanitizeBenchScenarioToken(in string) string {
	in = strings.ToLower(strings.TrimSpace(in))
	if in == "" {
		return "scenario"
	}
	var b strings.Builder
	b.Grow(len(in))
	prevUnderscore := false
	for _, r := range in {
		isAlpha := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if isAlpha || isDigit {
			b.WriteRune(r)
			prevUnderscore = false
			continue
		}
		if !prevUnderscore {
			b.WriteByte('_')
			prevUnderscore = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "scenario"
	}
	return out
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func parseIntDefault(raw string, def int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	n := 0
	sign := 1
	for i, r := range raw {
		if i == 0 && r == '-' {
			sign = -1
			continue
		}
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	return sign * n
}

func parseEnvIntStrict(name string) int {
	return parseIntDefault(os.Getenv(name), -1)
}
