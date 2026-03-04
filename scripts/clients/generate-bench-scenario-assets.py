#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import math
import random
import urllib.request
from pathlib import Path

import numpy as np
from PIL import Image, ImageChops, ImageDraw, ImageEnhance, ImageFilter, ImageOps

ROOT = Path(__file__).resolve().parents[2]
ASSET_ROOT = ROOT / "packages" / "api" / "internal" / "grpcv1" / "testdata" / "find-bench-assets"
SCENARIO_ROOT = ASSET_ROOT / "scenario"
SOURCE_DIR = SCENARIO_ROOT / "source"
BENCH_DIR = SCENARIO_ROOT / "benchmark"
PREVIEW_DIR = SCENARIO_ROOT / "preview"
CACHE_DIR = ASSET_ROOT / "source-cache"
REGIONS_PATH = ASSET_ROOT / "regions.json"
SIKULIX_BG_PATH = ROOT / "docs" / "bench" / "assets" / "sikulix" / "sikulix-script.png"
PHOTO_CLUTTER_PATH = ROOT / "docs" / "bench" / "assets" / "photo" / "4256_clutter_crop_zoom.jpg"

SIKULIX_BG_URL = "https://sikulix.github.io/img/sikulix-script.png"

BASE_IDS = {
    "noise_stress_random": 290,
    "orb_feature_rich": 160,
    "perspective_skew_sweep": 260,
    "scale_rotate_sweep": 170,
    "hybrid_gate_conflicts": 130,
    "template_control_exact": 140,
}

