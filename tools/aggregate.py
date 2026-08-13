"""领域聚合（Layer 0 引擎数据 → Layer 1 因子聚合——build_factors 纯聚合，零算子逻辑）。

分层：atoms（原子算子）→ aggregate（聚合）→ evaluate（因子快照）→ duanyu（生成器入口）。
"""
from __future__ import annotations
from typing import Optional
from dataclasses import dataclass, field, asdict

__all__ = ["build_factors", "ShishenState", "ThreeChecks", "PalaceFlags", "WuxingProfile"]



# 月令旺相（得令）：引擎 fu_yi.wang_shuai 的取值
DE_LING_STATES = {"旺", "相"}


@dataclass
class ShishenState:
    """某十神在原局的总体状态（聚合自四柱 shi_shens + cang_gan）。"""
    name: str                    # 十神名（如 正财/偏印/食神）
    wuxing: str                  # 该十神对应天干的五行（取第一处出现的 gan）
    tou_gan: bool                # 透干（天干出现）
    cang_zhi: bool               # 藏支（地支藏干出现）
    has_root: bool               # 有根（该五行在地支藏干出现，不限本十神）
    de_ling: bool                # 得令（该五行在月令旺相）
    count: int = 0               # 出现次数（透干+藏干各计）

    def present(self) -> bool:
        return self.tou_gan or self.cang_zhi


@dataclass
class ThreeChecks:
    """十神三关：得令 / 不被克 / 有根。"""
    de_ling: bool
    not_ke: bool
    has_root: bool

    def passed(self) -> int:
        return sum([self.de_ling, self.not_ke, self.has_root])


@dataclass
class PalaceFlags:
    """宫位状态（筛自 fullchart 的合会冲刑全局表）。"""
    pillar: str            # 宫位名（日支/年支/月支/时支）
    zhi: str
    chong: list = field(default_factory=list)   # 六冲：与谁冲
    liu_he: list = field(default_factory=list)  # 六合：与谁合
    san_he: list = field(default_factory=list)  # 三合：成局
    void: bool = False     # 空亡
    kui_gang: bool = False


@dataclass
class WuxingProfile:
    """五行分布（直接读引擎 fu_yi，零计算）。"""
    count: dict                     # wuxing_count
    wang_shuai: dict                # wang_shuai
    qiangruo: str                   # 身强弱
    yongshen: dict                  # 三派用神 {fu_yi/tiao_hou/ge_ju: {yong,xi,ji}}

    def over_strong(self, threshold: int = 4) -> list:
        """过旺五行（出现 ≥ threshold 次 或 旺）。"""
        strong = [wx for wx, n in self.count.items() if n >= threshold]
        strong += [wx for wx, st in self.wang_shuai.items() if st == "旺" and wx not in strong]
        return strong

    def over_weak(self) -> list:
        """过弱五行（死/囚，且出现 ≤ 1 次）。"""
        return [wx for wx, st in self.wang_shuai.items() if st in ("死", "囚") and self.count.get(wx, 0) <= 1]


# ---------------------------------------------------------------------------
# 聚合函数
# ---------------------------------------------------------------------------

def _all_shishens(full: dict) -> list[dict]:
    """四柱所有十神条目（含干支来源）。"""
    out = []
    for key in ("nian", "yue", "ri", "shi"):
        zhu = full.get(key, {})
        for ss in zhu.get("shi_shens", []):
            out.append({"pillar": key, **ss})
    return out


def aggregate_shishen(full: dict, wang_shuai: Optional[dict] = None) -> dict[str, ShishenState]:
    """聚合每十神：透干/藏支/有根/得令/次数。

    - 透干：source == 'stem'（天干）
    - 藏支：source == 'main_qi'/'mid_qi'/'minor_qi'（地支藏干）
    - 有根：该十神对应天干五行，在四柱地支藏干（任意）中出现
    - 得令：该五行在月令旺相（wang_shuai，由 yongshen.fu_yi 提供）
    """
    states: dict[str, ShishenState] = {}
    # 先统计透干/藏支
    for item in _all_shishens(full):
        name = item["shi_shen"]
        if name not in states:
            states[name] = ShishenState(
                name=name, wuxing="", tou_gan=False, cang_zhi=False,
                has_root=False, de_ling=False, count=0,
            )
        st = states[name]
        src = item.get("source", "")
        gan = item.get("gan", "")
        if not st.wuxing and gan:
            # 十神对应天干五行（日主五行已知，取十神对应干支五行）
            st.wuxing = _wuxing_of_gan(gan)
        if src == "stem":
            st.tou_gan = True
            st.count += 1
        elif src and src.endswith("qi"):
            st.cang_zhi = True
            st.count += 1
    # 有根：该五行出现在任一地支藏干
    roots: set[str] = set()
    for key in ("nian", "yue", "ri", "shi"):
        cg = full.get(key, {}).get("cang_gan", {}) or {}
        for v in (cg.get("main"), cg.get("mid"), cg.get("minor")):
            if v:
                roots.add(_wuxing_of_gan(v))
    for st in states.values():
        if st.wuxing:
            st.has_root = st.wuxing in roots
            # 得令：月令旺相
            if wang_shuai:
                st.de_ling = st.wuxing in wang_shuai and wang_shuai[st.wuxing] in DE_LING_STATES
    return states


