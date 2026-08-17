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
    # 基础因子
    yong_pos = yong_shen.get('position', 0)
    if yong_pos > 0:
        lines = chart.get('lines', [])
        wang_shuai = chart.get('wang_shuai', [])
        if yong_pos <= len(lines) and yong_pos <= len(wang_shuai):
            line = lines[yong_pos - 1]
            factors['yongshen_wangshuai'] = wang_shuai[yong_pos - 1]
            factors['yongshen_yuepo'] = line.get('yue_po', False)
            factors['yongshen_xunkong'] = line.get('xun_kong', False)
            factors['dong_sheng'] = line.get('dong_sheng', False)
            factors['dong_ke'] = line.get('dong_ke', False)
    # 特殊格局因子
    patterns = chart.get('patterns', [])
    for p in patterns:
        factors[f'pattern_{p["type"]}'] = p.get('sub_type', '')
    return factors


def query(category: str, factors: Dict[str, Any]) -> List[Dict[str, Any]]:
    """查询断语"""
    table = load_table(category)
    results = []
    for row in table:
        match = True
        for key, value in factors.items():
            if key in row and row[key] != '':
                # 处理布尔值与字符串 '1'/'0' 的转换
                row_value = row[key]
                if isinstance(value, bool):
                    row_value = row_value == '1'
                elif isinstance(value, str):
                    row_value = str(row_value)
                if str(value) != str(row_value):
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
