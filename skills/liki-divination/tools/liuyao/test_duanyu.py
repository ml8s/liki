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


def test_load_table_career():
    """测试加载事业断语表"""
    table = load_table('career')
    assert len(table) > 0
    assert 'id' in table[0]
    assert '结论' in table[0]


def test_load_table_wealth():
    """测试加载财运断语表"""
    table = load_table('wealth')
    assert len(table) > 0


def test_load_table_patterns():
    """测试加载特殊格局断语表"""
    table = load_table('patterns')
    assert len(table) > 0


def test_load_table_not_found():
    """测试加载不存在的断语表"""
    table = load_table('not_found')
    assert len(table) == 0


def test_query_career_wang():
    """测试查询事业断语（旺相+动爻生用）"""
    factors = {
        'yongshen_wangshuai': '旺',
        'dong_sheng': True,
    }
    results = query('career', factors)
    assert len(results) > 0
    assert any('升迁可期' in r.get('结论', '') for r in results)


def test_query_career_xiu():
    """测试查询事业断语（休囚+动爻克用）"""
    factors = {
        'yongshen_wangshuai': '休',
        'dong_ke': True,
    }
    results = query('career', factors)
    assert len(results) > 0
    assert any('难成' in r.get('结论', '') for r in results)


def test_query_wealth_wang():
    """测试查询财运断语（旺相+动爻生用）"""
    factors = {
        'yongshen_wangshuai': '旺',
        'dong_sheng': True,
    }
    results = query('wealth', factors)
    assert len(results) > 0


def test_query_patterns_xunkong():
    """测试查询旬空断语"""
    factors = {
        'yongshen_wangshuai': '旺',
        'yongshen_xunkong': True,
    }
    results = query('patterns', factors)
    assert len(results) > 0


def test_query_no_match():
    """测试无匹配断语"""
    factors = {
        'yongshen_wangshuai': '旺',
        'dong_sheng': True,
        'dong_ke': True,
    }
    results = query('career', factors)
    assert len(results) == 0


def test_format_output():
    """测试格式化输出"""
    results = [
        {
            '结论': '升迁可期',
            '依据': '官鬼旺相+动爻生用',
            '经典原文': '《增删卜易》',
            'yehu_tip': '野鹤提示',
            'pattern_interaction': '格局交互',
            'common_misjudge': '常见误判',
        }
    ]
    output = format_output(results)
    assert '升迁可期' in output
    assert '野鹤提示' in output


def test_format_output_empty():
    """测试格式化输出（空）"""
    output = format_output([])
    assert '无匹配断语' in output


def test_evaluate_factors():
    """测试因子计算"""
    chart = {
        'lines': [
            {'yue_po': False, 'xun_kong': False, 'dong_sheng': True, 'dong_ke': False},
            {'yue_po': False, 'xun_kong': False, 'dong_sheng': False, 'dong_ke': False},
            {'yue_po': False, 'xun_kong': False, 'dong_sheng': False, 'dong_ke': True},
        ],
        'wang_shuai': ['旺', '相', '休'],
        'patterns': [
            {'type': '旬空', 'sub_type': '假空'},
        ],
    }
    yong_shen = {'position': 1}
    factors = evaluate_factors(chart, yong_shen)
    assert factors['yongshen_wangshuai'] == '旺'
    assert factors['dong_sheng'] == True
    assert factors['pattern_旬空'] == '假空'
