"""六爻断语生成器测试"""
import os
import sys

# 添加 tools 目录到路径
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from duanyu import load_constants, load_table, evaluate_factors, query, format_output


def test_load_constants():
    """测试加载基础常量"""
    constants = load_constants()
    assert '五行生克' in constants
    assert '地支五行' in constants
    assert '六亲定义' in constants
    assert '六神定义' in constants
    assert '八宫归属' in constants


def test_load_table_enum_general():
    """测试加载综合断语表"""
    table = load_table('enum_general')
    assert len(table) > 0
    assert '用神旺衰' in table[0]
    assert '主要动爻关系' in table[0]
    assert '格局' in table[0]


def test_load_table_yingqi():
    """测试加载应期表"""
    table = load_table('yingqi')
    assert len(table) > 0
    assert '用神旺衰' in table[0]


def test_load_table_not_found():
    """测试加载不存在的断语表"""
    table = load_table('not_found')
    assert len(table) == 0


def test_format_output():
    """测试格式化输出"""
    results = [
        {
            '结论': '事可成',
            '依据': '用神旺相+动爻生用',
            '经典原文': '《增删卜易》',
        }
    ]
    output = format_output(results)
    assert '事可成' in output
    assert '用神旺相' in output


def test_format_output_empty():
    """测试格式化输出（空）"""
    output = format_output([])
    assert '无匹配断语' in output


# ===== 综合断语表（enum_general）=====

def test_query_enum_wang_shengyong():
    """旺+生用 → 事可成"""
    factors = {'用神旺衰': '旺', '月破': False, '旬空': False, '主要动爻关系': '生用', '格局': ''}
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('事可成' in r.get('结论', '') for r in results)


def test_query_enum_xiu_keyong():
    """休+克用 → 事难成"""
    factors = {'用神旺衰': '休', '月破': False, '旬空': False, '主要动爻关系': '克用', '格局': ''}
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('事难成' in r.get('结论', '') for r in results)


def test_query_enum_yuepo():
    """旺+月破+生用 → 先挫后成"""
    factors = {'用神旺衰': '旺', '月破': True, '旬空': False, '主要动爻关系': '生用', '格局': ''}
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('先挫后成' in r.get('结论', '') for r in results)


def test_query_enum_xunkong():
    """休+旬空+生用 → 事不实"""
    factors = {'用神旺衰': '休', '月破': False, '旬空': True, '主要动爻关系': '生用', '格局': ''}
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('事不实' in r.get('结论', '') for r in results)


def test_query_enum_geju():
    """旺+生用+六冲 → 反复"""
    factors = {'用神旺衰': '旺', '月破': False, '旬空': False, '主要动爻关系': '生用', '格局': '六冲'}
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('反复' in r.get('结论', '') for r in results)


def test_query_enum_shengyuan():
    """旺+生原神 → 事可成"""
    factors = {'用神旺衰': '旺', '月破': False, '旬空': False, '主要动爻关系': '生原神', '格局': ''}
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('事可成' in r.get('结论', '') for r in results)


def test_query_enum_wudongyao():
    """旺+无动爻 → 顺势"""
    factors = {'用神旺衰': '旺', '月破': False, '旬空': False, '主要动爻关系': '无动爻', '格局': ''}
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('顺势' in r.get('结论', '') for r in results)


def test_query_no_match():
    """不存在的表 → 无匹配"""
    results = query('not_exist_table', {'用神旺衰': '旺'})
    assert len(results) == 0


# ===== 应期表（yingqi）=====

def test_query_yingqi_wang():
    """旺 → 逢值或逢合"""
    factors = {'用神旺衰': '旺', '月破': False, '旬空': False}
    results = query('yingqi', factors)
    assert len(results) > 0
    assert any('逢值或逢合' in r.get('结论', '') for r in results)


def test_query_yingqi_xiu():
    """休 → 待旺相"""
    factors = {'用神旺衰': '休', '月破': False, '旬空': False}
    results = query('yingqi', factors)
    assert len(results) > 0
    assert any('待旺相' in r.get('结论', '') for r in results)


def test_query_yingqi_yuepo():
    """旺+月破 → 逢合补破"""
    factors = {'用神旺衰': '旺', '月破': True, '旬空': False}
    results = query('yingqi', factors)
    assert len(results) > 0
    assert any('逢合补破' in r.get('结论', '') for r in results)


def test_query_yingqi_xunkong():
    """旺+旬空 → 出空填实"""
    factors = {'用神旺衰': '旺', '月破': False, '旬空': True}
    results = query('yingqi', factors)
    assert len(results) > 0
    assert any('出空' in r.get('结论', '') for r in results)


# ===== 因子提取 =====

def test_evaluate_factors():
    """测试因子计算"""
    chart = {
        'dong_yao_relations': [
            {'position': 1, 'relation': '生用'},
        ],
        'patterns': [
            {'type': '六冲', 'sub_type': ''},
        ],
    }
    yong_shen = {
        'position': 1,
        'wang_shuai': '旺',
        'yue_po': False,
        'xun_kong': False,
        'mu_ku': False,
    }
    factors = evaluate_factors(chart, yong_shen)
    assert factors['yongshen_wangshuai'] == '旺'
    assert factors['main_dongyao_relation'] == '生用'
    assert factors['pattern'] == '六冲'


def test_evaluate_factors_legacy():
    """测试因子计算（旧格式兼容：用神无聚合字段时从 line 取）"""
    chart = {
        'lines': [
            {'yue_po': False, 'xun_kong': False, 'mu_ku': False},
        ],
        'wang_shuai': ['旺'],
        'dong_yao_relations': [],
        'patterns': [],
    }
    yong_shen = {'position': 1}
    factors = evaluate_factors(chart, yong_shen)
    assert factors['yongshen_wangshuai'] == '旺'
    assert factors['main_dongyao_relation'] == '无动爻'


def test_evaluate_factors_multiple_relations():
    """测试多动爻关系：取第一个为主要关系"""
    chart = {
        'dong_yao_relations': [
            {'position': 1, 'relation': '生用'},
            {'position': 3, 'relation': '克用'},
        ],
        'patterns': [],
    }
    yong_shen = {'position': 1, 'wang_shuai': '旺'}
    factors = evaluate_factors(chart, yong_shen)
    assert factors['main_dongyao_relation'] == '生用'
