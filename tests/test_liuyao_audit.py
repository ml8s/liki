from __future__ import annotations

import importlib.util
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

from liuyao_audit import audit_report  # noqa: E402


def _chart() -> dict:
    return {
        "chart": {
            "ben_gua": "乾",
            "bian_gua": "姤",
            "lines": [
                {"position": index, "liu_qin": "子孙", "shi_ying": "世" if index == 1 else "应" if index == 4 else ""}
                for index in range(1, 7)
            ],
        }
    }


def test_audit_accepts_correct_deterministic_facts():
    result = audit_report("本卦为乾，变卦是姤。初爻为子孙，世爻在初爻，应爻在四爻。", _chart())
    assert result["accepted"] is True
    assert result["errors"] == []


def test_audit_rejects_chart_and_relative_errors():
    result = audit_report(
        "本卦为坤，变卦是复。二爻为妻财，世爻在上爻，应爻在三爻。",
        _chart(),
    )
    assert result["accepted"] is False
    types = {error["type"] for error in result["errors"]}
    assert types == {"本卦_name", "变卦_name", "line_relative", "世_position", "应_position"}


def test_audit_leaves_inference_unchecked():
    result = audit_report("这件事大概率困难，应在三个月内。", _chart())
    assert result["accepted"] is True
    assert "应期推断" in result["unchecked"]
