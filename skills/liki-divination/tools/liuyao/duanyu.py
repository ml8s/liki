"""六爻断语生成器（表驱动，零命理逻辑）。

分层：constants.json（基础常量）→ factors.csv（因子定义）→ csv/*.csv（断语表）。
输入：引擎返回的因子组合
输出：匹配的断语列表
"""
import csv
import json
import os
from typing import Any, Dict, List

_TABLE_CACHE: Dict[str, List[Dict[str, Any]]] = {}


def load_constants() -> Dict[str, Any]:
    """加载基础常量"""
    path = os.path.join(os.path.dirname(__file__), 'constants.json')
    with open(path, encoding='utf-8') as f:
        return json.load(f)


def load_table(name: str) -> List[Dict[str, Any]]:
    """加载 CSV 断语表"""
    if name in _TABLE_CACHE:
        return _TABLE_CACHE[name]
    fname = name if name.endswith('.csv') else name + '.csv'
    path = os.path.join(os.path.dirname(__file__), 'csv', fname)
    if not os.path.exists(path):
        return []
    with open(path, encoding='utf-8') as f:
        reader = csv.DictReader(f)
        _TABLE_CACHE[name] = list(reader)
    return _TABLE_CACHE[name]


def evaluate_factors(chart: Dict[str, Any], yong_shen: Dict[str, Any]) -> Dict[str, Any]:
    """计算因子组合（从引擎输出提取）"""
    factors: Dict[str, Any] = {}
    # 用神聚合字段（引擎已聚合）；旧格式兼容：从 line 取
    factors['yongshen_wangshuai'] = yong_shen.get('wang_shuai', '')
    factors['yongshen_yuepo'] = yong_shen.get('yue_po', False)
    factors['yongshen_xunkong'] = yong_shen.get('xun_kong', False)
    factors['yongshen_muku'] = yong_shen.get('mu_ku', False)
    if not factors['yongshen_wangshuai']:
        yong_pos = yong_shen.get('position', 0)
        if yong_pos > 0:
            lines = chart.get('lines', [])
            wang_shuai = chart.get('wang_shuai', [])
            if yong_pos <= len(lines) and yong_pos <= len(wang_shuai):
                line = lines[yong_pos - 1]
                factors['yongshen_wangshuai'] = wang_shuai[yong_pos - 1]
                factors['yongshen_yuepo'] = line.get('yue_po', False)
                factors['yongshen_xunkong'] = line.get('xun_kong', False)
                factors['yongshen_muku'] = line.get('mu_ku', False)
    # 动爻关系（枚举集合，引擎已计算）
    relations = chart.get('dong_yao_relations', [])
    rel_list = [r.get('relation', '') for r in relations if r.get('relation')]
    factors['dong_yao_relations'] = rel_list
    # 主要动爻关系（枚举）：多动爻时取第一个（生克力量最直接者，命理上取关键一动）
    factors['main_dongyao_relation'] = rel_list[0] if rel_list else '无动爻'
    # 特殊格局因子
    patterns = chart.get('patterns', [])
    for p in patterns:
        factors[f'pattern_{p["type"]}'] = p.get('sub_type', '')
    # 格局枚举（合并 pattern_*：取第一个独立格局类型；空=无）
    pattern_types = [p.get('type', '') for p in patterns if p.get('type')]
    factors['pattern'] = pattern_types[0] if pattern_types else ''
    return factors


def query(category: str, factors: Dict[str, Any]) -> List[Dict[str, Any]]:
    """查询断语（支持枚举断语表）

    语义（命理逻辑在表）：
    - factors 中显式传入非空/非 False 值的列 → 必须精确匹配（行该列非空且相等）
    - factors 中为空的列（未传该维度）→ 不关心（行任何值都行）
    - 布尔 False 视为"未指定该维度"（不参与匹配，即"无修饰"也匹配）
    """
    table = load_table(category)
    results = []
    for row in table:
        match = True
        for key, value in factors.items():
            if key not in row:
                continue
            if value is None or value == '':
                continue  # 未传该维度 → 不关心
            row_value = row[key]
            if isinstance(value, bool):
                # 布尔 False 视为未指定（无修饰也匹配）；True 必须行=1
                if value is True and row_value != '1':
                    match = False
                    break
            elif isinstance(value, (list, tuple)):
                if value and row_value not in [str(v) for v in value]:
                    match = False
                    break
            else:
                if str(value) != row_value:
                    match = False
                    break
        if match:
            results.append(row)
    return results


def format_output(results: List[Dict[str, Any]]) -> str:
    """格式化输出"""
    if not results:
        return "无匹配断语"
    output = []
    for r in results:
        output.append(f"结论：{r.get('结论', '')}")
        output.append(f"依据：{r.get('依据', '')}")
        output.append(f"经典原文：{r.get('经典原文', '')}")
        if r.get('yehu_tip'):
            output.append(f"野鹤提示：{r.get('yehu_tip')}")
        if r.get('pattern_interaction'):
            output.append(f"格局交互：{r.get('pattern_interaction')}")
        if r.get('common_misjudge'):
            output.append(f"常见误判：{r.get('common_misjudge')}")
    return '\n'.join(output)