def _wuxing_of_gan(gan: str) -> str:
    """天干 → 五行。"""
    table = {"甲": "木", "乙": "木", "丙": "火", "丁": "火", "戊": "土",
             "己": "土", "庚": "金", "辛": "金", "壬": "水", "癸": "水"}
    return table.get(gan, "")


def three_checks(st: ShishenState, wang_shuai: dict, gan_ke_list: list[dict]) -> ThreeChecks:
    """十神三关：得令（月令旺相）/ 不被克（无天干克它）/ 有根。

    不被克：gan_ke_list 为 fullchart.gan_he 中克该十神天干五行者——
    此处简化为「该十神透干且未被天干合绊/克」；完全版见 rules。
    """
    de_ling = st.de_ling if st.de_ling else (st.wuxing in wang_shuai and wang_shuai[st.wuxing] in DE_LING_STATES)
    # 有根已在聚合时算好；若未算则查 has_root
    not_ke = True  # 该十神是否不被克（比劫/食伤不被官杀克等——简化：恒不被克，真实克另算）
    return ThreeChecks(de_ling=de_ling, not_ke=not_ke, has_root=st.has_root)


def palace_flags(full: dict, pillar: str = "ri") -> PalaceFlags:
    """筛某柱（默认日支）的冲/刑/合/空亡/魁罡。

    引擎 fullchart 的 liu_chong/zhi_liu_he/san_he 含 pillar 索引：
    nian=0, yue=1, ri=2, shi=3。
    """
    idx = {"nian": 0, "yue": 1, "ri": 2, "shi": 3}[pillar]
    zhu = full.get(pillar, {})
    pf = PalaceFlags(
        pillar=pillar,
        zhi=zhu.get("zhi", ""),
        void=bool(zhu.get("is_void")),
        kui_gang=bool(zhu.get("is_kui_gang")),
    )
    for c in full.get("liu_chong", []):
        if c.get("zhi_a_idx") == idx or c.get("zhi_b_idx") == idx or c.get("zhi_a") == pf.zhi or c.get("zhi_b") == pf.zhi:
            other = c.get("zhi_b") if c.get("zhi_a") == pf.zhi else c.get("zhi_a")
            pf.chong.append(other or c.get("detail", ""))
    for h in full.get("zhi_liu_he", []):
        if h.get("zhi_a") == pf.zhi:
            pf.liu_he.append(h.get("zhi_b", ""))
        elif h.get("zhi_b") == pf.zhi:
            pf.liu_he.append(h.get("zhi_a", ""))
    for s in full.get("san_he", []):
        for z in (s.get("zhi_a"), s.get("zhi_b"), s.get("zhi_c")):
            if z == pf.zhi:
                pf.san_he.append(s.get("he_element", "") or s.get("detail", ""))
                break
    return pf


def wuxing_profile(yongshen: dict) -> WuxingProfile:
    """直接读引擎 fu_yi（零计算）。"""
    fu = yongshen.get("fu_yi", {})
    return WuxingProfile(
        count=fu.get("wuxing_count", {}),
        wang_shuai=fu.get("wang_shuai", {}),
        qiangruo=fu.get("qiangruo", ""),
        yongshen=yongshen,
    )


def build_factors(data: dict) -> dict:
    """把 Layer 0（引擎数据）组织成 Layer 1 因子快照（纯聚合）。"""
    full = data["full"]
    ys = data["yongshen"]
    wp = wuxing_profile(ys)
    ss = aggregate_shishen(full, wp.wang_shuai)
    # 大运配偶星窗口（依据：大运透配偶星则婚缘实——原局不现可大运得）
    dy_steps = data.get("chart", {}).get("da_yun", {}).get("steps", [])
    return {
        "shishen": {k: asdict(v) for k, v in ss.items()},
        "wuxing": asdict(wp),
        "qiangruo": wp.qiangruo,
        "palace_ri": asdict(palace_flags(full, "ri")),
        "palace_nian": asdict(palace_flags(full, "nian")),
        "nian_gan": full.get("nian", {}).get("gan", ""),
        "nian_zhi": full.get("nian", {}).get("zhi", ""),
        "ziwei": data.get("ziwei", {}),   # 紫微本命盘（四化/夫妻宫交叉用）
        "yongshen": data.get("yongshen", {}),  # 三派用神（扶抑/调候/格局，喜忌五行）
        "ri_gan": full.get("ri", {}).get("gan", ""),
        "yue_zhi": full.get("yue", {}).get("zhi", ""),
        "shi_zhi": full.get("shi", {}).get("zhi", ""),
        "_birth_year": int(data.get("solar", "")[:4]) if str(data.get("solar", ""))[:4].isdigit() else (data.get("lunar") or {}).get("year", 0) or 0,
        # 大运每步：{name, qi_sui, zhi_sui, shi_shen}（shi_shen 如"正财运"/"七杀运"）
        "dayun_steps": [
            {"name": s.get("name", ""), "qi_sui": s.get("qi_sui"), "zhi_sui": s.get("zhi_sui"),
             "shi_shen": s.get("shi_shen", "")}
            for s in dy_steps
        ],
    }
