#!/usr/bin/env python3
"""重建 Liki 的品牌图形：专家头像与连接器图标。

两族图形同一套母题——金色田字格 + 一个汉字，只是底不同：

    专家头像 512×512  深色圆角底 + 金字   `expert-packs/liki-*/avatars/expert.svg`  命 / 卦 / 宅 / 名
    连接器图标 64×64  透明底 + 单色深蓝   `workbuddy/liki-*-connector/icon.svg`     盘 / 断

头像在列表里按图片渲染，深色圆角牌自带底色、亮暗主题都成立；图标按平台规范
（「背景建议透明」）不铺底，改用品牌深蓝单色，压在亮底卡片上可辨。

汉字取自阿里妈妈刀隶体（`Alimama DaoLiTi`，阿里巴巴免费商用字体，本仓不附带字体
文件）：先用 `inkscape -T` 把字转成正体轮廓路径，再按目标字高归一化居中，于是矢量源
里不出现 `<text>`、字体或 `dominant-baseline`——市场宿主的渲染字体不可控，取字必须
落在路径上（栅格化门见 `scripts/build-workbuddy-avatars.sh`，`make check` 侧见
`scripts/check_expert_docs.py` 与 `scripts/check_workbuddy_connectors.py`）。

    python3 scripts/gen_brand_marks.py    # 重写六份矢量源（幂等）
    make brand-marks                      # 再栅格化头像 PNG（图标按规范直接交 SVG）

换字改 `AVATARS` / `ICONS`；改配色与占比改下面的几何预设，然后重跑。
"""
from __future__ import annotations

import hashlib
import re
import shutil
import subprocess
import tempfile
from dataclasses import dataclass
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

FONT_FAMILY = "Alimama DaoLiTi"
FONT_SIZE = 360  # 转路径时的字号；只影响中间精度，最终按 glyph_height 归一化
BACKDROP = "#0C1A2E"  # 头像的深色圆角底
INK = "#FFBF00"  # 深底上的品牌金
MARK_INK = "#0C1A2E"  # 透明底上的品牌深蓝


@dataclass(frozen=True)
class Preset:
    """一族图形的几何：底、用色、田字格占比、字高。"""

    canvas: float
    backdrop: str | None  # None = 透明底
    backdrop_rx: float
    ink: str
    frame_size: float  # 田字格边长
    frame_rx: float
    frame_stroke: float
    frame_opacity: float
    grid_stroke: float
    grid_dash: str | None  # None = 实线十字
    grid_opacity: float
    glyph_height: float


# 头像按 40~96px 展示：田字格占画布 75%，汉字占田字格内框约 80%（再大就顶格线）。
AVATAR = Preset(canvas=512, backdrop=BACKDROP, backdrop_rx=72, ink=INK, frame_size=384,
                frame_rx=28, frame_stroke=18, frame_opacity=0.8, grid_stroke=8,
                grid_dash="20 16", grid_opacity=0.3, glyph_height=290)
# 图标按 16~64px 展示：比例同头像档位，但透明底、单色深蓝，十字改实线（64px 上的虚线会糊）。
ICON = Preset(canvas=64, backdrop=None, backdrop_rx=14, ink=MARK_INK, frame_size=48,
              frame_rx=5, frame_stroke=3, frame_opacity=1.0, grid_stroke=1.5,
              grid_dash=None, grid_opacity=0.3, glyph_height=36)

# 顶层领域 → 头像用字（见 expert-packs/README.md「头像」）。
AVATARS: tuple[tuple[str, str], ...] = (
    ("expert-packs/liki-natal/avatars/expert.svg", "命"),
    ("expert-packs/liki-divination/avatars/expert.svg", "卦"),
    ("expert-packs/liki-fengshui/avatars/expert.svg", "宅"),
    ("expert-packs/liki-naming/avatars/expert.svg", "名"),
)
# 连接器 → 图标用字（见 workbuddy/README.md「连接器」）。
ICONS: tuple[tuple[str, str], ...] = (
    ("workbuddy/liki-engine-connector/icon.svg", "盘"),
    ("workbuddy/liki-counsel-connector/icon.svg", "断"),
)


