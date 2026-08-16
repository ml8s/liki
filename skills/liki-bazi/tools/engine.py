"""断语表查表器（纯函数——真值表匹配）。

用法：
    from engine import match
    from duanyu import evaluate_factors
    snapshot = evaluate_factors(factors, gender, chart)   # 因子快照
    hits = match(table, snapshot)                         # 命中条目（按表序）

匹配语义：
- 行内 AND：所有约束列的期望值 == 因子快照值 → 行成立
- 约束值：0/1、字符串枚举（相等比较——无 >=N/any_of/优先级）
- 行间：多行成立 = 多面性，返回全部（断语表按确定性排序，程序不排序）
"""
from __future__ import annotations


def _val_match(cond, actual):
    """单约束匹配：条件值（0/1、字符串枚举）与因子实际值相等即命中。"""
    return actual == cond
def match(table: list[dict], snapshot: dict, exclusive: bool = False) -> list[dict]:
    """查表：返回命中条目（按断语表顺序——命理师按确定性排序，程序不排序）。table = csv 加载的条目列表。
    exclusive=True 时返回表序最前命中（单一事实域——命理上状态互斥，不可多面）。"""
    hits = []
    for e in table:
        cons = dict(e.get("约束", {}) or {})
        if not all(_val_match(v, snapshot.get(k)) for k, v in cons.items()):
            continue
        hits.append(e)
    # 权重 = 断语表设计（条目顺序 + 约束内容）——程序不做权重定义/排序
    # 命理：断语表由命理师按确定性排序（铁断/具体断语在前），match 按表顺序返回
    if exclusive and hits:
        # 取表序最前的命中（单一事实域——婚姻状态互斥）
        return [hits[0]]
    return hits
