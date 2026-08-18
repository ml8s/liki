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
    """测试无匹配断语（枚举表）"""
    factors = {
        '用神旺衰': '死',
        '月破': False,
        '旬空': False,
        '主要动爻关系': '克忌神',
        '格局': '',
    }
    results = query('enum_general', factors)
    # enum_general 无"死+克忌神"组合 → 无匹配
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
        'dong_yao_relations': [
            {'position': 1, 'relation': '生用'},
        ],
        'patterns': [
            {'type': '旬空', 'sub_type': '假空'},
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
    assert factors['dong_yao_relations'] == ['生用']
    assert factors['pattern_旬空'] == '假空'


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


def test_query_wealth_boji():
    """测试查询博戏断语"""
    factors = {
        'yongshen_wangshuai': '旺',
        'dong_sheng': True,
    }
    results = query('wealth', factors)
    assert len(results) > 0
    boji_results = [r for r in results if r.get('subcategory') == '']
    assert len(boji_results) > 0


def test_query_academic_keju():
    """测试查询科举断语"""
    factors = {
        'yongshen_wangshuai': '旺',
        'dong_sheng': True,
    }
    results = query('academic', factors)
    assert len(results) > 0


def test_query_health_shuangtai():
    """测试查询疾病双时态断语"""
    factors = {
        'yongshen_wangshuai': '旺',
        'dong_sheng': True,
    }
    results = query('health', factors)
    assert len(results) > 0


def test_query_general_nianyun():
    """测试查询年运断语"""
    factors = {
        'yongshen_wangshuai': '旺',
        'dong_sheng': True,
    }
    results = query('general', factors)
    assert len(results) > 0


def test_load_enum_general():
    """测试加载枚举断语表"""
    table = load_table('enum_general')
    assert len(table) > 0
    assert '用神旺衰' in table[0]


def test_query_enum_wang_shengyong():
    """测试枚举查询：旺+生用 → 事可成"""
    factors = {
        '用神旺衰': '旺',
        '月破': False,
        '旬空': False,
        '主要动爻关系': '生用',
        '格局': '',
    }
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('事可成' in r.get('结论', '') for r in results)


def test_query_enum_xiu_keyong():
    """测试枚举查询：休+克用 → 事难成"""
    factors = {
        '用神旺衰': '休',
        '月破': False,
        '旬空': False,
        '主要动爻关系': '克用',
        '格局': '',
    }
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('事难成' in r.get('结论', '') for r in results)


def test_query_enum_yuepo():
    """测试枚举查询：旺+月破+生用 → 先挫后成"""
    factors = {
        '用神旺衰': '旺',
        '月破': True,
        '旬空': False,
        '主要动爻关系': '生用',
        '格局': '',
    }
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('先挫后成' in r.get('结论', '') for r in results)


def test_query_enum_geju():
    """测试枚举查询：旺+生用+六冲 → 反复"""
    factors = {
        '用神旺衰': '旺',
        '月破': False,
        '旬空': False,
        '主要动爻关系': '生用',
        '格局': '六冲',
    }
    results = query('enum_general', factors)
    assert len(results) > 0
    assert any('反复' in r.get('结论', '') for r in results)


def test_evaluate_factors_main_relation():
    """测试主要动爻关系提取"""
    chart = {
        'dong_yao_relations': [
            {'position': 1, 'relation': '生用'},
            {'position': 3, 'relation': '克用'},
        ],
        'patterns': [
            {'type': '六冲', 'sub_type': ''},
        ],
    }
    yong_shen = {'position': 1, 'wang_shuai': '旺'}
    factors = evaluate_factors(chart, yong_shen)
    assert factors['main_dongyao_relation'] == '生用'
    assert factors['pattern'] == '六冲'
