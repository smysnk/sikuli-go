#!/usr/bin/env python3
"""Render benchmark documentation pages from benchmark JSON artifacts."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


def liquid_url(path: str) -> str:
    normalized = "/" + path.strip("/")
    return f"{{{{ '{normalized}' | relative_url }}}}"


def md_link(label: str, path: str) -> str:
    return f"[{label}]({liquid_url(path)})"


def md_image(alt: str, path: str) -> str:
    return f"![{alt}]({liquid_url(path)})"


def front_matter(title: str, lead: str, kicker: str = "Benchmarks") -> list[str]:
    return [
        "---",
        "layout: guide",
        f"title: {title}",
        "nav_key: benchmarks",
        f"kicker: {kicker}",
        f"lead: {lead}",
        "---",
        "",
    ]


def fmt_float(value: Any, digits: int = 1) -> str:
    try:
        return f"{float(value):.{digits}f}"
    except Exception:
        return "0.0"


def fmt_int(value: Any) -> str:
    try:
        return str(int(round(float(value))))
    except Exception:
        return "0"


def _locate_marker(parts: tuple[str, ...], marker: tuple[str, ...]) -> int | None:
    limit = len(parts) - len(marker) + 1
    for index in range(max(limit, 0)):
        if parts[index : index + len(marker)] == marker:
            return index
    return None


def resolve_repo_path(root: Path, report_dir: Path, published_report_dir: Path, path: str | Path | None) -> Path | None:
    if not path:
        return None

    raw = Path(path)
    candidates: list[Path] = []
    if raw.is_absolute():
        candidates.append(raw)
        parts = raw.parts
        docs_index = _locate_marker(parts, ("docs",))
        if docs_index is not None:
            candidates.append(root / Path(*parts[docs_index:]))
        bench_index = _locate_marker(parts, (".test-results", "bench"))
        if bench_index is not None:
            suffix = Path(*parts[bench_index + 2 :])
            candidates.append(published_report_dir / suffix)
            candidates.append(report_dir / suffix)
    else:
        candidates.extend(
            [
                root / raw,
                report_dir / raw,
                published_report_dir / raw,
            ]
        )

    seen: set[Path] = set()
    for candidate in candidates:
        try:
            resolved = candidate.resolve()
        except FileNotFoundError:
            resolved = candidate
        if resolved in seen:
            continue
        seen.add(resolved)
        if resolved.exists():
            return resolved

    return candidates[0] if candidates else None


def display_path(root: Path, report_dir: Path, published_report_dir: Path, path: str | Path | None) -> str:
    candidate = resolve_repo_path(root, report_dir, published_report_dir, path)
    if candidate is None:
        return "(unset)"
    try:
        return candidate.relative_to(root).as_posix()
    except Exception:
        return str(path)


def docs_site_path(root: Path, report_dir: Path, published_report_dir: Path, path: str | Path | None) -> str | None:
    candidate = resolve_repo_path(root, report_dir, published_report_dir, path)
    if candidate is None:
        return None
    docs_root = (root / "docs").resolve()
    try:
        return candidate.resolve().relative_to(docs_root).as_posix()
    except Exception:
        return None


def guide_card(path: str, eyebrow: str, title: str, body: str, subtle: bool = False) -> list[str]:
    klass = "guide-card guide-card--subtle" if subtle else "guide-card"
    return [
        f'<a class="{klass}" href="{liquid_url(path)}">',
        f'  <span class="guide-card__eyebrow">{eyebrow}</span>',
        f'  <span class="guide-card__title">{title}</span>',
        f'  <span class="guide-card__body">{body}</span>',
        "</a>",
    ]


def guide_grid(cards: list[list[str]]) -> list[str]:
    lines = ["<div class=\"guide-grid\">"]
    for card in cards:
        lines.extend(card)
    lines.extend(["</div>", ""])
    return lines


def guide_meta(items: list[tuple[str, str]]) -> list[str]:
    lines = ["<div class=\"guide-meta\">"]
    for label, value in items:
        lines.append('  <div class="guide-meta__item">')
        lines.append(f'    <span class="guide-meta__label">{label}</span>')
        lines.append(f"    {value}")
        lines.append("  </div>")
    lines.extend(["</div>", ""])
    return lines


def guide_callout(title: str, body: str) -> list[str]:
    return [
        '<div class="guide-callout">',
        f"  <strong>{title}</strong>",
        f"  <span>{body}</span>",
        "</div>",
        "",
    ]


def write_table(headers: list[str], rows: list[list[str]]) -> list[str]:
    if not rows:
        return ["_No data available._", ""]
    lines = [
        "| " + " | ".join(headers) + " |",
        "| " + " | ".join(["---"] * len(headers)) + " |",
    ]
    for row in rows:
        lines.append("| " + " | ".join(row) + " |")
    lines.append("")
    return lines


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def pick_accuracy_winner(metrics_rows: list[dict[str, Any]]) -> dict[str, Any]:
    return max(metrics_rows, key=lambda row: (float(row.get("success_rate_pct", 0.0)), -float(row.get("avg_ms_per_op", 0.0))))


def pick_fastest(metrics_rows: list[dict[str, Any]]) -> dict[str, Any]:
    return min(metrics_rows, key=lambda row: float(row.get("avg_ms_per_op", 0.0)))


def pick_lowest_false_positive(metrics_rows: list[dict[str, Any]]) -> dict[str, Any]:
    return min(metrics_rows, key=lambda row: (float(row.get("false_positive_rate_pct", 0.0)), float(row.get("avg_ms_per_op", 0.0))))


def artifact_rows(published_report_dir_rel: str, strategy_available: bool) -> list[list[str]]:
    rows = [
        ["Overview", "Current benchmark summary at the section root.", md_link("Open", "bench/")],
        ["Reports Hub", "Artifact map for the current benchmark run.", md_link("Open", f"{published_report_dir_rel}/")],
        ["Detailed E2E", "Full engine, resolution, and scenario breakdown.", md_link("Open", f"{published_report_dir_rel}/find-on-screen-e2e")],
        ["Benchmark JSON", "Machine-readable benchmark summary.", md_link("Open", f"{published_report_dir_rel}/find-on-screen-e2e.json")],
        ["Benchmark Text", "Raw `go test` benchmark output.", md_link("Open", f"{published_report_dir_rel}/find-on-screen-e2e.txt")],
    ]
    if strategy_available:
        rows.append(["Scenario Strategy", "Scenario corpus and engine-selection rationale.", md_link("Open", f"{published_report_dir_rel}/find-on-screen-scenario-strategy")])
        rows.append(["Strategy JSON", "Machine-readable strategy summary.", md_link("Open", f"{published_report_dir_rel}/find-on-screen-scenario-strategy.json")])
    rows.extend(
        [
            ["Visual Gallery", "Generated benchmark screenshots and summaries.", md_link("Open", f"{published_report_dir_rel}/visuals/")],
            ["Scenario Intent", "What each scenario is intended to prove.", md_link("Open", "bench/FIND_ON_SCREEN_SCENARIO_INTENT")],
            ["Scenario Schema", "Manifest schema and region workflow.", md_link("Open", "bench/FIND_ON_SCREEN_SCENARIO_SCHEMA")],
        ]
    )
    return rows


def build_placeholder_overview(overview_path: Path) -> None:
    lines: list[str] = []
    lines.extend(
        front_matter(
            title="Benchmarks",
            lead="Benchmark pages are generated from the current benchmark artifacts and were not available for this build.",
            kicker="Secondary",
        )
    )
    lines.extend(
        guide_callout(
            "No Benchmark Artifacts Found",
            "Run `make benchmark` to generate the benchmark report, visuals, and `/bench/` guide pages before previewing or publishing this section.",
        )
    )
    lines.append("## Related Docs")
    lines.append("")
    lines.extend(
        write_table(
            ["Document", "Purpose", "Link"],
            [
                ["Build From Source", "Benchmark command reference and artifact list.", md_link("Open", "guides/build-from-source")],
                ["Scenario Intent", "Why each scenario exists and what it should prove.", md_link("Open", "bench/FIND_ON_SCREEN_SCENARIO_INTENT")],
                ["Scenario Schema", "Manifest structure and region-selection workflow.", md_link("Open", "bench/FIND_ON_SCREEN_SCENARIO_SCHEMA")],
            ],
        )
    )
    overview_path.parent.mkdir(parents=True, exist_ok=True)
    overview_path.write_text("\n".join(lines), encoding="utf-8")


def build_overview_page(
    root: Path,
    report_dir: Path,
    published_report_dir: Path,
    overview_path: Path,
    published_report_dir_rel: str,
    report: dict[str, Any],
    strategy: dict[str, Any] | None,
) -> None:
    meta = report.get("metadata") or {}
    summary = report.get("summary") or {}
    metrics_rows = summary.get("metrics_chart") or []
    if not metrics_rows:
        return

    fastest = pick_fastest(metrics_rows)
    accurate = pick_accuracy_winner(metrics_rows)
    lowest_fp = pick_lowest_false_positive(metrics_rows)
    strategy_summary = (strategy or {}).get("summary") or {}
    strategy_available = strategy is not None

    lines: list[str] = []
    lines.extend(
        front_matter(
            title="Benchmarks",
            lead="Latest find-on-screen benchmark results, scenario strategy, and published artifacts rendered into the guide shell.",
            kicker="Secondary",
        )
    )
    lines.extend(
        guide_grid(
            [
                guide_card(f"{published_report_dir_rel}/find-on-screen-e2e", "Detailed Report", "Engine Breakdown", "Inspect engine latency, success, and resolution breakdowns."),
                guide_card(f"{published_report_dir_rel}/find-on-screen-scenario-strategy", "Scenario Strategy", "Corpus Design", "Review the scenario matrix, visual examples, and engine-selection rationale."),
                guide_card(f"{published_report_dir_rel}/visuals/", "Artifacts", "Visual Gallery", "Open the generated screenshot gallery and summary images."),
                guide_card("bench/FIND_ON_SCREEN_SCENARIO_INTENT", "Scenario Docs", "Scenario Intent", "See what each benchmark scenario is trying to prove.", subtle=True),
            ]
        )
    )

    lines.append("## Latest Run")
    lines.append("")
    lines.extend(
        guide_meta(
            [
                ("Generated", f"`{meta.get('timestamp_utc', '(unknown)')}`"),
                ("Platform", f"`{meta.get('goos', '?')}/{meta.get('goarch', '?')}` on `{meta.get('cpu', '(unknown)')}`"),
                ("Target", f"`{meta.get('bench_name', '(unknown)')}` with benchtime `{meta.get('benchtime', '(unknown)')}` and count `{meta.get('count', '(unknown)')}`"),
                (
                    "Scenario Corpus",
                    f"`{strategy_summary.get('scenario_type_count', 0)}` scenario types across `{strategy_summary.get('resolution_group_count', 0)}` resolution groups"
                    if strategy_available
                    else f"Manifest `{display_path(root, report_dir, published_report_dir, meta.get('manifest'))}`",
                ),
            ]
        )
    )
    lines.extend(
        guide_callout(
            "Key Findings",
            f"Most accurate: `{accurate.get('engine')}` at `{fmt_float(accurate.get('success_rate_pct'))}%` success. "
            f"Fastest: `{fastest.get('engine')}` at `{fmt_float(fastest.get('avg_ms_per_op'), 3)}` ms/op. "
            f"Lowest false-positive rate: `{lowest_fp.get('engine')}` at `{fmt_float(lowest_fp.get('false_positive_rate_pct'))}%`.",
        )
    )

    lines.append("## Engine Snapshot")
    lines.append("")
    snapshot_rows = [
        [
            f"`{row.get('engine', '')}`",
            fmt_int(row.get("cases", 0)),
            fmt_float(row.get("avg_ms_per_op", 0.0), 3),
            fmt_float(row.get("median_ms_per_op", 0.0), 3),
            fmt_float(row.get("success_rate_pct", 0.0)),
            fmt_float(row.get("false_positive_rate_pct", 0.0)),
            fmt_float(row.get("not_found_rate_pct", 0.0)),
        ]
        for row in metrics_rows
    ]
    lines.extend(write_table(["Engine", "Cases", "Avg ms/op", "Median ms/op", "Success %", "False Positive %", "No Match %"], snapshot_rows))

    lines.append("## Charts")
    lines.append("")
    for alt, path in [
        ("Performance chart", f"{published_report_dir_rel}/find-on-screen-performance.svg"),
        ("Accuracy chart", f"{published_report_dir_rel}/find-on-screen-accuracy.svg"),
        ("Resolution time chart", f"{published_report_dir_rel}/find-on-screen-resolution-time.svg"),
        ("Resolution matches chart", f"{published_report_dir_rel}/find-on-screen-resolution-matches.svg"),
    ]:
        lines.append(md_image(alt, path))
        lines.append("")

    lines.append("## Artifact Map")
    lines.append("")
    lines.extend(write_table(["Artifact", "Purpose", "Link"], artifact_rows(published_report_dir_rel, strategy_available)))

    lines.append("## Scenario Corpus")
    lines.append("")
    lines.extend(
        write_table(
            ["Document", "Purpose", "Link"],
            [
                ["Intent", "Why each scenario exists and what it should stress.", md_link("Open", "bench/FIND_ON_SCREEN_SCENARIO_INTENT")],
                ["Schema", "Manifest structure, region-selection flow, and validation inputs.", md_link("Open", "bench/FIND_ON_SCREEN_SCENARIO_SCHEMA")],
            ],
        )
    )
    overview_path.write_text("\n".join(lines), encoding="utf-8")


def build_reports_index_page(
    root: Path,
    report_dir: Path,
    published_report_dir: Path,
    reports_index_path: Path,
    published_report_dir_rel: str,
    report: dict[str, Any],
    strategy: dict[str, Any] | None,
) -> None:
    meta = report.get("metadata") or {}
    summary = report.get("summary") or {}
    metrics_rows = summary.get("metrics_chart") or []
    strategy_available = strategy is not None

    lines: list[str] = []
    lines.extend(front_matter("Benchmark Reports", "Current benchmark artifacts, detailed reports, and raw outputs for the latest published run."))
    lines.extend(
        guide_grid(
            [
                guide_card("bench/", "Overview", "Benchmarks", "Return to the benchmark overview at the section root."),
                guide_card(f"{published_report_dir_rel}/find-on-screen-e2e", "Detailed Report", "FindOnScreen E2E", "Open the full benchmark breakdown page."),
                guide_card(f"{published_report_dir_rel}/find-on-screen-scenario-strategy", "Strategy", "Scenario Strategy", "Review the scenario set and visual examples.", subtle=not strategy_available),
                guide_card(f"{published_report_dir_rel}/visuals/", "Artifacts", "Visual Gallery", "Open the generated image gallery and summaries."),
            ]
        )
    )

    lines.append("## Run Summary")
    lines.append("")
    lines.extend(
        guide_meta(
            [
                ("Generated", f"`{meta.get('timestamp_utc', '(unknown)')}`"),
                ("Target", f"`{meta.get('bench_name', '(unknown)')}`"),
                ("Engines", ", ".join(f"`{row.get('engine', '')}`" for row in metrics_rows) if metrics_rows else "(none)"),
                ("Manifest", f"`{display_path(root, report_dir, published_report_dir, meta.get('manifest'))}`"),
            ]
        )
    )

    lines.append("## Artifacts")
    lines.append("")
    lines.extend(write_table(["Artifact", "Purpose", "Link"], artifact_rows(published_report_dir_rel, strategy_available)))
    reports_index_path.write_text("\n".join(lines), encoding="utf-8")


def build_e2e_page(
    root: Path,
    report_dir: Path,
    published_report_dir: Path,
    e2e_md_path: Path,
    published_report_dir_rel: str,
    report: dict[str, Any],
) -> None:
    meta = report.get("metadata") or {}
    summary = report.get("summary") or {}
    by_engine = summary.get("by_engine") or []
    metrics_rows = summary.get("metrics_chart") or []
    by_resolution = summary.get("by_resolution") or []
    by_kind = summary.get("by_kind") or []
    by_resolution_kind = summary.get("by_resolution_kind") or []
    results = report.get("results") or []

    lines: list[str] = []
    lines.extend(front_matter("FindOnScreen Benchmark Report", "Full engine, resolution, scenario-kind, and artifact breakdown for the latest benchmark run."))
    lines.extend(
        guide_grid(
            [
                guide_card("bench/", "Overview", "Benchmark Overview", "Return to the section summary and latest charts."),
                guide_card(f"{published_report_dir_rel}/", "Reports Hub", "Artifact Index", "Browse the raw outputs and related benchmark pages."),
                guide_card(f"{published_report_dir_rel}/find-on-screen-scenario-strategy", "Strategy", "Scenario Strategy", "Inspect the scenario corpus and visual examples."),
                guide_card(f"{published_report_dir_rel}/visuals/", "Artifacts", "Visual Gallery", "Open generated images and summary boards.", subtle=True),
            ]
        )
    )

    lines.append("## Run Metadata")
    lines.append("")
    lines.extend(
        guide_meta(
            [
                ("Generated", f"`{meta.get('timestamp_utc', '(unknown)')}`"),
                ("Package", f"`{meta.get('package', '(unknown)')}`"),
                ("Target", f"`{meta.get('bench_name', '(unknown)')}`"),
                ("Benchtime", f"`{meta.get('benchtime', '(unknown)')}`"),
                ("Count", f"`{meta.get('count', '(unknown)')}`"),
                ("Tags", f"`{meta.get('tags') or '(none)'}`"),
                ("Platform", f"`{meta.get('goos', '?')}/{meta.get('goarch', '?')}` on `{meta.get('cpu', '(unknown)')}`"),
                ("Manifest", f"`{display_path(root, report_dir, published_report_dir, meta.get('manifest'))}`"),
            ]
        )
    )

    lines.append("## Engine Summary")
    lines.append("")
    engine_rows = [
        [
            f"`{row.get('engine', '')}`",
            fmt_int(row.get("cases", 0)),
            fmt_int(row.get("query_ok", 0)),
            fmt_int(row.get("query_partial", 0)),
            fmt_int(row.get("query_not_found", 0)),
            fmt_int(row.get("query_unsupported", 0)),
            fmt_int(row.get("query_error", 0)),
            fmt_int(row.get("query_overlap_miss", 0)),
            fmt_float(row.get("avg_ms_per_op", 0.0), 3),
            fmt_float(row.get("median_ms_per_op", 0.0), 3),
            f"`{row.get('best_scenario', '')}`",
            f"`{row.get('worst_scenario', '')}`",
        ]
        for row in by_engine
    ]
    lines.extend(
        write_table(
            ["Engine", "Cases", "OK", "Partial", "Not Found", "Unsupported", "Error", "Overlap Miss", "Avg ms/op", "Median ms/op", "Best Scenario", "Worst Scenario"],
            engine_rows,
        )
    )

    lines.append("## Summary Metrics")
    lines.append("")
    metric_rows = [
        [
            f"`{row.get('engine', '')}`",
            fmt_int(row.get("cases", 0)),
            fmt_int(row.get("rows", 0)),
            fmt_float(row.get("avg_ms_per_op", 0.0), 3),
            fmt_float(row.get("median_ms_per_op", 0.0), 3),
            fmt_float(row.get("success_rate_pct", 0.0)),
            fmt_float(row.get("false_positive_rate_pct", 0.0)),
            fmt_float(row.get("not_found_rate_pct", 0.0)),
            fmt_float(row.get("unsupported_rate_pct", 0.0)),
            fmt_float(row.get("error_rate_pct", 0.0)),
        ]
        for row in metrics_rows
    ]
    lines.extend(write_table(["Engine", "Cases", "Rows", "Avg ms/op", "Median ms/op", "Success %", "False Positive %", "No Match %", "Unsupported %", "Error %"], metric_rows))

    lines.append("## Charts")
    lines.append("")
    for alt, path in [
        ("Performance chart", f"{published_report_dir_rel}/find-on-screen-performance.svg"),
        ("Accuracy chart", f"{published_report_dir_rel}/find-on-screen-accuracy.svg"),
        ("Scenario kind time chart", f"{published_report_dir_rel}/find-on-screen-kind-time.svg"),
        ("Scenario kind success chart", f"{published_report_dir_rel}/find-on-screen-kind-success.svg"),
        ("Resolution time chart", f"{published_report_dir_rel}/find-on-screen-resolution-time.svg"),
        ("Resolution matches chart", f"{published_report_dir_rel}/find-on-screen-resolution-matches.svg"),
        ("Resolution misses chart", f"{published_report_dir_rel}/find-on-screen-resolution-misses.svg"),
        ("Resolution false positives chart", f"{published_report_dir_rel}/find-on-screen-resolution-false-positives.svg"),
    ]:
        lines.append(md_image(alt, path))
        lines.append("")

    lines.append("## Resolution Breakdown")
    lines.append("")
    resolution_rows: list[list[str]] = []
    for row in by_resolution:
        resolution = row.get("resolution", "")
        for engine, metrics in sorted((row.get("engines") or {}).items()):
            resolution_rows.append(
                [
                    f"`{resolution}`",
                    f"`{engine}`",
                    fmt_int(metrics.get("cases", 0)),
                    fmt_float(metrics.get("avg_ms_per_op", 0.0), 3),
                    fmt_int(metrics.get("match_count", 0)),
                    fmt_int(metrics.get("miss_count", 0)),
                    fmt_int(metrics.get("false_positive_count", 0)),
                ]
            )
    lines.extend(write_table(["Resolution", "Engine", "Cases", "Avg ms/op", "Matches", "Misses", "False Positives"], resolution_rows))

    lines.append("## Scenario Kind Breakdown")
    lines.append("")
    kind_rows: list[list[str]] = []
    for row in by_kind:
        kind = row.get("kind", "")
        for engine, metrics in sorted((row.get("engines") or {}).items()):
            kind_rows.append(
                [
                    f"`{kind}`",
                    f"`{engine}`",
                    fmt_int(metrics.get("cases", 0)),
                    fmt_float(metrics.get("avg_ms_per_op", 0.0), 3),
                    fmt_float(metrics.get("success_pct", 0.0)),
                    fmt_float(metrics.get("not_found_pct", 0.0)),
                    fmt_float(metrics.get("false_positive_pct", 0.0)),
                ]
            )
    lines.extend(write_table(["Kind", "Engine", "Cases", "Avg ms/op", "Success %", "Not Found %", "False Positive %"], kind_rows))

    lines.append("## Resolution + Kind Breakdown")
    lines.append("")
    rk_rows: list[list[str]] = []
    for row in by_resolution_kind:
        resolution = row.get("resolution", "")
        kind = row.get("kind", "")
        for engine, metrics in sorted((row.get("engines") or {}).items()):
            rk_rows.append(
                [
                    f"`{resolution}`",
                    f"`{kind}`",
                    f"`{engine}`",
                    fmt_int(metrics.get("cases", 0)),
                    fmt_float(metrics.get("avg_ms_per_op", 0.0), 3),
                    fmt_float(metrics.get("success_pct", 0.0)),
                    fmt_float(metrics.get("not_found_pct", 0.0)),
                    fmt_float(metrics.get("false_positive_pct", 0.0)),
                ]
            )
    lines.extend(write_table(["Resolution", "Kind", "Engine", "Cases", "Avg ms/op", "Success %", "Not Found %", "False Positive %"], rk_rows))

    lines.append("## Slowest Benchmark Rows")
    lines.append("")
    slow_rows = sorted(results, key=lambda row: float(row.get("ms_per_op", 0.0)), reverse=True)[:12]
    slow_table = [
        [
            f"`{row.get('scenario', '')}`",
            f"`{row.get('engine', '')}`",
            f"`{row.get('status', '')}`",
            fmt_int(row.get("iterations", 0)),
            fmt_float(row.get("ms_per_op", 0.0), 3),
            fmt_int(row.get("bytes_per_op", 0)),
            fmt_int(row.get("allocs_per_op", 0)),
        ]
        for row in slow_rows
    ]
    lines.extend(write_table(["Scenario", "Engine", "Status", "Iterations", "ms/op", "Bytes/op", "Allocs/op"], slow_table))

    lines.append("## Raw Artifacts")
    lines.append("")
    lines.extend(write_table(["Artifact", "Purpose", "Link"], artifact_rows(published_report_dir_rel, strategy_available=True)))
    e2e_md_path.write_text("\n".join(lines), encoding="utf-8")


def metric_dict_rows(title: str, payload: dict[str, Any]) -> list[str]:
    rows = [[f"`{key}`", f"`{payload[key]}`"] for key in sorted(payload.keys())]
    lines = [f"### {title}", ""]
    lines.extend(write_table(["Metric", "Value"], rows))
    return lines


def build_strategy_page(
    root: Path,
    report_dir: Path,
    published_report_dir: Path,
    strategy_md_path: Path,
    published_report_dir_rel: str,
    strategy: dict[str, Any],
) -> None:
    summary = strategy.get("summary") or {}
    scenarios = strategy.get("scenarios") or []
    visual_examples = strategy.get("visual_examples") or {}
    engine_cmp = summary.get("engine_match_comparison") or {}

    lines: list[str] = []
    lines.extend(front_matter("FindOnScreen Scenario Strategy", "Scenario-corpus design, engine comparison context, and generated visual examples for the current benchmark set."))
    lines.extend(
        guide_grid(
            [
                guide_card("bench/", "Overview", "Benchmark Overview", "Return to the benchmark section summary."),
                guide_card(f"{published_report_dir_rel}/find-on-screen-e2e", "Detailed Report", "E2E Results", "Compare strategy guidance against the latest benchmark outcome."),
                guide_card("bench/FIND_ON_SCREEN_SCENARIO_INTENT", "Scenario Docs", "Scenario Intent", "Review what each scenario is intended to prove."),
                guide_card("bench/FIND_ON_SCREEN_SCENARIO_SCHEMA", "Schema", "Scenario Schema", "Inspect manifest structure and region-selection workflow.", subtle=True),
            ]
        )
    )

    lines.append("## Strategy Metadata")
    lines.append("")
    lines.extend(
        guide_meta(
            [
                ("Generated", f"`{summary.get('generated_at_utc', '(unknown)')}`"),
                ("Manifest", f"`{display_path(root, report_dir, published_report_dir, summary.get('manifest_path'))}`"),
                ("Schema Version", f"`{summary.get('schema_version', '(unknown)')}`"),
                ("Engines", ", ".join(f"`{engine}`" for engine in summary.get("engines", [])) or "(none)"),
            ]
        )
    )

    lines.append("## Strategy Summary")
    lines.append("")
    if engine_cmp.get("available"):
        lines.extend(
            write_table(
                ["Metric", "Value"],
                [
                    ["Engine", f"`{engine_cmp.get('engine')}`"],
                    ["Engine Match Rate", f"`{fmt_float(engine_cmp.get('engine_match_rate_pct', 0.0))}%`"],
                    ["Other Engines Match Rate", f"`{fmt_float(engine_cmp.get('others_match_rate_pct', 0.0))}%`"],
                    ["Delta vs Others", f"`{fmt_float(engine_cmp.get('delta_pct_points', 0.0))} pts`"],
                    ["Engine Rank", f"`{fmt_int(engine_cmp.get('engine_rank', 0))}/{fmt_int(engine_cmp.get('engine_count', 0))}`"],
                    ["Benchmark Source", f"`{display_path(root, report_dir, published_report_dir, engine_cmp.get('bench_report_path'))}`"],
                ],
            )
        )
    else:
        lines.extend(guide_callout("Match-Rate Comparison Unavailable", engine_cmp.get("error", "No benchmark summary was available for comparison.")))

    lines.append("## Visual Examples")
    lines.append("")
    lines.extend(
        guide_meta(
            [
                ("Resolution", f"`{visual_examples.get('resolution', '(unknown)')}`"),
                ("Engine", f"`{visual_examples.get('engine', '(unknown)')}`"),
                ("Log", f"`{display_path(root, report_dir, published_report_dir, visual_examples.get('log_path'))}`"),
                ("Command", f"`{visual_examples.get('command', '(unknown)')}`"),
            ]
        )
    )
    visual_rows: list[list[str]] = []
    for item in visual_examples.get("summaries") or []:
        docs_path = docs_site_path(root, report_dir, published_report_dir, item.get("path"))
        if docs_path is None:
            continue
        scenario_id = item.get("scenario_id") or item.get("scenario_name") or "scenario"
        visual_rows.append([f"`{scenario_id}`", md_image(str(scenario_id), docs_path)])
    lines.extend(write_table(["Scenario", "Example"], visual_rows))

    lines.append("## Diversity Summary")
    lines.append("")
    lines.extend(
        write_table(
            ["Metric", "Value"],
            [
                ["Scenario Types", f"`{fmt_int(summary.get('enabled_scenario_type_count', 0))}/{fmt_int(summary.get('scenario_type_count', 0))}`"],
                ["Resolution Groups", f"`{fmt_int(summary.get('resolution_group_count', 0))}`"],
                ["Scenarios Per Resolution", f"`{fmt_int(summary.get('scenarios_per_resolution', 0))}`"],
                ["Expected Positive", f"`{fmt_int((summary.get('distribution') or {}).get('expected_positive', 0))}`"],
                ["Expected Negative", f"`{fmt_int((summary.get('distribution') or {}).get('expected_negative', 0))}`"],
            ],
        )
    )
    distribution = summary.get("distribution") or {}
    lines.extend(metric_dict_rows("Kinds", distribution.get("kinds") or {}))
    lines.extend(metric_dict_rows("Styles", distribution.get("styles") or {}))
    lines.extend(metric_dict_rows("Target Sources", distribution.get("target_sources") or {}))
    lines.extend(metric_dict_rows("Decoy Placements", distribution.get("decoy_placements") or {}))
    lines.extend(metric_dict_rows("Noise Types", distribution.get("noise_types") or {}))
    lines.extend(metric_dict_rows("Transform Coverage", summary.get("transform_coverage") or {}))

    lines.append("## Scenario Intent")
    lines.append("")
    intent_rows = [
        [f"`{scenario.get('id', '')}`", f"`{scenario.get('kind', '')}`", f"`{scenario.get('style', '')}`", str(scenario.get("goal", ""))]
        for scenario in scenarios
    ]
    lines.extend(write_table(["Scenario ID", "Kind", "Style", "Looking For"], intent_rows))

    lines.append("## Scenario Configuration Details")
    lines.append("")
    for scenario in scenarios:
        target = scenario.get("target") or {}
        background = scenario.get("background") or {}
        transforms = scenario.get("transforms") or {}
        photometric = scenario.get("photometric") or {}
        decoys = scenario.get("decoys") or {}
        occlusion = scenario.get("occlusion") or {}
        expected = scenario.get("expected") or {}
        hybrid_policy = scenario.get("hybrid_policy")
        monitor_selector = scenario.get("monitor_selector")
        noise_types = ((photometric.get("noise") or {}).get("types") or [])

        lines.append(f"### `{scenario.get('id', '')}`")
        lines.append("")
        lines.append(f"- Kind: `{scenario.get('kind', '')}`")
        lines.append(f"- Style: `{scenario.get('style', '')}`")
        lines.append(f"- Target: source=`{target.get('source', '')}` size=`{target.get('size_px', '')}` rotation=`{target.get('rotation_degrees', '')}` assets=`{', '.join(target.get('asset_pool') or []) or 'none'}`")
        lines.append(f"- Background: palette=`{background.get('palette', '')}` clutter=`{background.get('clutter_density', '')}` continuous_canvas=`{background.get('continuous_canvas', '')}`")
        lines.append(f"- Transforms: scale=`{transforms.get('scale', '')}` rotate=`{transforms.get('rotate', '')}` perspective_enabled=`{transforms.get('perspective_enabled', '')}` skew_x=`{transforms.get('skew_x', '')}` skew_y=`{transforms.get('skew_y', '')}`")
        lines.append(f"- Photometric: brightness=`{photometric.get('brightness', '')}` contrast=`{photometric.get('contrast', '')}` gamma=`{photometric.get('gamma', '')}` blur=`{photometric.get('blur_sigma', '')}` jpeg_quality=`{photometric.get('jpeg_quality', '')}` noise_types=`{', '.join(noise_types) or 'none'}`")
        lines.append(f"- Decoys: enabled=`{decoys.get('enabled', '')}` count=`{decoys.get('count', '')}` similarity=`{decoys.get('similarity', '')}` placement=`{decoys.get('placement', '')}`")
        lines.append(f"- Occlusion: enabled=`{occlusion.get('enabled', '')}` coverage=`{occlusion.get('target_coverage_pct', '')}`")
        lines.append(f"- Expected: positive=`{expected.get('positive', '')}` iou_min=`{expected.get('iou_min', '')}` area_ratio_max=`{expected.get('area_ratio_max', '')}` allow_partial=`{expected.get('allow_partial', '')}`")
        if monitor_selector:
            lines.append(f"- Monitor Selector: `{monitor_selector}`")
        if hybrid_policy:
            lines.append(f"- Hybrid Policy: `{hybrid_policy}`")
        lines.append("")

    lines.append("## Raw Artifacts")
    lines.append("")
    lines.extend(
        write_table(
            ["Artifact", "Purpose", "Link"],
            [
                ["Benchmark Overview", "Latest benchmark section root.", md_link("Open", "bench/")],
                ["Detailed E2E Report", "Compare strategy against measured results.", md_link("Open", f"{published_report_dir_rel}/find-on-screen-e2e")],
                ["Strategy JSON", "Machine-readable strategy summary.", md_link("Open", f"{published_report_dir_rel}/find-on-screen-scenario-strategy.json")],
                ["Visual Gallery", "Generated screenshots and summary boards.", md_link("Open", f"{published_report_dir_rel}/visuals/")],
            ],
        )
    )
    strategy_md_path.write_text("\n".join(lines), encoding="utf-8")


def build_visuals_index_page(
    root: Path,
    report_dir: Path,
    published_report_dir: Path,
    visuals_index_path: Path,
    published_report_dir_rel: str,
    strategy: dict[str, Any] | None,
) -> None:
    visual_examples = (strategy or {}).get("visual_examples") or {}
    summaries = visual_examples.get("summaries") or []
    summaries_dir = report_dir / "visuals" / "summaries"
    if not summaries and summaries_dir.exists():
        for image in sorted(summaries_dir.glob("summary-*.png"))[:10]:
            summaries.append({"scenario_id": image.stem, "path": str(image)})

    mega_summary = None
    for candidate in [
        report_dir / "visuals" / "summaries" / "summary-run-mega.jpg",
        report_dir / "visuals" / "summaries" / "summary-run-mega.png",
    ]:
        if candidate.exists():
            mega_summary = candidate
            break

    lines: list[str] = []
    lines.extend(front_matter("Benchmark Visual Gallery", "Generated run-level and per-scenario benchmark visuals rendered inside the guide shell."))
    lines.extend(
        guide_grid(
            [
                guide_card("bench/", "Overview", "Benchmark Overview", "Return to the benchmark summary."),
                guide_card(f"{published_report_dir_rel}/find-on-screen-e2e", "Detailed Report", "E2E Report", "Inspect the full metric breakdown."),
                guide_card(f"{published_report_dir_rel}/find-on-screen-scenario-strategy", "Strategy", "Scenario Strategy", "Compare the visuals to the scenario design."),
                guide_card(f"{published_report_dir_rel}/find-on-screen-e2e.json", "Raw Data", "Benchmark JSON", "Open the machine-readable benchmark summary.", subtle=True),
            ]
        )
    )

    lines.append("## Gallery Metadata")
    lines.append("")
    lines.extend(
        guide_meta(
            [
                ("Directory", f"`{display_path(root, report_dir, published_report_dir, report_dir / 'visuals')}`"),
                ("Resolution", f"`{visual_examples.get('resolution', '(mixed)')}`"),
                ("Engine", f"`{visual_examples.get('engine', '(mixed)')}`"),
                ("Log", f"`{display_path(root, report_dir, published_report_dir, visual_examples.get('log_path'))}`"),
            ]
        )
    )

    if mega_summary is not None:
        mega_path = docs_site_path(root, report_dir, published_report_dir, mega_summary)
        if mega_path is not None:
            lines.append("## Run-Level Summary")
            lines.append("")
            lines.append(md_image("Run-level benchmark summary", mega_path))
            lines.append("")

    lines.append("## Scenario Summaries")
    lines.append("")
    visual_rows: list[list[str]] = []
    for item in summaries:
        docs_path = docs_site_path(root, report_dir, published_report_dir, item.get("path"))
        if docs_path is None:
            continue
        scenario_id = item.get("scenario_id") or item.get("scenario_name") or Path(str(item.get("path"))).stem
        visual_rows.append([f"`{scenario_id}`", md_image(str(scenario_id), docs_path)])
    lines.extend(write_table(["Scenario", "Example"], visual_rows))
    visuals_index_path.parent.mkdir(parents=True, exist_ok=True)
    visuals_index_path.write_text("\n".join(lines), encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--project-root", required=True)
    parser.add_argument("--report-dir", required=True)
    parser.add_argument("--published-report-dir", required=True)
    parser.add_argument("--overview-output", required=True)
    args = parser.parse_args()

    root = Path(args.project_root).resolve()
    report_dir = Path(args.report_dir).resolve()
    published_report_dir = Path(args.published_report_dir).resolve()
    overview_path = Path(args.overview_output).resolve()

    if not report_dir.exists():
        build_placeholder_overview(overview_path)
        print(f"[bench-docs] report directory not found, wrote placeholder overview: {overview_path}")
        return 0

    e2e_json_path = report_dir / "find-on-screen-e2e.json"
    strategy_json_path = report_dir / "find-on-screen-scenario-strategy.json"
    if not e2e_json_path.is_file():
        build_placeholder_overview(overview_path)
        print(f"[bench-docs] benchmark json not found, wrote placeholder overview: {overview_path}")
        return 0

    report = load_json(e2e_json_path)
    strategy = load_json(strategy_json_path) if strategy_json_path.is_file() else None
    docs_root = (root / "docs").resolve()
    try:
        published_report_dir_rel = published_report_dir.resolve().relative_to(docs_root).as_posix()
    except Exception:
        published_report_dir_rel = display_path(root, report_dir, published_report_dir, published_report_dir)

    overview_path.parent.mkdir(parents=True, exist_ok=True)
    build_overview_page(root, report_dir, published_report_dir, overview_path, published_report_dir_rel, report, strategy)
    build_reports_index_page(root, report_dir, published_report_dir, report_dir / "index.md", published_report_dir_rel, report, strategy)
    build_e2e_page(root, report_dir, published_report_dir, report_dir / "find-on-screen-e2e.md", published_report_dir_rel, report)
    if strategy is not None:
        build_strategy_page(root, report_dir, published_report_dir, report_dir / "find-on-screen-scenario-strategy.md", published_report_dir_rel, strategy)
    build_visuals_index_page(root, report_dir, published_report_dir, report_dir / "visuals" / "index.md", published_report_dir_rel, strategy)

    print(f"[bench-docs] wrote overview: {overview_path}")
    print(f"[bench-docs] wrote reports index: {report_dir / 'index.md'}")
    print(f"[bench-docs] wrote report: {report_dir / 'find-on-screen-e2e.md'}")
    if strategy is not None:
        print(f"[bench-docs] wrote strategy: {report_dir / 'find-on-screen-scenario-strategy.md'}")
    print(f"[bench-docs] wrote visuals index: {report_dir / 'visuals' / 'index.md'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