SCENARIO_IDS = [
    "hybrid_gate_conflicts",
    "multi_monitor_dpi_shift",
    "noise_stress_random",
    "orb_feature_rich",
    "perspective_skew_sweep",
    "photo_clutter",
    "repetitive_grid_camouflage",
    "scale_rotate_sweep",
    "template_control_exact",
    "vector_ui_baseline",
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate benchmark scenario assets, benchmark variants, and regions")
    parser.add_argument("--width", type=int, default=2560)
    parser.add_argument("--height", type=int, default=1440)
    parser.add_argument("--offline", action="store_true", help="Do not download missing source images")
    parser.add_argument("--skip-regions", action="store_true", help="Generate images only; do not write regions.json")
    return parser.parse_args()


def ensure_dirs() -> None:
    for path in (SOURCE_DIR, BENCH_DIR, PREVIEW_DIR, CACHE_DIR, SIKULIX_BG_PATH.parent):
        path.mkdir(parents=True, exist_ok=True)


def ensure_sikulix_background(offline: bool) -> Image.Image:
    if not SIKULIX_BG_PATH.exists():
        if offline:
            raise FileNotFoundError(f"missing SikuliX screenshot in offline mode: {SIKULIX_BG_PATH}")
        with urllib.request.urlopen(SIKULIX_BG_URL, timeout=30) as response:
            SIKULIX_BG_PATH.write_bytes(response.read())
    return Image.open(SIKULIX_BG_PATH).convert("RGB")


def fetch_picsum(pic_id: int, width: int, height: int, offline: bool) -> Image.Image:
    cache_path = CACHE_DIR / f"picsum_{pic_id}_{width}x{height}.jpg"
    if not cache_path.exists():
        if offline:
            raise FileNotFoundError(f"missing cached image in offline mode: {cache_path}")
        url = f"https://picsum.photos/id/{pic_id}/{width}/{height}"
        with urllib.request.urlopen(url, timeout=30) as response:
            cache_path.write_bytes(response.read())
    return Image.open(cache_path).convert("RGB")


def fit(img: Image.Image, width: int, height: int) -> Image.Image:
    return ImageOps.fit(img.convert("RGB"), (width, height), method=Image.Resampling.LANCZOS)


def add_noise(img: Image.Image, amount: float, seed: int) -> Image.Image:
    rng = np.random.default_rng(seed)
    arr = np.asarray(img).astype(np.float32)
    noise = rng.normal(0.0, amount * 255.0, size=arr.shape).astype(np.float32)
    out = np.clip(arr + noise, 0, 255).astype(np.uint8)
    return Image.fromarray(out, "RGB")


def jpeg_roundtrip(img: Image.Image, quality: int) -> Image.Image:
    import io

    buf = io.BytesIO()
    img.save(buf, format="JPEG", quality=max(25, min(quality, 100)), optimize=True)
    buf.seek(0)
    return Image.open(buf).convert("RGB")


def smooth_background(width: int, height: int, c1: tuple[int, int, int], c2: tuple[int, int, int]) -> Image.Image:
    canvas = Image.new("RGB", (width, height), c1)
    px = canvas.load()
    for y in range(height):
        t = y / max(1, height - 1)
        r = int(c1[0] * (1 - t) + c2[0] * t)
        g = int(c1[1] * (1 - t) + c2[1] * t)
        b = int(c1[2] * (1 - t) + c2[2] * t)
        for x in range(width):
            px[x, y] = (r, g, b)
    return canvas


def pick_regions(img: Image.Image, count: int = 3) -> list[tuple[int, int, int, int]]:
    gray = np.asarray(ImageOps.grayscale(img), dtype=np.float32)
    h, w = gray.shape

    # keep targets >= 52px in smallest benchmark resolution (480x270)
    floor_scale = max(480.0 / float(w), 270.0 / float(h))
    min_win = max(60, int(math.ceil(52.0 / floor_scale)))

    win_w = max(min_win, int(w * 0.18))
    win_h = max(min_win, int(h * 0.18))
    win_w = min(win_w, int(w * 0.30))
    win_h = min(win_h, int(h * 0.30))

    margin_x = max(32, int(w * 0.05))
    margin_y = max(24, int(h * 0.05))
    step_x = max(24, win_w // 4)
    step_y = max(20, win_h // 4)

    gx = np.abs(np.diff(gray, axis=1))
    gy = np.abs(np.diff(gray, axis=0))
    grad = np.zeros_like(gray)
    grad[:, 1:] += gx
    grad[1:, :] += gy

    candidates: list[tuple[float, int, int, int, int]] = []
    for y in range(margin_y, max(margin_y + 1, h - win_h - margin_y), step_y):
        for x in range(margin_x, max(margin_x + 1, w - win_w - margin_x), step_x):
            patch = gray[y : y + win_h, x : x + win_w]
            gpatch = grad[y : y + win_h, x : x + win_w]
            score = float(patch.std()) * 0.42 + float(gpatch.mean()) * 1.35
            cx = x + win_w / 2.0
            cy = y + win_h / 2.0
            center_penalty = abs(cx - w / 2.0) / (w / 2.0) * 0.16 + abs(cy - h / 2.0) / (h / 2.0) * 0.16
            score *= max(0.55, 1.0 - center_penalty)
            candidates.append((score, x, y, win_w, win_h))

    candidates.sort(key=lambda item: item[0], reverse=True)

    def iou(a: tuple[int, int, int, int], b: tuple[int, int, int, int]) -> float:
        ax, ay, aw, ah = a
        bx, by, bw, bh = b
        x1 = max(ax, bx)
        y1 = max(ay, by)
        x2 = min(ax + aw, bx + bw)
        y2 = min(ay + ah, by + bh)
        if x2 <= x1 or y2 <= y1:
            return 0.0
        inter = (x2 - x1) * (y2 - y1)
        union = aw * ah + bw * bh - inter
        return inter / union if union else 0.0

    chosen: list[tuple[int, int, int, int]] = []
    for _, x, y, rw, rh in candidates:
        rect = (x, y, rw, rh)
        if any(iou(rect, prev) > 0.18 for prev in chosen):
            continue
        chosen.append(rect)
        if len(chosen) >= count:
            break

    fallback = [
        (w // 2 - win_w // 2, h // 2 - win_h // 2, win_w, win_h),
        (w // 3 - win_w // 2, h // 3 - win_h // 2, win_w, win_h),
        (2 * w // 3 - win_w // 2, 2 * h // 3 - win_h // 2, win_w, win_h),
    ]
    for rect in fallback:
        if len(chosen) >= count:
            break
        x, y, rw, rh = rect
        x = max(margin_x, min(w - rw - margin_x, x))
        y = max(margin_y, min(h - rh - margin_y, y))
        chosen.append((x, y, rw, rh))

    return chosen[:count]


def save_image(path: Path, image: Image.Image) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    ext = path.suffix.lower()
    if ext in {".jpg", ".jpeg"}:
        image.save(path, quality=92, optimize=True)
    else:
        image.save(path, optimize=True)


def rel(path: Path) -> str:
    return path.relative_to(ROOT).as_posix()


def variant_tile(size: int, variant: str = "base") -> Image.Image:
    tile = Image.new("RGB", (size, size), (208, 212, 220))
    d = ImageDraw.Draw(tile)
    for i in range(0, size, 8):
        tone = 170 + (i % 28)
        d.line((0, i, size, i), fill=(tone, tone, tone), width=1)
    d.rounded_rectangle((14, 14, size - 14, size - 14), radius=12, outline=(92, 98, 108), width=2)
    d.rectangle((34, 36, size - 34, 56), fill=(56, 62, 76))
    d.rectangle((34, size - 58, size - 34, size - 34), fill=(245, 245, 245))
    d.rectangle((46, 76, 66, size - 72), fill=(42, 48, 64))

    if variant == "target-a":
        d.rectangle((size - 84, 76, size - 44, 90), fill=(236, 239, 245))
        d.rectangle((size - 84, 96, size - 52, 108), fill=(212, 218, 230))
        d.polygon([(size - 54, 120), (size - 40, 128), (size - 54, 136)], fill=(24, 30, 44))
        d.rounded_rectangle((74, 84, 142, 144), radius=8, outline=(84, 92, 108), width=2)
    elif variant == "target-b":
        d.rectangle((80, 80, 96, 96), outline=(24, 30, 44), width=2)
        d.rectangle((104, 80, size - 44, 92), fill=(92, 100, 116))
        d.rectangle((80, 106, 96, 122), outline=(24, 30, 44), width=2)
        d.rectangle((104, 108, size - 54, 120), fill=(120, 128, 144))
        d.rectangle((80, 132, 96, 148), outline=(24, 30, 44), width=2)
        d.rectangle((104, 134, size - 64, 146), fill=(146, 156, 174))
    elif variant == "target-c":
        d.rounded_rectangle((74, 82, size - 44, 116), radius=10, fill=(228, 232, 240), outline=(96, 106, 124), width=2)
        d.rectangle((82, 90, size - 52, 98), fill=(120, 130, 146))
        d.ellipse((86, 122, 108, 144), fill=(70, 82, 104))
        d.rectangle((120, 126, size - 48, 140), fill=(168, 178, 196))
    return tile


def build_vector_ui_baseline(width: int, height: int, sikulix_bg: Image.Image) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]]]:
    source = fit(sikulix_bg, width, height)
    benchmark = source.copy()
    previews = [("original", source)]
    return source, benchmark, previews


def build_template_control_exact(width: int, height: int) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]]]:
    source = smooth_background(width, height, (24, 34, 48), (14, 22, 34))
    d = ImageDraw.Draw(source)
    d.rounded_rectangle((260, 120, width - 260, height - 160), radius=16, fill=(34, 46, 64), outline=(94, 114, 142), width=2)
    d.rectangle((260, 120, width - 260, 190), fill=(52, 69, 94))

    for i in range(6):
        y = 260 + i * 160
        d.rounded_rectangle((350, y, width - 350, y + 108), radius=10, fill=(57, 75, 103), outline=(98, 122, 156), width=2)
        d.rectangle((380, y + 28, width - 980, y + 52), fill=(224, 234, 248))
        d.rectangle((380, y + 64, 560, y + 92), fill=(68, 180, 255))
        d.rectangle((590, y + 64, 780, y + 92), fill=(251, 184, 90))

    # singular exact target block
    x, y = 930, 560
    d.rounded_rectangle((x, y, x + 390, y + 220), radius=12, fill=(82, 105, 138), outline=(198, 214, 236), width=2)
    d.rectangle((x + 28, y + 34, x + 362, y + 62), fill=(244, 248, 254))
    d.rectangle((x + 28, y + 88, x + 172, y + 130), fill=(74, 186, 255))
    d.rectangle((x + 198, y + 88, x + 362, y + 130), fill=(252, 189, 102))
    d.rectangle((x + 28, y + 164, x + 362, y + 194), fill=(31, 42, 62))

    benchmark = source.copy()
    previews = [("anti-alias + compression", jpeg_roundtrip(source.filter(ImageFilter.GaussianBlur(0.35)), 68))]
    return source, benchmark, previews


