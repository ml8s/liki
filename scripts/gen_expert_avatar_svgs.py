#!/usr/bin/env python3
"""重新生成四张专家头像的矢量源 `expert-packs/liki-*/avatars/expert.svg`。

四张头像是同一套设计：深蓝圆角底 + 金色田字格 + 一个本域汉字（命 / 卦 / 名 /
宅）。字形取自阿里妈妈刀隶体（`Alimama DaoLiTi`，阿里巴巴免费商用字体）：先用
`inkscape -T` 把字转成正体轮廓路径，再按字高归一化居中，于是矢量源里不出现
`<text>`、字体或 `dominant-baseline`——市场宿主的渲染字体不可控，取字必须落在
路径上（由 `scripts/build-workbuddy-avatars.sh` 的栅格化门守护）。

    python3 scripts/gen_expert_avatar_svgs.py    # 重写四份矢量源（幂等）
    make expert-avatars                          # 再跑一遍栅格化出 expert.png

改字面（换字）只改 `GLYPHS`；改配色/占比改本文件顶部的设计常量，然后重跑。
"""
from __future__ import annotations

import hashlib
import re
import shutil
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PACKS = ROOT / "expert-packs"

# 顶层领域 → 头像用字（本域汉字，见 expert-packs/README.md「头像」）。
GLYPHS: tuple[tuple[str, str], ...] = (
    ("natal", "命"),
    ("divination", "卦"),
    ("fengshui", "宅"),
    ("naming", "名"),
)

FONT_FAMILY = "Alimama DaoLiTi"
FONT_SIZE = 360  # 转路径时的字号；只影响中间精度，最终按 GLYPH_HEIGHT 归一化
CANVAS = 512.0
CENTER = CANVAS / 2
BACKDROP = "#0C1A2E"
INK = "#FFBF00"

# 田字格与汉字的占比：头像按 40~96px 展示，画布几乎铺满——田字格占 75%，
# 汉字占田字格内框的 ~80%（字再大就会顶到格线）。
BACKDROP_RX = 72.0
FRAME_SIZE = 384.0  # 田字格边长
FRAME_RX = 28.0
FRAME_STROKE = 18.0
GRID_STROKE = 8.0
GRID_DASH = "20 16"
GLYPH_HEIGHT = 290.0  # 汉字目标字高

FRAME = CENTER - FRAME_SIZE / 2
FRAME_FAR = CENTER + FRAME_SIZE / 2

TEMPLATE = """<svg xmlns="http://www.w3.org/2000/svg" width="512" height="512" viewBox="0 0 512 512">
  <rect width="512" height="512" rx="{backdrop_rx}" fill="{backdrop}"/>
  <rect x="{frame}" y="{frame}" width="{frame_size}" height="{frame_size}" rx="{frame_rx}" fill="none" stroke="{ink}" stroke-width="{frame_stroke}" opacity="0.85"/>
  <g stroke="{ink}" stroke-width="{grid_stroke}" opacity="0.3" stroke-dasharray="{grid_dash}">
    <line x1="{center}" y1="{frame}" x2="{center}" y2="{frame_far}"/>
    <line x1="{frame}" y1="{center}" x2="{frame_far}" y2="{center}"/>
  </g>
  <!-- 田字格与「{glyph}」（{font}轮廓） -->
  <g aria-label="{glyph}" fill="{ink}" transform="translate({tx},{ty}) scale({scale})">
    <path d="{path}"/>
  </g>
</svg>
"""


def _command(name: str) -> str:
    found = shutil.which(name)
    if not found:
        raise SystemExit(f"[expert-avatars] error: 需要 {name}，未找到")
    return found


def _font_file() -> Path:
    """解析字体文件路径与指纹：轮廓可追溯，换字体版本时会显式变化。"""
    fc_match = shutil.which("fc-match")
    if fc_match:
        out = subprocess.run(
            [fc_match, "-f", "%{file}", FONT_FAMILY], check=True, capture_output=True, text=True
        ).stdout.strip()
        if out and Path(out).is_file():
            return Path(out)
    candidates = sorted((Path.home() / ".local/share/fonts").glob("AlimamaDaoLiTi*.ttf"))
    candidates += sorted(Path("/usr/share/fonts").rglob("AlimamaDaoLiTi*.ttf"))
    if not candidates:
        raise SystemExit(
            f"[expert-avatars] error: 缺少字体 {FONT_FAMILY}"
            "（阿里妈妈刀隶体，免费商用：https://fonts.alibabagroup.com/）"
        )
    return candidates[0]


def _glyph_outline(inkscape: str, glyph: str, workdir: Path) -> tuple[str, tuple[float, float, float, float]]:
    """把单字转成正体轮廓路径，并返回它在原始坐标系里的包围盒。"""
    src = workdir / "glyph.svg"
    src.write_text(
        '<svg xmlns="http://www.w3.org/2000/svg" width="512" height="512" viewBox="0 0 512 512">'
        f'<text id="glyph" x="0" y="0" font-family="{FONT_FAMILY}" font-size="{FONT_SIZE}"'
        f' fill="#000">{glyph}</text></svg>',
        encoding="utf-8",
    )
    plain = workdir / "glyph-path.svg"
    subprocess.run(
        [inkscape, str(src), "-T", "--export-plain-svg", "--export-filename", str(plain)],
        check=True,
        capture_output=True,
    )
    match = re.search(r'<path\s+d="([^"]+)"', plain.read_text(encoding="utf-8"))
    if not match:
        raise SystemExit(f"[expert-avatars] error: 「{glyph}」未能转成路径（字体缺该字形？）")
    path = re.sub(r"\s+", " ", match.group(1)).strip()
    box = subprocess.run(
        [inkscape, str(plain), "--query-id=glyph", "--query-x", "--query-y", "--query-width", "--query-height"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.split()
    return path, tuple(float(value) for value in box)


def main() -> int:
    inkscape = _command("inkscape")
    font = _font_file()
    digest = hashlib.sha256(font.read_bytes()).hexdigest()[:16]
    print(f"[expert-avatars] 字体 {font.name} sha256:{digest}…")

    with tempfile.TemporaryDirectory() as tmp:
        for domain, glyph in GLYPHS:
            path, (x0, y0, width, height) = _glyph_outline(inkscape, glyph, Path(tmp))
            scale = GLYPH_HEIGHT / height
            tx = round(CENTER - scale * (x0 + width / 2), 4)
            ty = round(CENTER - scale * (y0 + height / 2), 4)
            svg = TEMPLATE.format(
                backdrop=BACKDROP,
                backdrop_rx=BACKDROP_RX,
                frame=FRAME,
                frame_size=FRAME_SIZE,
                frame_rx=FRAME_RX,
                frame_stroke=FRAME_STROKE,
                frame_far=FRAME_FAR,
                grid_stroke=GRID_STROKE,
                grid_dash=GRID_DASH,
                center=CENTER,
                ink=INK,
                glyph=glyph,
                font=FONT_FAMILY,
                path=path,
                tx=tx,
                ty=ty,
                scale=round(scale, 5),
            )
            target = PACKS / f"liki-{domain}" / "avatars" / "expert.svg"
            if not target.parent.is_dir():
                raise SystemExit(f"[expert-avatars] error: {target.parent} 不存在")
            target.write_text(svg, encoding="utf-8")
            print(f"  ✓ {target.relative_to(ROOT)} 「{glyph}」 字高 {height:.1f} → {GLYPH_HEIGHT:.0f} (×{scale:.4f})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
