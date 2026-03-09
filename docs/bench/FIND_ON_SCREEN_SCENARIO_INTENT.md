---
layout: guide
title: FindOnScreen Scenario Intent
nav_key: benchmarks
kicker: Benchmarks
lead: What each benchmark scenario is intended to validate, which failure modes it stresses, and what healthy matcher behavior should look like.
---

This document explains what each benchmark scenario is intended to validate.
Source manifest: [find-on-screen-scenarios.example.json]({{ '/bench/find-on-screen-scenarios.example.json' | relative_url }}).

## How to read this

- `Goal`: What reliability/performance behavior we are testing.
- `Stressors`: What kinds of visual or geometric difficulty are intentionally introduced.
- `What should pass`: Expected matcher behavior when the scenario is healthy.

## Scenario Types (Manifest)

### `vector_ui_baseline` (`kind=vector_ui`)

- Goal: Validate baseline precision on clean, UI-like shapes.
- Stressors: Flat fills, crisp edges, light photometric jitter, structured decoys.
- What should pass: High IoU and tight bounding boxes with low false positives.

### `photo_clutter` (`kind=photographic`)

- Goal: Validate matching against photographic content under clutter.
- Stressors: Real asset content, blur/noise/compression variation, mixed decoy placement, optional occlusion.
- What should pass: Stable positive matches even when local texture and photometric conditions drift.

### `repetitive_grid_camouflage` (`kind=repetitive_grid`)

- Goal: Detect true target among many near-repeating distractors.
- Stressors: Dense repeated structures with high similarity and camouflage-like layouts.
- What should pass: Correct localization with minimized false positives in repetitive fields.

### `noise_stress_random` (`kind=noise_stress`)

- Goal: Evaluate robustness under heavy corruption.
- Stressors: Gaussian + salt/pepper + compression artifacts, stronger photometric shifts, random decoy placement, occlusion.
- What should pass: Best-effort matching with controlled miss/false-positive rates in degraded inputs.

### `scale_rotate_sweep` (`kind=scale_rotate`)

- Goal: Measure scale and rotation tolerance.
- Stressors: Broad scale range and large rotation range, mixed decoys, moderate occlusion.
- What should pass: Engines that claim transform robustness should maintain usable success at nontrivial scale/rotation offsets.

### `perspective_skew_sweep` (`kind=perspective_skew`)

- Goal: Measure geometric invariance beyond pure rotation/scale.
- Stressors: Perspective corner shifts, skew, and additional photometric variation.
- What should pass: Feature-based/hybrid approaches should localize targets under projective distortions better than pure template-only modes.

### `orb_feature_rich` (`kind=orb_feature_rich`)

- Goal: Validate feature detector/descriptor strengths in textured scenes.
- Stressors: Feature-rich synthetic textures, wider transform/photometric variation, clustered decoys.
- What should pass: ORB-family/feature engines should outperform template in success rate and maintain acceptable latency.

### `template_control_exact` (`kind=template_control`)

- Goal: Control scenario for exact-ish template matching.
- Stressors: Minimal geometric change, constrained photometric variation, cleaner layout.
- What should pass: Template should perform strongly (high success, low false positives, low latency).

### `hybrid_gate_conflicts` (`kind=hybrid_gate`)

- Goal: Verify hybrid policy selection logic under conflicting evidence across engines.
- Stressors: Mixed transforms/noise/decoys where some engines may disagree.
- What should pass: Hybrid should avoid obvious misses when at least one underlying engine has a strong, geometrically valid candidate.

### `multi_monitor_dpi_shift` (`kind=multi_monitor_dpi`)

- Goal: Validate behavior across monitor profile differences (DPI/gamma/sharpness/color shift).
- Stressors: Round-robin monitor profiles with differing scale and photometric signatures.
- What should pass: Stable matching across profile changes without large regression in localization quality.

## Default Fallback Pack (when no manifest is provided)

Fallback scenarios are generated in `resolutionScenarioPack(...)` inside:
`packages/api/internal/grpcv1/find_on_screen_benchmark_test.go`.

Per resolution, six scenarios are generated:

- `vector_r0_<res>`: Baseline vector-like matching.
- `photo_r90_<res>`: Photo-style matching with mixed decoys.
- `ui_r180_<res>`: UI-style matching on darker palette.
- `mix_resize_<xx>_<res>` (down): Scale-down robustness.
- `mix_resize_<xx>_<res>` (up): Scale-up robustness.
- `mix_rotate_<deg>_<res>`: Rotation robustness with partial-acceptance behavior.

These fallback cases are intentionally simpler than manifest-driven scenarios and mainly serve as a deterministic baseline set.
