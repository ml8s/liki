"""断语库 schema 校验（数据库约束：断语引用的因子必须存在）。

校验内容：
1. 断语表约束键（含 any_of 内）必须 ∈ 因子全集（factors.csv 因子名 + 引擎直读字段）
2. 断语约束键必须 ∈ 该表"引用因子"声明（若表有声明）——防声明与实际脱节
3. 引用因子声明中的因子名必须存在
4. 各断语表条目 id 唯一
5. 死列（表头列无任何行引用）——error（2026-08 存量 227 列已清零，此后新增即拦截）

用法：
    python3 tests/check_schema.py        # 校验全部，退出码 0/1
"""
from __future__ import annotations
import glob
import os
import sys
import json

_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DY = os.path.join(_ROOT, "skills", "liki-bazi", "tools")   # skill 内容在 skills/liki-bazi/（工程根=仓库根）

# 排盘上下文不是因子，但断语真值表可消费。
CONTEXT_KEYS = {"性别"}
# 应期层因子（factors_liunian.csv 定义——yingqi 域用，不在 factors.csv）
def load_liunian_names() -> set:
    """流年因子名——从 factors_liunian.csv（真值表单一权威——json 已删）。"""
    import csv as _csv
    return {r["因子"] for r in _csv.DictReader(open(os.path.join(DY, "factors", "factors_liunian.csv"), encoding="utf-8"))}


LIUNIAN_KEYS = load_liunian_names()


def load_liunian_reachability() -> tuple:
    """流年因子可达性（外部评审 #17/#18：跨术数死条件 / 域错配死列防回潮）。

    返回 (bazi_reachable, ziwei_factors)：
    - bazi_reachable: 八字流年快照可达键（factors_liunian.csv 表头算子列 ∪ 术数=bazi 因子名）
    - ziwei_factors: 紫微流年因子名（术数=ziwei）——八字流年表引用 = 跨术数死条件
    """
    import csv as _csv
    import os as _os
    path = _os.path.join(DY, "factors", "factors_liunian.csv")
    bz, zw, cols = set(), set(), set()
    with open(path, encoding="utf-8") as fh:
        rd = _csv.DictReader(fh)
        cols = {c for c in rd.fieldnames if c not in ("因子", "术数", "原语直通", "依据")}
        rows = list(rd)
    for r in rows:
        if (r.get("术数") or "bazi").strip() == "ziwei":
            zw.add(r["因子"])
        else:
            bz.add(r["因子"])
    bz |= cols  # 算子列名（流年宫化[X]/三刑[X]…）同为可达键——同名约定，因子行术数定归属
    return bz, zw


def load_factor_shushi() -> dict:
    """因子名 → 术数（bazi/ziwei）——交叉校验：bazi 表只用八字因子、ziwei 表只用紫微因子。"""
    import csv as _csv
    mapping = {}
    for r in _csv.DictReader(open(os.path.join(DY, "factors", "factors.csv"), encoding="utf-8")):
        mapping[r["因子"]] = (r.get("术数") or "bazi").strip()
    return mapping


def load_factors_names() -> set:
    """因子名清单——从 factors.csv（真值表单一权威——json 已删）。"""
    import csv as _csv
    return {r["因子"] for r in _csv.DictReader(open(os.path.join(DY, "factors", "factors.csv"), encoding="utf-8"))}