def build_repetitive_grid_camouflage(width: int, height: int) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]], list[tuple[int, int, int, int]]]:
    base_tile = variant_tile(180, variant="base")
    source = Image.new("RGB", (width, height), (198, 202, 210))
    for y in range(0, height, 156):
        for x in range(0, width, 156):
            source.paste(base_tile, (x, y))

    # three subtle-but-unique targets within the repeating grid
    target_coords = [
        (int(width * 0.18), int(height * 0.30), "target-a"),
        (int(width * 0.49), int(height * 0.56), "target-b"),
        (int(width * 0.73), int(height * 0.34), "target-c"),
    ]
    target_regions: list[tuple[int, int, int, int]] = []
    for tx, ty, variant in target_coords:
        tile = variant_tile(210, variant=variant)
        tx = max(24, min(width - 234, tx))
        ty = max(24, min(height - 234, ty))
        source.paste(tile, (tx, ty))
        target_regions.append((tx, ty, 210, 210))

    benchmark = add_noise(source.filter(ImageFilter.GaussianBlur(0.5)), amount=0.04, seed=71)
    previews = [("camouflage benchmark", benchmark)]
    return source, benchmark, previews, target_regions


def build_noise_stress_random(width: int, height: int, base: Image.Image) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]]]:
    source = ImageEnhance.Contrast(ImageOps.grayscale(fit(base, width, height)).convert("RGB")).enhance(1.2)
    benchmark = add_noise(source, amount=0.16, seed=81)
    d = ImageDraw.Draw(benchmark)
    rnd = random.Random(81)
    for _ in range(2200):
        x1, y1 = rnd.randrange(width), rnd.randrange(height)
        x2, y2 = x1 + rnd.randrange(-26, 27), y1 + rnd.randrange(-26, 27)
        c = rnd.randrange(30, 240)
        d.line((x1, y1, x2, y2), fill=(c, c, c), width=rnd.randrange(1, 3))
    benchmark = jpeg_roundtrip(benchmark, 46)
    previews = [("noise + compression", benchmark)]
    return source, benchmark, previews


