#!/usr/bin/env python3

"""Convert CSS hex/rgb colors to HSL.

Usage examples:
  python scripts/color_to_hsl.py --color "#1a65b8"
  python scripts/color_to_hsl.py --color "rgb(26, 101, 184)"
  python scripts/color_to_hsl.py --scan static/css/base.css
"""

from __future__ import annotations

import argparse
import colorsys
import re
from pathlib import Path

HEX_RE = re.compile(r"#[0-9a-fA-F]{3,8}\b")
RGB_RE = re.compile(
    r"rgba?\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})(?:\s*,\s*(0|0?\.\d+|1(?:\.0+)?))?\s*\)",
    re.IGNORECASE,
)


def rgb_to_hsl(r: int, g: int, b: int) -> tuple[int, int, int]:
    h, l, s = colorsys.rgb_to_hls(r / 255, g / 255, b / 255)
    return round(h * 360) % 360, round(s * 100), round(l * 100)


def parse_hex(value: str) -> tuple[int, int, int]:
    h = value.strip().lstrip("#")
    if len(h) == 3:
        r, g, b = [int(ch * 2, 16) for ch in h]
        return r, g, b
    if len(h) == 6:
        return int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16)
    raise ValueError(f"Unsupported hex color: {value}")


def to_hsl(value: str) -> str:
    v = value.strip()
    if v.startswith("#"):
        r, g, b = parse_hex(v)
        h, s, l = rgb_to_hsl(r, g, b)
        return f"hsl({h} {s}% {l}%)"

    m = RGB_RE.fullmatch(v)
    if m:
        r, g, b = (int(m.group(1)), int(m.group(2)), int(m.group(3)))
        for c in (r, g, b):
            if c < 0 or c > 255:
                raise ValueError(f"Invalid rgb color: {value}")
        h, s, l = rgb_to_hsl(r, g, b)
        return f"hsl({h} {s}% {l}%)"

    raise ValueError(f"Unsupported color format: {value}")


def scan_file(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    colors = {m.group(0) for m in HEX_RE.finditer(text)}
    colors.update({m.group(0) for m in RGB_RE.finditer(text)})
    return sorted(colors)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--color", action="append", default=[], help="Color literal (#hex or rgb(...))")
    parser.add_argument("--scan", action="append", default=[], help="Scan file and convert discovered literals")
    args = parser.parse_args()

    inputs = list(args.color)
    for file_path in args.scan:
        inputs.extend(scan_file(Path(file_path)))

    seen = set()
    for raw in inputs:
        if raw in seen:
            continue
        seen.add(raw)
        print(f"{raw} -> {to_hsl(raw)}")


if __name__ == "__main__":
    main()