def main() -> int:
    constants = json.load(open(os.path.join(DY, "constants.json"), encoding="utf-8"))
    factor_names = load_factors_names()
    factor_shushi = load_factor_shushi()
    bz_reach, zw_factors = load_liunian_reachability()
    import csv as _csv2
    _LIUNIAN_ALL = list(_csv2.DictReader(open(os.path.join(DY, "factors", "factors_liunian.csv"), encoding="utf-8")))
    # 因子 → 是否字符串直通（直读[..,任意]）——强化⑩ 校验断语字符串约束列值域
    _STR_ZHITONG = {}
    for _filename in ("factors.csv", "factors_liunian.csv"):
        for _r in _csv2.DictReader(open(os.path.join(DY, "factors", _filename), encoding="utf-8")):
            _zt = (_r.get("原语直通") or "").strip()
            _STR_ZHITONG.setdefault(_r["因子"], set()).add(("任意" in _zt) if _zt else False)
    errors = []
    warnings = []
    seen_ids = {}
    # 强化⑥：引用本命[X] 的 X 必须是本命因子名。算子按本命快照通用读取。
    import csv as _csv
    for r in _csv.DictReader(open(os.path.join(DY, "factors", "factors_liunian.csv"), encoding="utf-8")):
        for c, v in r.items():
            if c.startswith("引用本命[") and (v or "").strip():
                inner = c[len("引用本命["):-1]
                if inner not in factor_names:
                    errors.append(f"[factors_liunian] 引用本命[{inner}] 不是本命因子——恒 0 死条件")
    # 因子定义必须有直通或条件；空定义会恒假且绕过多数表结构检查。
    for _filename in ("factors.csv", "factors_liunian.csv"):
        with open(os.path.join(DY, "factors", _filename), encoding="utf-8") as _fh:
            for _r in _csv.DictReader(_fh):
                _has_direct = bool((_r.get("原语直通") or "").strip())
                _has_conds = any(
                    (_v or "").strip() for _k, _v in _r.items()
                    if _k not in ("因子", "术数", "原语直通", "依据")
                )
                if not _has_direct and not _has_conds:
                    errors.append(f"[{_filename}] 因子 {_r['因子']} 空定义（恒 0）")
    files = glob.glob(os.path.join(DY, "**", "*.csv"), recursive=True)
    for f in sorted(files):
        if os.path.basename(f) in ("factors.csv", "factors_liunian.csv"):
            continue
        import csv as _csv
        dom = os.path.basename(f)[:-4]
        _rel = os.path.relpath(f, DY).replace(os.sep, "/")
        rows = []
        _hdr = None
        with open(f, encoding="utf-8") as fh:
            _rd = _csv.DictReader(fh)
            _hdr = [c for c in _rd.fieldnames if c not in ("id", "事件", "结论", "依据", "经典原文")]
            for r in _rd:
                # 参差行防御：行字段数 ≠ 表头（短缺 → 值补 None；多余 → 进 restkey=None）。
                # 在解析口明确报错并跳过该行，而非下游 item["结论"].strip() 裸 AttributeError。
                if None in r or any(v is None for v in r.values()):
                    errors.append(f"[{_rel}] 第 {_rd.line_num} 行列数与表头不一致（参差 CSV）——请对齐列数后重查")
                    continue
                cons = {}
                for k, v in r.items():
                    if k in ("id", "事件", "结论", "依据", "经典原文"):
                        continue
                    if (v or "").strip():
                        cons[k] = v
                if not cons:
                    errors.append(f"[{_rel}/{r.get('id')}] 断语无约束（恒命中）")
                # 交叉校验：八字表只用八字因子、紫微表只用紫微因子（防混合回潮——真分开）
                # 表文件在 bazi/ziwei 子目录（load_table 按目录定位），expect 按目录判定——
                # 文件名无 bazi_/ziwei_ 前缀，不能用 dom（basename）判断（历史盲区：expect 恒 None）
                expect = "bazi" if _rel.startswith("bazi/") else ("ziwei" if _rel.startswith("ziwei/") else None)
                if expect:
                    for ck in cons:
                        cs = factor_shushi.get(ck)
                        if cs and cs not in (expect, "common"):
                            warnings.append(f"[{_rel}] 跨术数条件列 '{ck}'（{cs}）——{expect} 表应纯{expect}因子")
                rows.append({"id": r.get("id", ""), "约束": cons, "结论": r.get("结论", ""),
                    "依据": r.get("依据", ""), "经典原文": r.get("经典原文", "")})
        used_keys = set()
        for item in rows:
            eid = item.get("id")
            if eid:
                if eid in seen_ids:
                    errors.append(f"断语 id 重复: {eid}（{seen_ids[eid]} 与 {f}）")
                seen_ids[eid] = f
            for k in (item.get("约束") or {}):
                used_keys.add(k)
        # 约束键必须存在
        for k in used_keys:
            if k not in factor_names and k not in CONTEXT_KEYS and k not in LIUNIAN_KEYS:
                errors.append(f"[{dom}] 约束键 '{k}' 不在因子全集（factors.csv 无此因子）")
        # 强化④（外部评审 #17/#18）：流年域表（yearly_*/yingqi）键可达性——
        # 八字流年表引用紫微流年因子 = 跨术数死条件；引用本命因子（非引用本命[X]）= 流年快照不可达死列
        if _rel.startswith("bazi/") and (dom.startswith("yearly_") or dom == "yingqi"):
            for k in used_keys:
                if k in zw_factors:
                    errors.append(f"[{_rel}] 八字流年表引用紫微流年因子 '{k}'——跨术数死条件（八字流年快照恒无此键，永不命中）")
                elif k not in bz_reach and k not in CONTEXT_KEYS and k in factor_names:
                    errors.append(f"[{_rel}] 流年表引用本命因子 '{k}'（非「引用本命[X]」形式）——流年快照不可达，死列")
        # 强化⑤（外部评审 #20）：死列检测——表头列无任何行引用（纯冗余，易滋生死条件）。
        # 2026-08 起升级为 error：存量 227 列已清零（生成器时代的统一超集表头遗物），
        # 基线为 0 后新增死列 = 回潮，直接拦（与 #17/#18 同级别的硬约束）。
        _unused = [c for c in _hdr if c not in used_keys]
        if _unused:
            _s = ",".join(_unused[:8]) + ("…" if len(_unused) > 8 else "")
            errors.append(f"[{dom}] 死列 {len(_unused)} 个（表头列无任何行引用——删除该列，运行时无效纯冗余）: {_s}")
        # 强化⑦（自查 2026-08：重复断语行——同约束多条近似结论 → agent 输出冗余/矛盾）
        _seen_cons = {}
        for item in rows:
            _cons = tuple(sorted((k, v) for k, v in (item.get("约束") or {}).items()))
            if _cons and _cons in _seen_cons:
                warnings.append(f"[{dom}] 重复约束行: {_seen_cons[_cons]} 与 {item['id']} 约束完全相同（结论近似=冗余，应合并）")
            else:
                _seen_cons[_cons] = item['id']
        # 强化⑧（自查 2026-08：约束值域——非枚举列出现 0/1 以外取值 = 静默永不匹配）
        _ENUM_COLS = {"月令格", "扶抑从格", "日主五行", "日主", "日主长生状态", "性别", "十神",
                              "身强弱", "调候季节", "日支神煞类型", "月令本气十神", "大运十神类",
                              "流年日主长生状态"}
        for item in rows:
            for _k, _v in (item.get("约束") or {}).items():
                if _k in _ENUM_COLS:
                    continue
                if _v not in ("0", "1"):
                    errors.append(f"[{dom}/{item['id']}] 约束列 {_k} 取值 {_v!r} 非 0/1（枚举列才可用字符串）")
        # 标量列值域必须来自 constants 闭集，防止不可达字符串条件。
        _ENUM_SOURCES = {
            "月令格": "月令格局",
            "扶抑从格": "扶抑从格",
            "身强弱": "身强弱状态",
            "调候季节": "调候季节",
            "日主": "天干",
            "日主五行": "五行",
            "日主长生状态": "十二长生",
            "日支神煞类型": "日支神煞",
            "月令本气十神": "十神",
            "大运十神类": "十神大类",
            "流年日主长生状态": "十二长生",
        }
        for item in rows:
            for _k, _v in (item.get("约束") or {}).items():
                _source = _ENUM_SOURCES.get(_k)
                if _source and _v not in constants[_source]:
                    errors.append(
                        f"[{dom}/{item['id']}] 标量列 {_k}={_v!r} 不在 constants.{_source} 闭集"
                    )
        # 强化⑩（自查 2026-08：断语字符串约束列 vs 因子值域——月令格=XX格 需因子为字符串直通，
        # 否则因子返回 0/1 与字符串约束永不匹配 → 断语全灭）
        for item in rows:
            for _k, _v in (item.get("约束") or {}).items():
                if _v in ("0", "1") or _k in CONTEXT_KEYS:
                    continue
                _defs = _STR_ZHITONG.get(_k)
                if _defs is None:
                    errors.append(f"[{dom}/{item['id']}] 字符串约束列 {_k}={_v!r} 不是因子名（或引擎直读键）")
                elif not any(_d for _d in _defs):
                    errors.append(f"[{dom}/{item['id']}] 列 {_k}={_v!r} 因子非字符串直通（值域错配→永不匹配）")
        # 强化⑨（自查 2026-08：紫微流年表引用八字流年因子 = 跨术数死条件——与八字侧对称）
        if _rel.startswith("ziwei/") and dom.startswith("yearly_"):
            _bazi_liu_factors = {r["因子"] for r in _LIUNIAN_ALL if (r.get("术数") or "bazi").strip() == "bazi"}
            for k in used_keys:
                if k in _bazi_liu_factors:
                    errors.append(f"[{_rel}] 紫微流年表引用八字流年因子 '{k}'——跨术数死条件（紫微流年快照恒无此键）")
        # 强化①：结论=评测状态标签（3.6.0 去异化——结论须命理表达——防标签回潮）
        _LABELS = {"已婚", "未婚", "独身", "离异", "夫早亡", "已婚波折", "博士", "硕士", "大学",
                   "专科", "中学", "小学", "主妇", "老板", "老板+管理层", "管理层/高管", "稳定职业",
                   "打工有积蓄", "普通打工", "富贵", "小康", "普通", "贫穷", "婚姻复杂"}
        for item in rows:
            if item["结论"].strip() in _LABELS:
                warnings.append(f"[{dom}/{item['id']}] 结论为评测状态标签（{item['结论']}）——应为命理表达（标签在测试层映射）")
        # 强化②：必填列（结论/依据/经典原文非空）
        for item in rows:
            for col in ("结论", "依据", "经典原文"):
                if not (item.get(col) or "").strip():
                    warnings.append(f"[{dom}/{item['id']}] 必填列 '{col}' 为空")
        # 强化③：经典原文覆盖率
        _total = len(rows)
        _filled = sum(1 for it in rows if (it.get("经典原文") or "").strip())
        if _total and _filled < _total:
            warnings.append(f"[{dom}] 经典原文覆盖 {_filled}/{_total}（应为全量——缺 {_total - _filled} 条）")

    print(f"因子全集: {len(factor_names)} 个")
    print(f"错误: {len(errors)} 个")
    for e in errors:
        print("  ✗", e)
    print(f"警告: {len(warnings)} 个")
    for w in warnings[:40]:
        print("  ⚠", w)
    if len(warnings) > 40:
        print(f"  ... 共 {len(warnings)} 条警告")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