def build_orb_feature_rich(width: int, height: int, base: Image.Image) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]]]:
    source = fit(base, width, height)
    source = ImageEnhance.Sharpness(ImageEnhance.Contrast(ImageOps.grayscale(source).convert("RGB")).enhance(1.65)).enhance(2.0)
    benchmark = add_noise(source, amount=0.09, seed=91)
    benchmark = ImageEnhance.Contrast(benchmark.filter(ImageFilter.GaussianBlur(0.5))).enhance(1.1)
    previews = [("feature noise", benchmark)]
    return source, benchmark, previews


def build_perspective_skew_sweep(width: int, height: int, base: Image.Image) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]]]:
    source = fit(base, width, height)
    # one global perspective-skewed preview, same output size
    w, h = source.size
    quad = (
        120, 90,
        w - 140, 60,
        w - 40, h - 40,
        80, h - 70,
    )
    preview = source.transform((w, h), Image.Transform.QUAD, quad, resample=Image.Resampling.BICUBIC)
    benchmark = preview.copy()
    previews = [("perspective sweep", preview)]
    return source, benchmark, previews


def build_scale_rotate_sweep(width: int, height: int, base: Image.Image) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]]]:
    source = fit(base, width, height)
    rotated = source.rotate(17.0, resample=Image.Resampling.BICUBIC, expand=False)
    benchmark = rotated.copy()
    previews = [("global rotate", rotated)]
    return source, benchmark, previews


