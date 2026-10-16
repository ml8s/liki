#!/usr/bin/env python3
"""重新生成四张专家头像的矢量源 `expert-packs/liki-*/avatars/expert.svg`。

四张头像是同一套设计：深蓝圆角底 + 金色田字格 + 一个本域汉字（命 / 卦 / 名 /
宅）。字形取自阿里妈妈刀隶体（`Alimama DaoLiTi`，阿里巴巴免费商用字体）：先用
`inkscape -T` 把字转成正体轮廓路径，再按字高归一化居中，于是矢量源里不出现
`<text>`、字体或 `dominant-baseline`——市场宿主的渲染字体不可控，取字必须落在
路径上（由 `scripts/build-workbuddy-avatars.sh` 的栅格化门守护）。

    python3 scripts/gen_expert_avatar_svgs.py    # 重写四份矢量源（幂等）
    make expert-avatars                          # 再跑一遍栅格化出 expert.png

改字面（换字）只改 `GLYPHS`；改配色/尺寸改本文件顶部的设计常量，然后重跑。
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
FONT_SIZE = 360  # 转路径时的字号；只影响中间精度，最终按 TARGET_HEIGHT 归一化
TARGET_HEIGHT = 190  # 汉字在 512×512 画布上的目标字高
CENTER = 256.0
BACKDROP = "#0C1A2E"
INK = "#FFBF00"

TEMPLATE = """<svg xmlns="http://www.w3.org/2000/svg" width="512" height="512" viewBox="0 0 512 512">
  <rect width="512" height="512" rx="72" fill="{backdrop}"/>
  <rect x="126" y="126" width="260" height="260" rx="18" fill="none" stroke="{ink}" stroke-width="12" opacity="0.8"/>
  <g stroke="{ink}" stroke-width="6" opacity="0.3" stroke-dasharray="14 12">
    <line x1="256" y1="126" x2="256" y2="386"/>
    <line x1="126" y1="256" x2="386" y2="256"/>
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
            scale = TARGET_HEIGHT / height
            tx = round(CENTER - scale * (x0 + width / 2), 4)
            ty = round(CENTER - scale * (y0 + height / 2), 4)
            svg = TEMPLATE.format(
                backdrop=BACKDROP,
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
            print(f"  ✓ {target.relative_to(ROOT)} 「{glyph}」 字高 {height:.1f} → {TARGET_HEIGHT} (×{scale:.4f})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
