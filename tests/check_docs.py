"""文档契约检查（外部评审 #15/#23/#26/#27/#28 类防回潮——文档引用可解析性）。

校验（liki-bazi skill 内文档：SKILL.md + app/*.md + domains/**/*.md）：
1. 断语 id 引用（如 `ymar_101`/xm_m06）必须存在于断语表 CSV（防引用已删除/改名的断语）
2. 文件路径引用（`tools/skill-tools.json`、`app/marriage.md`…）必须真实存在（防引用已移动/删除的文档）
3. RPC 方法名引用（`bazi.chart`、`ziwei.daxian`…）必须在方法白名单（引擎实际方法集，防拼错/臆造方法）

不做（误报率高、跨仓库）：
- 引擎返回字段名校验（字段字典在 liki-engine schema——由引擎侧测试负责）
- 自然语言/多行反引号内容（模板占位）——只校验单行标识符形态的引用

用法：python3 tests/check_docs.py [skill_dir]
"""
from __future__ import annotations
import csv
import glob
import os
import re
import sys

_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SKILL = sys.argv[1] if len(sys.argv) > 1 else os.path.join(_ROOT, "skills", "liki-bazi")

# 引擎实际方法集（与 liki-engine internal/agent 注册方法一致——新增/删除方法须同步本表）
METHOD_WHITELIST = {
    "rpc.discover",
    # bazi
    "bazi.chart", "bazi.fullchart", "bazi.liunian", "bazi.liuyue", "bazi.liuri",
    "bazi.liushi", "bazi.xiaoyun", "bazi.bond",
    # ziwei
    "ziwei.chart", "ziwei.fullchart", "ziwei.liunian", "ziwei.liuyue", "ziwei.liuri",
    "ziwei.liushi", "ziwei.daxian", "ziwei.bond",
    # 子流程三 skill（文档分流提示引用）
    "liuyao.qigua", "liuyao.chart",
    "qimen.chart",
    "huangli.days",
    "bazhai.chart", "bazhai.layout",
    "xuankong.chart", "xuankong.liunian",
    "qiming.pick", "qiming.build", "qiming.check", "qiming.char",
    # divination
    "liuyao.qigua", "liuyao.chart", "qimen.chart",
    # 基础
    "time.now", "city.coords", "tianwen.time",
}
# 点分 token 允许的方法前缀（过滤自然语言/文件名的误报）
_METHOD_PREFIXES = tuple(sorted({m.split(".")[0] for m in METHOD_WHITELIST}))
# 手调方法之外、rpc.discover 返回字段的引用（result.methods/params.properties 等）
_SKIP_DOTTED = {"params.properties", "result.methods", "result.info", "result.info.version"}


def load_duanyu_ids() -> set:
    """全部断语表 id（bazi/ + ziwei/ CSV）。"""
    ids = set()
    for f in glob.glob(os.path.join(SKILL, "tools", "bazi", "*.csv")) + \
            glob.glob(os.path.join(SKILL, "tools", "ziwei", "*.csv")):
        for r in csv.DictReader(open(f, encoding="utf-8")):
            if (r.get("id") or "").strip():
                ids.add(r["id"].strip())
    return ids