def draw_card(draw: ImageDraw.ImageDraw, x: int, y: int, w: int, h: int, palette: dict[str, tuple[int, int, int]], variant: int = 0) -> None:
    v = variant % 6
    radius = 8 + (v % 4)
    draw.rounded_rectangle((x, y, x + w, y + h), radius=radius, fill=palette["card"], outline=palette["stroke"], width=2)

    pad_x = max(12, int(w * 0.05))
    title_top = y + max(14, int(h * 0.10))
    title_h = max(16, int(h * 0.12))
    body_top = y + max(42, int(h * 0.34))
    body_h = max(24, int(h * 0.22))
    footer_top = y + h - max(40, int(h * 0.24))
    footer_h = max(14, int(h * 0.12))

    # deterministic per-card layout profile so all cards differ.
    title_right_gap = [0.08, 0.20, 0.13, 0.05, 0.16, 0.10][v]
    left_ratio = [0.42, 0.28, 0.36, 0.31, 0.47, 0.34][v]
    right_ratio = [0.36, 0.48, 0.40, 0.52, 0.30, 0.46][v]
    footer_ratio = [0.92, 0.72, 0.84, 0.65, 0.88, 0.78][v]

    title_right = x + w - pad_x - int((w - 2 * pad_x) * title_right_gap)
    draw.rectangle((x + pad_x, title_top, title_right, title_top + title_h), fill=palette["line1"])

    body_w = w - 2 * pad_x
    left_w = int(body_w * left_ratio)
    right_w = int(body_w * right_ratio)
    left_x1 = x + pad_x
    left_x2 = min(left_x1 + left_w, x + w - pad_x - 6)
    right_x2 = x + w - pad_x
    right_x1 = max(right_x2 - right_w, left_x2 + 8)
    draw.rectangle((left_x1, body_top, left_x2, body_top + body_h), fill=palette["line2"])
    draw.rectangle((right_x1, body_top, right_x2, body_top + body_h), fill=palette["line3"])

    footer_w = int((w - 2 * pad_x) * footer_ratio)
    draw.rectangle((x + pad_x, footer_top, x + pad_x + footer_w, footer_top + footer_h), fill=palette["line4"])

    # distinct per-card UI markers without changing global palette.
    if v in (0, 3):
        marker = max(8, int(min(w, h) * 0.06))
        mx = x + w - pad_x - marker
        my = body_top + body_h + max(8, int(h * 0.06))
        draw.ellipse((mx, my, mx + marker, my + marker), fill=palette["line1"])
    elif v in (1, 4):
        bx = x + pad_x
        by = body_top + body_h + max(8, int(h * 0.06))
        bh = max(8, int(h * 0.06))
        draw.rectangle((bx, by, bx + int(body_w * 0.24), by + bh), fill=palette["line1"])
        draw.rectangle((bx + int(body_w * 0.28), by, bx + int(body_w * 0.48), by + bh), fill=palette["line2"])
    else:
        sx = x + w - pad_x - int(body_w * 0.22)
        sy = body_top + body_h + max(8, int(h * 0.06))
        sh = max(10, int(h * 0.08))
        draw.rounded_rectangle((sx, sy, sx + int(body_w * 0.22), sy + sh), radius=sh // 2, outline=palette["line1"], width=2)
        knob_x = sx + int(body_w * 0.10)
        draw.ellipse((knob_x, sy + 2, knob_x + sh - 4, sy + sh - 2), fill=palette["line2"])


def paste_scaled_panel(canvas: Image.Image, panel: Image.Image, rect: tuple[int, int, int, int], scale_factor: float) -> None:
    ox, oy, rw, rh, _ = panel_render_geometry(panel.width, panel.height, rect, scale_factor)
    resized = panel.resize((rw, rh), Image.Resampling.LANCZOS)
    canvas.paste(resized, (ox, oy))


def panel_render_geometry(panel_w: int, panel_h: int, rect: tuple[int, int, int, int], scale_factor: float) -> tuple[int, int, int, int, float]:
    x, y, w, h = rect
    if w <= 0 or h <= 0:
        return x, y, 1, 1, 1.0
    fit_scale = min(w / max(1, panel_w), h / max(1, panel_h))
    render_scale = max(0.1, fit_scale * scale_factor)
    rw = max(1, int(panel_w * render_scale))
    rh = max(1, int(panel_h * render_scale))
    ox = x + (w - rw) // 2
    oy = y + (h - rh) // 2
    return ox, oy, rw, rh, render_scale


def map_regions_with_panel_geometry(
    regions: list[tuple[int, int, int, int]],
    panel_x: int,
    panel_y: int,
    panel_w: int,
    panel_h: int,
    rect: tuple[int, int, int, int],
    scale_factor: float,
) -> list[tuple[int, int, int, int]]:
    ox, oy, _, _, render_scale = panel_render_geometry(panel_w, panel_h, rect, scale_factor)
    mapped: list[tuple[int, int, int, int]] = []
    for rx, ry, rw, rh in regions:
        rel_x = max(0, rx - panel_x)
        rel_y = max(0, ry - panel_y)
        mx = int(round(ox + rel_x * render_scale))
        my = int(round(oy + rel_y * render_scale))
        mw = max(1, int(round(rw * render_scale)))
        mh = max(1, int(round(rh * render_scale)))
        mapped.append((mx, my, mw, mh))
    return mapped


def transform_regions_with_mask(
    regions: list[tuple[int, int, int, int]],
    width: int,
    height: int,
    transform_fn,
) -> list[tuple[int, int, int, int]]:
    mapped: list[tuple[int, int, int, int]] = []
    for x, y, w, h in regions:
        mask = Image.new("L", (width, height), 0)
        d = ImageDraw.Draw(mask)
        d.rectangle((x, y, x + w, y + h), fill=255)
        out = transform_fn(mask)
        bbox = out.getbbox()
        if not bbox:
            mapped.append((x, y, w, h))
            continue
        x1, y1, x2, y2 = bbox
        mapped.append((x1, y1, max(1, x2 - x1), max(1, y2 - y1)))
    return mapped


def rotate_regions(
    regions: list[tuple[int, int, int, int]],
    width: int,
    height: int,
    angle_deg: float,
) -> list[tuple[int, int, int, int]]:
    return transform_regions_with_mask(
        regions,
        width,
        height,
        lambda img: img.rotate(angle_deg, resample=Image.Resampling.BICUBIC, expand=False),
    )


def perspective_regions(
    regions: list[tuple[int, int, int, int]],
    width: int,
    height: int,
    quad: tuple[int, int, int, int, int, int, int, int],
) -> list[tuple[int, int, int, int]]:
    return transform_regions_with_mask(
        regions,
        width,
        height,
        lambda img: img.transform((width, height), Image.Transform.QUAD, quad, resample=Image.Resampling.BICUBIC),
    )


def build_multi_monitor_dpi_shift(
    width: int, height: int
) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]], list[tuple[int, int, int, int]], list[tuple[int, int, int, int]]]:
    source = smooth_background(width, height, (224, 232, 246), (210, 220, 236))
    d = ImageDraw.Draw(source)

    palette = {
        "card": (238, 244, 252),
        "stroke": (124, 142, 170),
        "line1": (78, 96, 124),
        "line2": (74, 182, 255),
        "line3": (255, 186, 98),
        "line4": (44, 58, 84),
    }

    # Source tab: single screen only.
    panel_w = int(width * 0.68)
    panel_h = int(height * 0.86)
    panel_x = (width - panel_w) // 2
    panel_y = (height - panel_h) // 2
    d.rounded_rectangle((panel_x, panel_y, panel_x + panel_w, panel_y + panel_h), radius=18, fill=(234, 241, 251), outline=(140, 158, 186), width=2)
    d.rectangle((panel_x, panel_y, panel_x + panel_w, panel_y + 64), fill=(214, 226, 242))

    card_w = int(panel_w * 0.58)
    card_h = int(panel_h * 0.20)
    cx = panel_x + int(panel_w * 0.10)
    cy = panel_y + int(panel_h * 0.12)
    gap = int(panel_h * 0.10)
    card_positions = [
        (cx, cy),
        (cx, cy + card_h + gap),
        (cx, cy + (card_h + gap) * 2),
    ]
    regions: list[tuple[int, int, int, int]] = []
    for idx, (x, y) in enumerate(card_positions):
        draw_card(d, x, y, card_w, card_h, palette, variant=idx)
        if idx < 3:
            regions.append((x + int(card_w * 0.07), y + int(card_h * 0.10), int(card_w * 0.80), int(card_h * 0.72)))

    # Preview tab: dual monitors with different DPI/UI scale rendering from the single-screen source.
    preview = smooth_background(width, height, (224, 232, 246), (210, 220, 236))
    pd = ImageDraw.Draw(preview)
    gap_px = max(8, width // 120)
    left_w = (width - gap_px) // 2
    right_w = width - gap_px - left_w
    left_rect = (0, 0, left_w, height)
    right_rect = (left_w + gap_px, 0, right_w, height)
    pd.rectangle((left_w, 0, left_w + gap_px, height), fill=(156, 170, 194))

    panel = source.crop((panel_x, panel_y, panel_x + panel_w, panel_y + panel_h))
    paste_scaled_panel(preview, panel, left_rect, 0.98)
    paste_scaled_panel(preview, panel, right_rect, 1.22)

    left_mapped = map_regions_with_panel_geometry(regions, panel_x, panel_y, panel_w, panel_h, left_rect, 0.98)
    right_mapped = map_regions_with_panel_geometry(regions, panel_x, panel_y, panel_w, panel_h, right_rect, 1.22)
    benchmark_regions: list[tuple[int, int, int, int]] = []
    for idx in range(len(regions)):
        if idx == 1:
            benchmark_regions.append(right_mapped[idx])
        elif idx == 2:
            benchmark_regions.append(right_mapped[idx])
        else:
            benchmark_regions.append(left_mapped[idx])

    benchmark = preview.copy()
    previews = [("dpi/layout shift preview", preview)]
    return source, benchmark, previews, regions, benchmark_regions


def build_photo_clutter() -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]]]:
    source = Image.open(PHOTO_CLUTTER_PATH).convert("RGB")
    benchmark = source.copy()
    previews = [("original", source)]
    return source, benchmark, previews


