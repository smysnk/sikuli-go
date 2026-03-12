# sikuli-go (Python)
<!-- DOCS_CANONICAL_TARGET: docs/python-client/index.md -->
<!-- DOCS_SOURCE_MAP: docs/strategy/documentation-source-map.md -->
<!-- DOCS_WORKFLOW: docs/contribution/docs-workflow.md -->

This directory contains the Python client for sikuli-go with Sikuli-style `Screen` + `Pattern` APIs.

## Canonical Documentation

Long-form Python docs now live in the published guide:

- [Python Client](https://smysnk.github.io/sikuli-go/python-client/)
- [Installation](https://smysnk.github.io/sikuli-go/python-client/installation)
- [First Script](https://smysnk.github.io/sikuli-go/python-client/first-script)
- [Runtime](https://smysnk.github.io/sikuli-go/python-client/runtime)
- [Troubleshooting](https://smysnk.github.io/sikuli-go/python-client/troubleshooting)
- [Getting Help](https://smysnk.github.io/sikuli-go/getting-help/)

## Quickstart

`init:py-examples` prompts for the target directory, creates `requirements.txt`, installs into `.venv`, and copies examples.
Each example bootstraps `sikuli-go` into `./.sikuli-go/bin` and prepends it to PATH for the process.
The published package name is `sikuli-go`; the import module remains `sikuligo`.

```bash
pipx run sikuli-go init:py-examples
cd sikuli-go-demo
python3 examples/click.py
```

runs:
```python
from __future__ import annotations
from sikuligo import Pattern, Screen

screen = Screen()
try:
    match = screen.click(Pattern("assets/pattern.png").exact())
    print(f"clicked match target at ({match.target_x}, {match.target_y})")
finally:
    screen.close()
```

## Install Into An Existing Environment

If you manage your own environment instead of using the scaffold:

```bash
python -m pip install sikuli-go
```

Package name vs import name:

- install: `sikuli-go`
- import: `sikuligo`

## Runtime Helpers

Install the runtime binaries on PATH:

```bash
pipx run sikuli-go install-binary
```

Start the runtime and dashboard:

```bash
pipx run sikuli-go -listen
```

Open:

- http://127.0.0.1:8080/dashboard

Run the standalone monitor after installing the binaries:

```bash
sikuli-go-monitor
```

By default it serves the monitor UI on `:8080` and reads `sikuli-go.db` from the current working directory.

## Dashboard Preview

The local runtime exposes both the dashboard and the standalone monitor workflow described in the guide:

- [Dashboard Guide](https://smysnk.github.io/sikuli-go/getting-started/dashboard)

![Dashboard Screenshot](https://smysnk.github.io/sikuli-go/images/dashboard.png)

![Monitor Screenshot](https://smysnk.github.io/sikuli-go/images/monitor.png)

## Common Package Entry Points

- `Screen()` or `Screen.start()` uses auto mode
- `Screen.connect()` attaches to an existing runtime
- `Screen.spawn()` forces a new runtime process
- `screen.region(x, y, w, h)` scopes the search area

For the full runtime model, environment inputs, and troubleshooting notes, use the Python guide pages above.

<!-- BEGIN: FIND_ON_SCREEN_BENCH_AUTOGEN -->
## FindOnScreen Benchmark Test Results

Generated: `2026-03-07T23:32:15.506029+00:00`

### Reports

- [Markdown Summary](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-e2e)
- [JSON Report](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-e2e.json)
- [Raw go test Output](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-e2e.txt)
- [Performance SVG](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-performance.svg)
- [Accuracy SVG](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-accuracy.svg)
- [Scenario Kind Match Time SVG](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-kind-time.svg)
- [Scenario Kind Success SVG](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-kind-success.svg)
- [Resolution Match Time SVG](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-time.svg)
- [Resolution Matches SVG](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-matches.svg)
- [Resolution Misses SVG](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-misses.svg)
- [Resolution False Positives SVG](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-false-positives.svg)

### Engine Summary

_Cases/OK metrics are query-level counts (regions x scenarios x resolutions), not just benchmark row count._

| Engine | Cases | OK | Partial | Not Found | Unsupported | Error | Overlap Miss | Avg ms/op | Median ms/op |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| akaze | 120 | 39 | 0 | 78 | 0 | 0 | 3 | 172.121 | 147.695 |
| brisk | 120 | 47 | 0 | 63 | 0 | 0 | 10 | 388.483 | 123.118 |
| hybrid | 120 | 69 | 0 | 45 | 0 | 0 | 6 | 171.017 | 134.411 |
| kaze | 120 | 63 | 0 | 50 | 0 | 0 | 7 | 824.898 | 640.512 |
| orb | 120 | 13 | 0 | 96 | 0 | 0 | 11 | 56.443 | 44.794 |
| sift | 120 | 56 | 0 | 55 | 0 | 0 | 9 | 256.756 | 198.264 |
| template | 120 | 64 | 0 | 56 | 0 | 0 | 0 | 154.257 | 114.466 |

### Run Mega Summary

![Run Mega Summary](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-run-mega.jpg)

- [Open run mega summary image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-run-mega.jpg)

### Benchmark Graphs

![Performance Graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-performance.svg)

- [Open performance graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-performance.svg)

![Accuracy Graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-accuracy.svg)

- [Open accuracy graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-accuracy.svg)

### Scenario Kind Graphs

![Scenario Kind Match Time](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-kind-time.svg)

- [Open scenario kind match time graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-kind-time.svg)

![Scenario Kind Success](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-kind-success.svg)

- [Open scenario kind success graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-kind-success.svg)

### Resolution Group Graphs

![Resolution Match Time](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-time.svg)

- [Open resolution match time graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-time.svg)

![Resolution Matches](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-matches.svg)

- [Open resolution matches graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-matches.svg)

![Resolution Misses](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-misses.svg)

- [Open resolution misses graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-misses.svg)

![Resolution False Positives](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-false-positives.svg)

- [Open resolution false positives graph](https://smysnk.github.io/sikuli-go/bench/reports/find-on-screen-resolution-false-positives.svg)

### Artifact Directories

- [Visual Root](https://smysnk.github.io/sikuli-go/bench/reports/visuals/)
- [Scenario Summaries](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/)

### Scenario Summary Images (10)

#### `hybrid_gate_conflicts_1920x1080_i09`

![hybrid_gate_conflicts_1920x1080_i09](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-hybrid_gate_conflicts_1920x1080_i09.png)

- [Open `hybrid_gate_conflicts_1920x1080_i09` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-hybrid_gate_conflicts_1920x1080_i09.png)

#### `multi_monitor_dpi_shift_1920x1080_i10`

![multi_monitor_dpi_shift_1920x1080_i10](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-multi_monitor_dpi_shift_1920x1080_i10.png)

- [Open `multi_monitor_dpi_shift_1920x1080_i10` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-multi_monitor_dpi_shift_1920x1080_i10.png)

#### `noise_stress_random_1920x1080_i04`

![noise_stress_random_1920x1080_i04](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-noise_stress_random_1920x1080_i04.png)

- [Open `noise_stress_random_1920x1080_i04` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-noise_stress_random_1920x1080_i04.png)

#### `orb_feature_rich_1920x1080_i07`

![orb_feature_rich_1920x1080_i07](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-orb_feature_rich_1920x1080_i07.png)

- [Open `orb_feature_rich_1920x1080_i07` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-orb_feature_rich_1920x1080_i07.png)

#### `perspective_skew_sweep_1920x1080_i06`

![perspective_skew_sweep_1920x1080_i06](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-perspective_skew_sweep_1920x1080_i06.png)

- [Open `perspective_skew_sweep_1920x1080_i06` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-perspective_skew_sweep_1920x1080_i06.png)

#### `photo_clutter_1920x1080_i02`

![photo_clutter_1920x1080_i02](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-photo_clutter_1920x1080_i02.png)

- [Open `photo_clutter_1920x1080_i02` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-photo_clutter_1920x1080_i02.png)

#### `repetitive_grid_camouflage_1920x1080_i03`

![repetitive_grid_camouflage_1920x1080_i03](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-repetitive_grid_camouflage_1920x1080_i03.png)

- [Open `repetitive_grid_camouflage_1920x1080_i03` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-repetitive_grid_camouflage_1920x1080_i03.png)

#### `scale_rotate_sweep_1920x1080_i05`

![scale_rotate_sweep_1920x1080_i05](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-scale_rotate_sweep_1920x1080_i05.png)

- [Open `scale_rotate_sweep_1920x1080_i05` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-scale_rotate_sweep_1920x1080_i05.png)

#### `template_control_exact_1920x1080_i08`

![template_control_exact_1920x1080_i08](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-template_control_exact_1920x1080_i08.png)

- [Open `template_control_exact_1920x1080_i08` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-template_control_exact_1920x1080_i08.png)

#### `vector_ui_baseline_1920x1080_i01`

![vector_ui_baseline_1920x1080_i01](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-vector_ui_baseline_1920x1080_i01.png)

- [Open `vector_ui_baseline_1920x1080_i01` image](https://smysnk.github.io/sikuli-go/bench/reports/visuals/summaries/summary-vector_ui_baseline_1920x1080_i01.png)

<!-- END: FIND_ON_SCREEN_BENCH_AUTOGEN -->