def _num(value: float) -> str:
    """去掉浮点尾巴：48.0 → 48，1.5 → 1.5。"""
    return f"{round(value, 5):g}"


def _command(name: str) -> str:
    found = shutil.which(name)
    if not found:
        raise SystemExit(f"[brand-marks] error: 需要 {name}，未找到")
    return found


def _font_file() -> Path:
    """解析字体文件路径：轮廓可追溯，换字体版本时会显式变化。"""
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
            f"[brand-marks] error: 缺少字体 {FONT_FAMILY}"
            "（阿里妈妈刀隶体，免费商用：https://fonts.alibabagroup.com/）"
        )
    return candidates[0]


def _glyph_outline(inkscape: str, glyph: str, workdir: Path) -> tuple[str, tuple[float, ...]]:
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
        raise SystemExit(f"[brand-marks] error: 「{glyph}」未能转成路径（字体缺该字形？）")
    path = re.sub(r"\s+", " ", match.group(1)).strip()
    box = subprocess.run(
        [inkscape, str(plain), "--query-id=glyph", "--query-x", "--query-y", "--query-width", "--query-height"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.split()
    return path, tuple(float(value) for value in box)


def _mark_svg(preset: Preset, glyph: str, path: str, box: tuple[float, ...]) -> str:
    x0, y0, width, height = box
    center = preset.canvas / 2
    near = center - preset.frame_size / 2
    far = center + preset.frame_size / 2
    scale = preset.glyph_height / height
    tx = _num(center - scale * (x0 + width / 2))
    ty = _num(center - scale * (y0 + height / 2))
    dash = f' stroke-dasharray="{preset.grid_dash}"' if preset.grid_dash else ""
    backdrop = ""
    if preset.backdrop:
        backdrop = (
            f'\n  <rect width="{_num(preset.canvas)}" height="{_num(preset.canvas)}"'
            f' rx="{_num(preset.backdrop_rx)}" fill="{preset.backdrop}"/>'
        )
    return f"""<svg xmlns="http://www.w3.org/2000/svg" width="{_num(preset.canvas)}" height="{_num(preset.canvas)}" viewBox="0 0 {_num(preset.canvas)} {_num(preset.canvas)}">{backdrop}
  <rect x="{_num(near)}" y="{_num(near)}" width="{_num(preset.frame_size)}" height="{_num(preset.frame_size)}" rx="{_num(preset.frame_rx)}" fill="none" stroke="{preset.ink}" stroke-width="{_num(preset.frame_stroke)}" opacity="{_num(preset.frame_opacity)}"/>
  <g stroke="{preset.ink}" stroke-width="{_num(preset.grid_stroke)}" opacity="{_num(preset.grid_opacity)}"{dash}>
    <line x1="{_num(center)}" y1="{_num(near)}" x2="{_num(center)}" y2="{_num(far)}"/>
    <line x1="{_num(near)}" y1="{_num(center)}" x2="{_num(far)}" y2="{_num(center)}"/>
  </g>
  <!-- 田字格与「{glyph}」（{FONT_FAMILY}轮廓） -->
  <g aria-label="{glyph}" fill="{preset.ink}" transform="translate({tx},{ty}) scale({_num(scale)})">
    <path d="{path}"/>
  </g>
</svg>
"""


def main() -> int:
    inkscape = _command("inkscape")
    font = _font_file()
    print(f"[brand-marks] 字体 {font.name} sha256:{hashlib.sha256(font.read_bytes()).hexdigest()[:16]}…")

    with tempfile.TemporaryDirectory() as tmp:
        for preset, marks in ((AVATAR, AVATARS), (ICON, ICONS)):
            for relative, glyph in marks:
                target = ROOT / relative
                if not target.parent.is_dir():
                    raise SystemExit(f"[brand-marks] error: {target.parent} 不存在")
                path, box = _glyph_outline(inkscape, glyph, Path(tmp))
                target.write_text(_mark_svg(preset, glyph, path, box), encoding="utf-8")
                print(
                    f"  ✓ {relative} 「{glyph}」 画布 {_num(preset.canvas)}"
                    f" 字高 {preset.glyph_height:g}（原始 {box[3]:.1f}）"
                )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