def build_hybrid_gate_conflicts(width: int, height: int, base: Image.Image) -> tuple[Image.Image, Image.Image, list[tuple[str, Image.Image]]]:
    source = fit(base, width, height)
    source = ImageEnhance.Color(source).enhance(0.7)
    d = ImageDraw.Draw(source)
    rnd = random.Random(117)
    for _ in range(28):
        x = rnd.randrange(80, width - 380)
        y = rnd.randrange(80, height - 250)
        w = rnd.randrange(220, 380)
        h = rnd.randrange(130, 240)
        if rnd.random() < 0.5:
            d.rounded_rectangle((x, y, x + w, y + h), radius=14, fill=(230, 236, 246), outline=(76, 92, 118), width=2)
            d.rectangle((x + 22, y + 22, x + w - 24, y + 42), fill=(60, 80, 112))
        else:
            d.rectangle((x, y, x + w, y + h), fill=(108, 118, 136), outline=(56, 68, 86), width=2)
            for ix in range(x + 8, x + w - 8, 12):
                d.line((ix, y + 6, ix, y + h - 6), fill=(128, 138, 154), width=1)

    benchmark = add_noise(ImageEnhance.Contrast(source).enhance(1.1), amount=0.06, seed=121)
    previews = [("hybrid conflict blend", benchmark)]
    return source, benchmark, previews


