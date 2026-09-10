"""模块边界：本命事实投影、pan 契约、表加载层不得反向依赖断语层。"""
from pathlib import Path

import _helpers  # noqa: F401
import natal_projection
import assertion_store
import errors
import factor_tables
import operators_liunian
import operators_natal
import pan_schema
import yearly_eval

ROOT = Path(natal_projection.__file__).parent


def _source(module):
    return Path(module.__file__).read_text(encoding="utf-8")


def test_natal_projection_has_no_factor_or_duanyu_dependency():
    src = _source(natal_projection)
    assert "import factors" not in src and "from factors" not in src
    assert "import duanyu" not in src and "from duanyu" not in src


def test_natal_projection_name_does_not_shadow_divination_snapshot():
    assert not (ROOT / "domain_snapshot.py").exists()
    assert not (ROOT / "domain_snapshot_contract.json").exists()
    assert (ROOT / "natal_projection.py").is_file()
    assert (ROOT / "natal_projection_contract.json").is_file()


def test_python_does_not_recompute_engine_ten_stem_lu():
    constants = (ROOT / "constants.json").read_text(encoding="utf-8")
    operators = (ROOT / "operators_natal.py").read_text(encoding="utf-8")
    assert "十干禄" not in constants
    assert "_target_lu_branches(chart, tens, const)" not in operators
    assert 'full.get("lu_roots"' in operators


def test_python_does_not_recompute_engine_atomic_facts():
    operators = (ROOT / "operators_natal.py").read_text(encoding="utf-8")
    assert 'full.get("atomic_facts"' in operators
    for fact in (
        "day_master_element",
        "month_longevity",
        "year_stem_ten_god",
        "month_main_ten_god",
        "hour_stem_ten_god",
        "pattern_god_transparent",
        "pillar_punishments",
        "officer_killing_cleaned",
        "wealth_tomb_present",
        "wealth_star_in_tomb",
        "spouse_palace_state",
        "day_branch_type",
        "year_officer_killing",
    ):
        assert fact in operators
    assert "wealthTomb" not in operators
    assert "OfficerKillingCleaned" not in operators




def test_python_does_not_recompute_engine_ten_god_and_element_states():
    operators = _source(operators_natal)
    constants = (ROOT / "constants.json").read_text(encoding="utf-8")
    assert 'full.get("ten_god_states"' in operators
    assert 'full.get("element_states"' in operators
    assert 'full.get("da_yun"' in operators
    assert 'selected.get("rooted"' in operators
    for forbidden in ("天干五行", "地支五行", "五行生克", "得令状态", "十神旺弱规则"):
        assert forbidden not in constants
        assert f'const["{forbidden}"]' not in operators
    for forbidden in ("算子柱位", "柱刑关系字段", "格局十神"):
        assert forbidden not in constants


def test_natal_python_layer_does_not_aggregate_pillar_ten_gods():
    operators = _source(operators_natal)
    assert "_ten_god_states_from_pan" in operators
    assert "_shishen_from_pan" not in operators
    projection_end = operators.index("def _pillar_key")
    assert 'get("shi_shens", [])' not in operators[:projection_end]


def test_natal_relation_groups_come_from_engine():
    operators = _source(operators_natal)
    assert 'get("relation_groups", [])' in operators
    assert "chart_zhi" not in operators
    assert "关系字段类型" not in (ROOT / "constants.json").read_text(encoding="utf-8")


def test_python_does_not_recompute_engine_liunian_atomic_facts():
    operators = (ROOT / "operators_liunian.py").read_text(encoding="utf-8")
    assert "atomic_facts" in operators
    for fact in (
        "controls_targets",
        "unfavorable_gan",
        "unfavorable_branch",
        "wealth_breaks_seal",
        "combinations",
        "year_branch_relations",
        "year_branch_controlled_by",
        "year_gan_controls_day_gan",
        "day_gan_controls_year_gan",
        "gan_combines",
        "dayun_gan_combines",
        "dayun_zhi_clashes_year",
        "day_void_branches",
        "year_equals_day_pillar",
        "dayun_equals_year_pillar",
        "year_gan_equals_natal_year_gan",
    ):
        assert fact in operators
    for forbidden_lookup in (
        "const['五行生克']",
        "const['天干五合']",
        "const['关系取冲类型']",
        "const['旬空']",
    ):
        assert forbidden_lookup not in operators


def test_yearly_eval_has_no_duanyu_dependency():
    src = _source(yearly_eval)
    assert "import duanyu" not in src and "from duanyu" not in src


def test_assertion_store_has_no_factor_or_duanyu_dependency():
    src = _source(assertion_store)
    assert "import factors" not in src and "from factors" not in src
    assert "import duanyu" not in src and "from duanyu" not in src


def test_error_module_is_standalone():
    src = _source(errors)
    assert "import factors" not in src
    assert "import duanyu" not in src
    assert "import operators" not in src


def test_pan_schema_has_no_factor_or_duanyu_dependency():
    src = _source(pan_schema)
    assert "import factors" not in src and "from factors" not in src
    assert "import duanyu" not in src and "from duanyu" not in src


def test_factor_tables_has_no_duanyu_dependency():
    src = _source(factor_tables)
    assert "import duanyu" not in src and "from duanyu" not in src


def test_operators_have_no_duanyu_dependency():
    src = _source(operators_natal) + _source(operators_liunian)
    assert "import duanyu" not in src and "from duanyu" not in src


def test_factor_facade_owns_snap_and_error_contract():
    src = (ROOT / "factors.py").read_text(encoding="utf-8")
    assert "project_natal_facts" in src
    assert "FactorEvaluateError" in src


def test_ziwei_gong_matching_is_atomic_not_python_logic():
    operators = _source(operators_natal)
    assert '"palace_facts"' in operators
    for forbidden in (
        "紫微亮度分组",
        "紫微星组别名",
        "紫微主星",
    ):
        assert f'const.get("{forbidden}"' not in operators
        assert f'const["{forbidden}"]' not in operators
