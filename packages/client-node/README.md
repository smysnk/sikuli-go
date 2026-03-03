# sikuligo (Node.js)

SikuliGO is a GoLang implementation of Sikuli visual automation. This package provides the Node.js SDK for launching `sikuligo` locally and executing automation with a small API surface.

## Links

- Main repository: [github.com/smysnk/SikuliGO](https://github.com/smysnk/SikuliGO)
- API reference: [smysnk.github.io/SikuliGO/api](https://smysnk.github.io/SikuliGO/api/)
- Node user flow: [smysnk.github.io/SikuliGO/node-package-user-flow](https://smysnk.github.io/SikuliGO/node-package-user-flow)
- Client strategy: [smysnk.github.io/SikuliGO/client-strategy](https://smysnk.github.io/SikuliGO/client-strategy)
- Architecture docs: [Port Strategy](https://smysnk.github.io/SikuliGO/port-strategy), [gRPC Strategy](https://smysnk.github.io/SikuliGO/grpc-strategy)

## Quickstart

`init:js-examples` prompts for a target directory, scaffolds a `package.json` with the latest `@sikuligo/sikuligo` dependency, runs `yarn install`, and copies `.mjs` examples into `examples/`.

```bash
yarn dlx @sikuligo/sikuligo init:js-examples
cd sikuligo-demo
yarn node examples/click.mjs
```

```js
import { Screen, Pattern } from "@sikuligo/sikuligo";

const screen = await Screen();
try {
  const match = await screen.click(Pattern("assets/pattern.png").exact());
  console.log(`clicked match target at (${match.targetX}, ${match.targetY})`);
} finally {
  await screen.close();
}
```

## Web Dashboard
```bash
yarn dlx @sikuligo/sikuligo -listen 127.0.0.1:50051 -admin-listen :8080
```

Open:

- http://127.0.0.1:8080/dashboard

Additional endpoints:

- http://127.0.0.1:8080/healthz
- http://127.0.0.1:8080/metrics
- http://127.0.0.1:8080/snapshot

Install permanently on PATH:

```bash
yarn dlx @sikuligo/sikuligo install-binary
source ~/.zshrc
# or
source ~/.bash_profile
```

<!-- BEGIN: FIND_ON_SCREEN_BENCH_AUTOGEN -->
## FindOnScreen Benchmark Test Results

Generated: `2026-03-03T03:21:00.700542+00:00`

### Reports

- [Markdown Summary](../../.test-results/bench/find-on-screen-e2e.md)
- [JSON Report](../../.test-results/bench/find-on-screen-e2e.json)
- [Raw go test Output](../../.test-results/bench/find-on-screen-e2e.txt)
- [Performance SVG](../../.test-results/bench/find-on-screen-performance.svg)
- [Accuracy SVG](../../.test-results/bench/find-on-screen-accuracy.svg)
- [Resolution Match Time SVG](../../.test-results/bench/find-on-screen-resolution-time.svg)
- [Resolution Matches SVG](../../.test-results/bench/find-on-screen-resolution-matches.svg)
- [Resolution Misses SVG](../../.test-results/bench/find-on-screen-resolution-misses.svg)
- [Resolution False Positives SVG](../../.test-results/bench/find-on-screen-resolution-false-positives.svg)

### Engine Summary

| Engine | Cases | OK | Partial | Not Found | Unsupported | Error | Overlap Miss | Avg ms/op | Median ms/op |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| hybrid | 15 | 15 | 0 | 0 | 0 | 0 | 0 | 13.238 | 11.158 |
| orb | 15 | 4 | 0 | 10 | 0 | 0 | 1 | 6.565 | 7.334 |
| template | 15 | 11 | 0 | 4 | 0 | 0 | 0 | 10.848 | 11.004 |

### Run Mega Summary

![Run Mega Summary](../../.test-results/bench/visuals/summaries/summary-run-mega.png)

- [Open run mega summary image](../../.test-results/bench/visuals/summaries/summary-run-mega.png)

### Benchmark Graphs

![Performance Graph](../../.test-results/bench/find-on-screen-performance.svg)

- [Open performance graph](../../.test-results/bench/find-on-screen-performance.svg)

![Accuracy Graph](../../.test-results/bench/find-on-screen-accuracy.svg)

- [Open accuracy graph](../../.test-results/bench/find-on-screen-accuracy.svg)

### Resolution Group Graphs

![Resolution Match Time](../../.test-results/bench/find-on-screen-resolution-time.svg)

- [Open resolution match time graph](../../.test-results/bench/find-on-screen-resolution-time.svg)

![Resolution Matches](../../.test-results/bench/find-on-screen-resolution-matches.svg)

- [Open resolution matches graph](../../.test-results/bench/find-on-screen-resolution-matches.svg)

![Resolution Misses](../../.test-results/bench/find-on-screen-resolution-misses.svg)

- [Open resolution misses graph](../../.test-results/bench/find-on-screen-resolution-misses.svg)

![Resolution False Positives](../../.test-results/bench/find-on-screen-resolution-false-positives.svg)

- [Open resolution false positives graph](../../.test-results/bench/find-on-screen-resolution-false-positives.svg)

### Scenario Summary Images (15)

#### `glyph_large_r90_800x450`

![glyph_large_r90_800x450](../../.test-results/bench/visuals/summaries/summary-glyph_large_r90_800x450.png)

- [Open `glyph_large_r90_800x450` image](../../.test-results/bench/visuals/summaries/summary-glyph_large_r90_800x450.png)

#### `glyph_medium_r0_640x360`

![glyph_medium_r0_640x360](../../.test-results/bench/visuals/summaries/summary-glyph_medium_r0_640x360.png)

- [Open `glyph_medium_r0_640x360` image](../../.test-results/bench/visuals/summaries/summary-glyph_medium_r0_640x360.png)

#### `glyph_small_r270_480x270`

![glyph_small_r270_480x270](../../.test-results/bench/visuals/summaries/summary-glyph_small_r270_480x270.png)

- [Open `glyph_small_r270_480x270` image](../../.test-results/bench/visuals/summaries/summary-glyph_small_r270_480x270.png)

#### `grid_large_r180_800x450`

![grid_large_r180_800x450](../../.test-results/bench/visuals/summaries/summary-grid_large_r180_800x450.png)

- [Open `grid_large_r180_800x450` image](../../.test-results/bench/visuals/summaries/summary-grid_large_r180_800x450.png)

#### `grid_medium_r90_640x360`

![grid_medium_r90_640x360](../../.test-results/bench/visuals/summaries/summary-grid_medium_r90_640x360.png)

- [Open `grid_medium_r90_640x360` image](../../.test-results/bench/visuals/summaries/summary-grid_medium_r90_640x360.png)

#### `grid_small_r0_480x270`

![grid_small_r0_480x270](../../.test-results/bench/visuals/summaries/summary-grid_small_r0_480x270.png)

- [Open `grid_small_r0_480x270` image](../../.test-results/bench/visuals/summaries/summary-grid_small_r0_480x270.png)

- 9 additional scenario images are linked below.

### Scenario Image Links

- [`glyph_large_r90_800x450`](../../.test-results/bench/visuals/summaries/summary-glyph_large_r90_800x450.png)
- [`glyph_medium_r0_640x360`](../../.test-results/bench/visuals/summaries/summary-glyph_medium_r0_640x360.png)
- [`glyph_small_r270_480x270`](../../.test-results/bench/visuals/summaries/summary-glyph_small_r270_480x270.png)
- [`grid_large_r180_800x450`](../../.test-results/bench/visuals/summaries/summary-grid_large_r180_800x450.png)
- [`grid_medium_r90_640x360`](../../.test-results/bench/visuals/summaries/summary-grid_medium_r90_640x360.png)
- [`grid_small_r0_480x270`](../../.test-results/bench/visuals/summaries/summary-grid_small_r0_480x270.png)
- [`noise_large_r270_960x540`](../../.test-results/bench/visuals/summaries/summary-noise_large_r270_960x540.png)
- [`noise_medium_r180_800x450`](../../.test-results/bench/visuals/summaries/summary-noise_medium_r180_800x450.png)
- [`orbtex_large_r180_1024x576`](../../.test-results/bench/visuals/summaries/summary-orbtex_large_r180_1024x576.png)
- [`orbtex_large_r90_960x540`](../../.test-results/bench/visuals/summaries/summary-orbtex_large_r90_960x540.png)
- [`orbtex_medium_r0_800x450`](../../.test-results/bench/visuals/summaries/summary-orbtex_medium_r0_800x450.png)
- [`orbx_perspective_960x540`](../../.test-results/bench/visuals/summaries/summary-orbx_perspective_960x540.png)
- [`orbx_resize_115_960x540`](../../.test-results/bench/visuals/summaries/summary-orbx_resize_115_960x540.png)
- [`orbx_rotate_12deg_960x540`](../../.test-results/bench/visuals/summaries/summary-orbx_rotate_12deg_960x540.png)
- [`orbx_skewx_010_960x540`](../../.test-results/bench/visuals/summaries/summary-orbx_skewx_010_960x540.png)

### Attempt Image Links (sample 12 of 75)

- [engine-hybrid-attempt-1-ok.png](../../.test-results/bench/visuals/attempts/glyph_large_r90_800x450/engine-hybrid-attempt-1-ok.png)
- [engine-orb-attempt-1-not_found.png](../../.test-results/bench/visuals/attempts/glyph_large_r90_800x450/engine-orb-attempt-1-not_found.png)
- [engine-orb-attempt-2-not_found.png](../../.test-results/bench/visuals/attempts/glyph_large_r90_800x450/engine-orb-attempt-2-not_found.png)
- [engine-template-attempt-1-ok.png](../../.test-results/bench/visuals/attempts/glyph_large_r90_800x450/engine-template-attempt-1-ok.png)
- [engine-hybrid-attempt-1-ok.png](../../.test-results/bench/visuals/attempts/glyph_medium_r0_640x360/engine-hybrid-attempt-1-ok.png)
- [engine-orb-attempt-1-not_found.png](../../.test-results/bench/visuals/attempts/glyph_medium_r0_640x360/engine-orb-attempt-1-not_found.png)
- [engine-orb-attempt-2-not_found.png](../../.test-results/bench/visuals/attempts/glyph_medium_r0_640x360/engine-orb-attempt-2-not_found.png)
- [engine-template-attempt-1-ok.png](../../.test-results/bench/visuals/attempts/glyph_medium_r0_640x360/engine-template-attempt-1-ok.png)
- [engine-hybrid-attempt-1-ok.png](../../.test-results/bench/visuals/attempts/glyph_small_r270_480x270/engine-hybrid-attempt-1-ok.png)
- [engine-orb-attempt-1-not_found.png](../../.test-results/bench/visuals/attempts/glyph_small_r270_480x270/engine-orb-attempt-1-not_found.png)
- [engine-orb-attempt-2-not_found.png](../../.test-results/bench/visuals/attempts/glyph_small_r270_480x270/engine-orb-attempt-2-not_found.png)
- [engine-template-attempt-1-ok.png](../../.test-results/bench/visuals/attempts/glyph_small_r270_480x270/engine-template-attempt-1-ok.png)
<!-- END: FIND_ON_SCREEN_BENCH_AUTOGEN -->