def save_bundle(
    sid: str,
    source: Image.Image,
    benchmark: Image.Image,
    previews: list[tuple[str, Image.Image]],
) -> tuple[str, str, list[dict[str, str]]]:
    source_path = SOURCE_DIR / f"{sid}.png"
    bench_path = BENCH_DIR / f"{sid}.png"
    save_image(source_path, source)
    save_image(bench_path, benchmark)

    preview_entries: list[dict[str, str]] = []
    for idx, (label, img) in enumerate(previews, start=1):
        safe = f"{sid}-{idx:02d}.png"
        preview_path = PREVIEW_DIR / safe
        save_image(preview_path, img)
        preview_entries.append({"label": label, "image_path": rel(preview_path)})

    return rel(source_path), rel(bench_path), preview_entries


def scenario_to_entry(
    sid: str,
    source_path: str,
    benchmark_path: str,
    preview_entries: list[dict[str, str]],
    targets: list[tuple[int, int, int, int]],
    benchmark_targets: list[tuple[int, int, int, int]] | None = None,
) -> dict:
    labels = ["primary", "secondary", "tertiary"]
    out_targets = []
    for idx, (x, y, w, h) in enumerate(targets[:3], start=1):
        out_targets.append(
            {
                "id": f"target-{idx:02d}",
                "label": labels[idx - 1],
                "x": int(x),
                "y": int(y),
                "w": int(w),
                "h": int(h),
            }
        )

    out_benchmark_targets = []
    if benchmark_targets:
        for idx, (x, y, w, h) in enumerate(benchmark_targets[:3], start=1):
            out_benchmark_targets.append(
                {
                    "id": f"target-{idx:02d}",
                    "label": labels[idx - 1],
                    "x": int(x),
                    "y": int(y),
                    "w": int(w),
                    "h": int(h),
                }
            )

    entry = {
        "id": sid,
        "scenario_type_ids": [sid],
        "image_path": source_path,
        "source_image_path": source_path,
        "benchmark_image_path": benchmark_path,
        "preview_images": preview_entries,
        "targets": out_targets,
    }
    if out_benchmark_targets:
        entry["benchmark_targets"] = out_benchmark_targets
    return entry


