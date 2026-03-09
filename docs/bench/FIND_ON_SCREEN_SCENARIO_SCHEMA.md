---
layout: guide
title: FindOnScreen Scenario Schema
nav_key: benchmarks
kicker: Benchmarks
lead: Manifest structure, region-selection workflow, and validation inputs behind the benchmark scenario corpus.
---

This schema defines a concrete manifest for generating benchmark scenarios across all scenario kinds:

- `vector_ui`
- `photographic`
- `repetitive_grid`
- `noise_stress`
- `scale_rotate`
- `perspective_skew`
- `orb_feature_rich`
- `template_control`
- `hybrid_gate`
- `multi_monitor_dpi`

## Files

- Schema: [find-on-screen-scenario.schema.json]({{ '/bench/find-on-screen-scenario.schema.json' | relative_url }})
- Example manifest: [find-on-screen-scenarios.example.json]({{ '/bench/find-on-screen-scenarios.example.json' | relative_url }})
- Region spec: `packages/api/internal/grpcv1/testdata/find-bench-assets/regions.json`

## Region Selection Workflow

Use the benchmark helper to define per-scenario image targets in real time:

```bash
cd packages/api
go run ./cmd/benchmark-helper -listen :8091
```

Open `http://127.0.0.1:8091` and:

- Navigate scenario images with sidebar or previous/next buttons.
- Draw one or more target regions.
- Drag existing regions to adjust placement.
- Delete a selected region with the delete button or keyboard delete key.
- Selections smaller than `50x50` are ignored with a toast error.

Edits are persisted live to `packages/api/internal/grpcv1/testdata/find-bench-assets/regions.json`.

Benchmarks consume this file automatically (or override path with `FIND_BENCH_REGION_SPEC`).

## Generate Scenario Assets

To regenerate scenario-specific image assets used by non-photo scenarios:

```bash
./scripts/clients/generate-bench-scenario-assets.py
```

Output folder:

- `packages/api/internal/grpcv1/testdata/find-bench-assets/scenario`

## Core sections

- `engines`: engine set to run (`template`, `orb`, `akaze`, `brisk`, `kaze`, `sift`, `hybrid`)
- `resolution_groups`: benchmark resolutions and DPI profiles
- `monitor_profiles`: multi-monitor simulation metadata
- `defaults`: baseline execution and acceptance thresholds
- `scenario_types`: per-kind generation rules (target, background, transforms, photometric, decoys, occlusion, expectations)
- `matrix`: which resolutions and scenario types are materialized, with per-resolution overrides
- `outputs`: report/image behavior (`summary_scale`, mega summary format/quality)

## Current benchmark mapping

This manifest maps directly to existing benchmark concepts in:

- `packages/api/internal/grpcv1/find_on_screen_benchmark_test.go`

Field mapping (current -> schema):

- `findBenchScenario.variant` -> `scenario_types[].style`
- `findBenchScenario.size` -> `defaults.target_size_px` + per-scenario `target.size_px`
- `findBenchScenario.rotation` -> `target.rotation_degrees`
- `findBenchScenario.screenW/screenH` -> `resolution_groups[].width/height`
- `findBenchScenario.tolerance` -> `defaults.tolerance_overlap_min` / `expected.iou_min`
- `findBenchScenario.maxAreaRatio` -> `defaults.max_area_ratio` / `expected.area_ratio_max`
- `findBenchScenario.transformKind/transformA/B/C` -> `transforms.*`
- `findBenchScenario.queryFromBase` -> `defaults.query_from_base`

## Validation

Quick local validation:

```bash
jq empty docs/bench/find-on-screen-scenario.schema.json
jq empty docs/bench/find-on-screen-scenarios.example.json
```

At runtime, benchmark loading now performs strict JSON Schema validation before materializing scenarios:

- manifest path env: `FIND_BENCH_SCENARIO_MANIFEST`
- optional schema override env: `FIND_BENCH_SCENARIO_SCHEMA`
  - default schema path: `docs/bench/find-on-screen-scenario.schema.json`
- optional region spec env: `FIND_BENCH_REGION_SPEC`
  - default region spec path: `packages/api/internal/grpcv1/testdata/find-bench-assets/regions.json`

## Next wiring slice

1. Add a parser that loads the manifest from `FIND_BENCH_SCENARIO_MANIFEST`.
2. Materialize `findBenchScenario` rows from `matrix` + `scenario_types` + `resolution_groups`.
3. Keep current hardcoded scenario pack as fallback when no manifest is provided.
