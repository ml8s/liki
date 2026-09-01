"""模块边界：领域快照、pan 契约、表加载层不得反向依赖断语层。"""
from pathlib import Path

import _helpers  # noqa: F401
import domain_snapshot
import assertion_store
import errors
import factor_tables
import operators_liunian
import operators_natal
import pan_schema
import yearly_eval

ROOT = Path(domain_snapshot.__file__).parent


def _source(module):
    return Path(module.__file__).read_text(encoding="utf-8")


def test_domain_snapshot_has_no_factor_or_duanyu_dependency():
    src = _source(domain_snapshot)
    assert "import factors" not in src and "from factors" not in src
    assert "import duanyu" not in src and "from duanyu" not in src


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
    assert "project_domain_facts" in src
    assert "FactorEvaluateError" in src