def main() -> None:
    args = parse_args()
    width = max(640, args.width)
    height = max(360, args.height)

    ensure_dirs()

    sikulix_bg = ensure_sikulix_background(args.offline)
    base_images = {name: fetch_picsum(pid, width, height, args.offline) for name, pid in BASE_IDS.items()}

    entries: dict[str, dict] = {}

    source, benchmark, previews = build_vector_ui_baseline(width, height, sikulix_bg)
    source_path, bench_path, preview_entries = save_bundle("vector_ui_baseline", source, benchmark, previews)
    targets = pick_regions(source, 3)
    entries["vector_ui_baseline"] = scenario_to_entry("vector_ui_baseline", source_path, bench_path, preview_entries, targets, targets)

    source, benchmark, previews = build_template_control_exact(width, height)
    source_path, bench_path, preview_entries = save_bundle("template_control_exact", source, benchmark, previews)
    targets = pick_regions(source, 3)
    entries["template_control_exact"] = scenario_to_entry("template_control_exact", source_path, bench_path, preview_entries, targets, targets)

    source, benchmark, previews, manual_targets = build_repetitive_grid_camouflage(width, height)
    source_path, bench_path, preview_entries = save_bundle("repetitive_grid_camouflage", source, benchmark, previews)
    entries["repetitive_grid_camouflage"] = scenario_to_entry("repetitive_grid_camouflage", source_path, bench_path, preview_entries, manual_targets, manual_targets)

    source, benchmark, previews = build_noise_stress_random(width, height, base_images["noise_stress_random"])
    source_path, bench_path, preview_entries = save_bundle("noise_stress_random", source, benchmark, previews)
    targets = pick_regions(source, 3)
    entries["noise_stress_random"] = scenario_to_entry("noise_stress_random", source_path, bench_path, preview_entries, targets, targets)

    source, benchmark, previews = build_orb_feature_rich(width, height, base_images["orb_feature_rich"])
    source_path, bench_path, preview_entries = save_bundle("orb_feature_rich", source, benchmark, previews)
    targets = pick_regions(source, 3)
    entries["orb_feature_rich"] = scenario_to_entry("orb_feature_rich", source_path, bench_path, preview_entries, targets, targets)

    source, benchmark, previews = build_perspective_skew_sweep(width, height, base_images["perspective_skew_sweep"])
    source_path, bench_path, preview_entries = save_bundle("perspective_skew_sweep", source, benchmark, previews)
    targets = pick_regions(source, 3)
    perspective_quad = (
        120, 90,
        width - 140, 60,
        width - 40, height - 40,
        80, height - 70,
    )
    entries["perspective_skew_sweep"] = scenario_to_entry(
        "perspective_skew_sweep",
        source_path,
        bench_path,
        preview_entries,
        targets,
        perspective_regions(targets, width, height, perspective_quad),
    )

    source, benchmark, previews = build_scale_rotate_sweep(width, height, base_images["scale_rotate_sweep"])
    source_path, bench_path, preview_entries = save_bundle("scale_rotate_sweep", source, benchmark, previews)
    targets = pick_regions(source, 3)
    entries["scale_rotate_sweep"] = scenario_to_entry(
        "scale_rotate_sweep",
        source_path,
        bench_path,
        preview_entries,
        targets,
        rotate_regions(targets, width, height, 17.0),
    )

    source, benchmark, previews, manual_targets, bench_targets = build_multi_monitor_dpi_shift(width, height)
    source_path, bench_path, preview_entries = save_bundle("multi_monitor_dpi_shift", source, benchmark, previews)
    entries["multi_monitor_dpi_shift"] = scenario_to_entry("multi_monitor_dpi_shift", source_path, bench_path, preview_entries, manual_targets, bench_targets)

    source, benchmark, previews = build_hybrid_gate_conflicts(width, height, base_images["hybrid_gate_conflicts"])
    source_path, bench_path, preview_entries = save_bundle("hybrid_gate_conflicts", source, benchmark, previews)
    targets = pick_regions(source, 3)
    entries["hybrid_gate_conflicts"] = scenario_to_entry("hybrid_gate_conflicts", source_path, bench_path, preview_entries, targets, targets)

    photo_source, photo_benchmark, _ = build_photo_clutter()
    photo_targets = pick_regions(photo_source, 3)
    entries["photo_clutter"] = {
        "id": "photo_clutter",
        "scenario_type_ids": ["photo_clutter"],
        "image_path": rel(PHOTO_CLUTTER_PATH),
        "source_image_path": rel(PHOTO_CLUTTER_PATH),
        "benchmark_image_path": rel(PHOTO_CLUTTER_PATH),
        "preview_images": [{"label": "original", "image_path": rel(PHOTO_CLUTTER_PATH)}],
        "targets": [
            {
                "id": f"target-{idx + 1:02d}",
                "label": ["primary", "secondary", "tertiary"][idx],
                "x": int(r[0]),
                "y": int(r[1]),
                "w": int(r[2]),
                "h": int(r[3]),
            }
            for idx, r in enumerate(photo_targets)
        ],
        "benchmark_targets": [
            {
                "id": f"target-{idx + 1:02d}",
                "label": ["primary", "secondary", "tertiary"][idx],
                "x": int(r[0]),
                "y": int(r[1]),
                "w": int(r[2]),
                "h": int(r[3]),
            }
            for idx, r in enumerate(photo_targets)
        ],
    }

    if not args.skip_regions:
        doc = {"schema_version": "1.0.0", "images": [entries[sid] for sid in SCENARIO_IDS]}
        with REGIONS_PATH.open("w", encoding="utf-8") as handle:
            json.dump(doc, handle, indent=2)
            handle.write("\n")

    print("Generated scenario assets:")
    for sid in SCENARIO_IDS:
        entry = entries[sid]
        print(f"  {sid}: source={entry['source_image_path']} benchmark={entry['benchmark_image_path']} previews={len(entry.get('preview_images', []))}")
    if not args.skip_regions:
        print(f"Updated region spec: {REGIONS_PATH.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