def main() -> int:
    ids = load_duanyu_ids()
    prefixes = sorted({i.split("_")[0] + "_" for i in ids if "_" in i})
    # 无断语表（子流程 skill：domains 表为 LLM 翻译表）→ 空正则退化为匹配任意数字——跳过 id 检查
    prefix_re = re.compile(r"\b(" + "|".join(re.escape(p) for p in prefixes) + r")(\d+[a-z]*|x+)\b") if prefixes else None
    path_re = re.compile(r"`((?:tools|app|domains|webapp)/[^`]+)`|((?:tools|app|domains|webapp)/[\w./\-]+\.(?:md|py|json|csv|sh))")
    method_re = re.compile(r"\b(" + "|".join(_METHOD_PREFIXES) + r")\.[a-z_]+\b")

    docs = [os.path.join(SKILL, "SKILL.md")]
    docs += sorted(glob.glob(os.path.join(SKILL, "app", "*.md")))
    docs += sorted(glob.glob(os.path.join(SKILL, "domains", "**", "*.md"), recursive=True))
    docs = [d for d in docs if os.path.exists(d)]
    # tools/*.py 中的 RPC 调用方法名（call("bazi.chart", ...)）——必须在引擎方法白名单
    # （防引擎改名/拼错后 skill 侧静默失败——跨仓库契约，2026-08 自查固化）
    py_docs = sorted(glob.glob(os.path.join(SKILL, "tools", "*.py")))

    errors, warnings = [], []
    for doc in docs:
        rel = os.path.relpath(doc, _ROOT)
        txt = open(doc, encoding="utf-8").read()
        # 1) 断语 id（仅断语表类 skill 校验）
        if prefix_re is not None:
            for m in prefix_re.finditer(txt):
                tok = m.group(0)
                if "x" in m.group(2):
                    continue  # 范围写法（xg_2xx 表示 xg_2xx 系列）——放行
                if tok not in ids:
                    errors.append(f"[{rel}] 引用断语 id '{tok}' 不存在于断语表")
        # 2) 文件路径
        for m in path_re.finditer(txt):
            p = m.group(1) or m.group(2)
            if not p:
                continue
            if "*" in p or "xxx" in p or "<" in p:
                continue  # glob 模式（tools/bazi/*.csv）或占位符（app/xxx.md、domains/<域>/）——放行
            if not os.path.exists(os.path.join(SKILL, p)):
                errors.append(f"[{rel}] 引用文件不存在: {p}")
        # 3) 方法名
        for m in method_re.finditer(txt):
            meth = m.group(0)
            if meth in _SKIP_DOTTED:
                continue
            if meth not in METHOD_WHITELIST:
                warnings.append(f"[{rel}] 引用的方法 '{meth}' 不在方法白名单（引擎方法集）——核对拼写/是否新增")
    # 4) tools/*.py 的 RPC 调用方法名（call("X", ...)）——契约校验
    _rpc_call_re = re.compile(r'\bcall\(\s*["\']([a-z]+\.[a-z_]+)["\']')
    for f in py_docs:
        rel = os.path.relpath(f, _ROOT)
        for m in _rpc_call_re.finditer(open(f, encoding="utf-8").read()):
            if m.group(1) not in METHOD_WHITELIST:
                errors.append(f"[{rel}] RPC 调用方法 '{m.group(1)}' 不在引擎方法白名单——skill 侧将静默失败")
    if not docs:
        warnings.append(f"SKILL 目录未找到文档（{SKILL}）")
    # 5) README 断语统计 vs 实际（仅主 skill——README 统计的是 liki-bazi 断语）
    _readme = os.path.join(_ROOT, "README.md")
    if os.path.exists(_readme) and os.path.abspath(SKILL) == os.path.abspath(os.path.join(_ROOT, "skills", "liki-bazi")):
        _actual = 0
        for f in glob.glob(os.path.join(SKILL, "tools", "**", "*.csv"), recursive=True):
            if os.path.basename(f) in ("factors.csv", "factors_liunian.csv", "factors_narrow.csv"):
                continue
            _actual += sum(1 for r in csv.DictReader(open(f, encoding="utf-8")) if r.get("id"))
        _m = re.search(r'共 \*\*(\d+) 条断语\*\*', open(_readme, encoding="utf-8").read())
        if _m and int(_m.group(1)) != _actual:
            errors.append(f"[README] 断语统计 {_m.group(1)} ≠ 实际 {_actual}——补/删断语后未更新（make build-archive 不覆盖，需手动）")

    print(f"扫描文档 {len(docs)} 个（断语 id 全集 {len(ids)}）")
    print(f"错误: {len(errors)} 个")
    for e in errors:
        print("  ✗", e)
    print(f"警告: {len(warnings)} 个")
    for w in warnings[:30]:
        print("  ⚠", w)
    if len(warnings) > 30:
        print(f"  ... 共 {len(warnings)} 条警告")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
