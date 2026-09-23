"""奇门 Python 编排层：事象路由与解释表只消费 engine 输出。"""
from __future__ import annotations

import json
import csv
import importlib.util
import sys
from pathlib import Path

import pytest
from jsonschema import validate


TOOLS = Path(__file__).resolve().parents[1] / "analysis/liki_analysis/divination/tools"
SKILL_ROOT = TOOLS.parent.parent
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import divination_rpc  # noqa: E402
import qimen_paipan as paipan  # noqa: E402
from qimen_duanyu import query  # noqa: E402
from qimen_projection import (  # noqa: E402
    load_factors_contract,
    project as project_chart,
    validate as validate_standard_factors,
)
from divination_rpc import HTTPError, RPCError  # noqa: E402
from qimen_errors import TableError  # noqa: E402
from qimen_interpretations import (  # noqa: E402
    load_interpretation_index,
    load_rule_table,
)
from qimen_matters import load_matter_table as load_routing_matter_table  # noqa: E402



def _load_agent_cli():
    spec = importlib.util.spec_from_file_location(
        "qimen_agent_cli", TOOLS / "agent_cli.py"
    )
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    sys.path.insert(0, str(TOOLS))
    try:
        spec.loader.exec_module(module)
    finally:
            return module


def _project(pan):
    """Project the chart payload returned by the local *_pan helpers."""
    if not isinstance(pan, dict) or not isinstance(pan.get("chart"), dict):
        raise AssertionError("qimen test fixture must contain chart")
    return project_chart(pan["chart"])


def _pan(**changes):
    chart = {
        "method": {
            "scope": "hour", "school": "zhuanpan",
            "spirit_mode": "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu",
        },
        "pan": {
            "yin_dun": False,
            "jushu": 5,
            "zhi_fu_xing": "天芮",
            "zhi_shi_men": "死门",
        "ri_gan": "甲",
        "ri_zhi": "辰",
        "nian_gan": "丙",
        "nian_zhi": "午",
        "yue_gan": "丙",
        "yue_zhi": "申",
        "shi_gan": "甲",
        "shi_zhi": "午",
        "wu_bu_yu_shi": False,
            "gong_wei": [
                {"gong": {"name": "坎", "luoshu": 1}},
                {"gong": {"name": "坤", "luoshu": 2}},
                {"gong": {"name": "震", "luoshu": 3}},
                {"gong": {"name": "巽", "luoshu": 4}},
                {"gong": {"name": "离", "luoshu": 9}},
                {"gong": {"name": "艮", "luoshu": 8}},
                {"gong": {"name": "兑", "luoshu": 7}},
                {"gong": {"name": "乾", "luoshu": 6}},
            ],
        },
        "ri_gan_gong": "震",
        "shi_gan_gong": "兑",
        "ri_gan_palace_facts": [{"palace": "震", "layer": "heaven"}],
        "shi_gan_palace_facts": [{"palace": "兑", "layer": "heaven"}],
        "shi_gan_domain": "outer",
        "ri_shi_relation": {
            "subject": "ri_gan_gong",
            "object": "shi_gan_gong",
            "relation": "same",
            "name": "日干宫与时干宫比和",
        },
        "patterns": [],
        "kong_wang_affected": [],
        "ma_xing_affected": [],
        "palace_wang_shuai": [],
        "shi_gan_gong_wang_shuai": {
            "gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死",
        },
        "ying_qi": {"candidates": []},
        "gan_interaction": [],
        "men_interaction": [],
        "xing_interaction": [],
        "xing_gong_wu_xing": [],
        "men_po": [],
        "men_zhi": [],
        "zhi_fu_xing_gong": "坎",
        "zhi_shi_men_gong": "坎",
        "specialized": {
            "geng_ge": [], "lost_context": [],
            "thief_context": [], "tianwang_context": [],
        },
    }
    chart.update(changes)
    return {"matter": None, "chart": chart}


def _lost_pan(**changes):
    pan = _pan()
    pan["chart"]["specialized"]["lost_context"] = [{
        "gong": "兑", "wang_shuai": "旺",
    }]
    pan["chart"].update({
        "patterns": [{
            "name": "反吟",
            "gong_wei": ["坎"],
            "auspicious": False,
        }],
        "kong_wang_affected": [{
            "symbol": "时干",
            "role": "pillar",
            "branch": "戌",
            "gong": "乾",
        }],
        "ri_shi_relation": {
            "subject": "ri_gan_gong",
            "object": "shi_gan_gong",
            "relation": "shi_generates_ri",
            "name": "时干宫生日干宫",
        },
        "palace_wang_shuai": [{
            "gong": "兑",
            "wuxing": "金",
            "wang_shuai": "旺",
            "wang_shuai_name": "旺",
        }],
        "shi_gan_gong_wang_shuai": {
            "gong": "兑",
            "wuxing": "金",
            "wang_shuai": "旺",
            "wang_shuai_name": "旺",
        },
    })
    pan["chart"].update(changes)
    return pan


def _thief_pan(**changes):
    pan = _pan()
    pan["chart"]["specialized"]["thief_context"] = [
        {"role": "大贼", "symbol": "天蓬", "gong": "坎", "wang_shuai_label": "囚", "lucky_pattern_names": []},
        {"role": "小贼", "symbol": "玄武", "gong": "坎", "wang_shuai_label": "囚", "lucky_pattern_names": []},
    ]
    pan["chart"]["pan"]["gong_wei"] = [
        {
            "gong": {"name": "坤", "luoshu": 2},
            "tian_pan": [{"gan": "己", "xing": "天芮"}],
            "men": "死门",
            "men_present": True,
            "shen": "勾陈",
            "shen_present": True,
        },
        {
            "gong": {"name": "坎", "luoshu": 1},
            "tian_pan": [{"gan": "戊", "xing": "天蓬"}],
            "men": "休门",
            "men_present": True,
            "shen": "朱雀",
            "shen_present": True,
        },
    ]
    pan["chart"].update({
        "palace_wang_shuai": [
            {"gong": "坤", "wuxing": "土", "wang_shuai": "相", "wang_shuai_name": "相"},
            {"gong": "坎", "wuxing": "水", "wang_shuai": "囚", "wang_shuai_name": "囚"},
            {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
        ],
    })
    pan["chart"].update(changes)
    return pan


def _rich_pan():
    rich = _pan(
        ri_shi_relation={
            "subject": "ri_gan_gong",
            "object": "shi_gan_gong",
            "relation": "shi_generates_ri",
            "name": "时干宫生日干宫",
        },
        patterns=[
            {"name": "反吟", "gong_wei": ["坎"], "auspicious": False},
            {"name": "青龙返首", "gong_wei": ["震"], "auspicious": True},
        ],
        kong_wang_affected=[{
            "symbol": "时干", "role": "pillar", "branch": "戌", "gong": "乾",
        }],
        ma_xing_affected=[{
            "symbol": "日干", "role": "pillar", "branch": "寅", "gong": "艮",
        }],
        ying_qi={"candidates": [
            {
                "type": "ma_xing", "branch": "寅", "gong": "艮",
                "related_to": [{"symbol": "日干", "role": "pillar"}],
            },
            {
                "type": "kong_wang", "branch": "戌", "gong": "乾",
                "related_to": [{"symbol": "时干", "role": "pillar"}],
            },
        ]},
        gan_interaction=[{
            "name": "青龙返首", "gong": "震",
            "tian_pan_gan": "戊", "di_pan_gan": "丙", "auspicious": True,
        }],
        men_interaction=[{"name": "门迫", "door": "休门", "gong": "坎"}],
        xing_interaction=[{
            "name": "星克宫", "xing": "天蓬", "gong": "离", "auspicious": False,
        }],
        xing_gong_wu_xing=[{
            "xing": "天蓬", "gong": "离", "relation": "xing_controls_gong",
            "relation_name": "星克宫", "traditional_label": "囚",
        }],
        palace_wang_shuai=[
            {"gong": "离", "wuxing": "火", "wang_shuai": "旺", "wang_shuai_name": "旺"},
            {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
        ],
        shi_gan_gong_wang_shuai={
            "gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死",
        },
        men_po=["坎"],
        men_zhi=["离"],
        zhi_fu_xing_gong="乾",
        zhi_shi_men_gong="开",
        yong_shen={
            "nian_gan_gong": "乾",
            "symbols": [
                {
                    "symbol": "生门",
                    "palace": "艮",
                    "tian_gan": ["戊"],
                    "kong_wang_branches": [],
                    "ma_xing_branch": None,
                },
                {
                    "symbol": "戊",
                    "palace": "震",
                    "tian_gan": ["庚"],
                    "kong_wang_branches": [],
                    "ma_xing_branch": None,
                },
            ],
        },
        method={
            "scope": "hour", "school": "zhuanpan",
            "spirit_mode": "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu",
        },
    )
    rich["chart"]["pan"]["gong_wei"] = [{
        "gong": {"name": "兑", "luoshu": 7},
        "tian_pan": [{"gan": "庚", "xing": "天蓬"}],
        "men": "杜门",
        "men_present": True,
        "shen": "勾陈",
        "shen_present": True,
    }]
    return rich


def _missing_pan(**changes):
    pan = _pan()
    pan["chart"]["shi_gan_gong"] = "坎"
    pan["chart"]["yong_shen"] = {
        "nian_gan_gong": None,
        "symbols": [{
            "symbol": "六合", "palace": "坎", "tian_gan": ["戊"],
            "kong_wang_branches": [], "ma_xing_branch": None,
        }],
    }
    pan["chart"]["gong_domains"] = [
        {"gong": "坎", "domain": "inner"},
        {"gong": "兑", "domain": "outer"},
    ]
    pan["chart"].update(changes)
    return pan


def _escape_pan(**changes):
    pan = _pan()
    pan["chart"].setdefault("specialized", {
        "geng_ge": [], "lost_context": [], "thief_context": [], "tianwang_context": [],
    })
    pan["chart"]["pan"]["gong_wei"] = [
        {
            "gong": {"name": "坎", "luoshu": 1},
            "tian_pan": [{"gan": "戊", "xing": "天蓬"}],
            "men": "休门",
            "men_present": True,
            "shen": "六合",
            "shen_present": True,
        },
        {
            "gong": {"name": "震", "luoshu": 3},
            "tian_pan": [{"gan": "庚", "xing": "天冲"}],
            "men": "伤门",
            "men_present": True,
            "shen": "值符",
            "shen_present": True,
        },
    ]
    pan["chart"].update({
        "palace_wang_shuai": [
            {"gong": "坎", "wuxing": "水", "wang_shuai": "囚", "wang_shuai_name": "囚"},
            {"gong": "震", "wuxing": "木", "wang_shuai": "休", "wang_shuai_name": "休"},
            {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
        ],
        "gan_interaction": [{
            "name": "伏吟", "gong": "坎", "tian_pan_gan": "戊",
            "di_pan_gan": "戊", "auspicious": False,
        }],
    })
    pan["chart"].update(changes)
    return pan


def test_matter_table_is_the_python_routing_fact(monkeypatch) -> None:
    matters = load_routing_matter_table()
    assert matters["wealth"]["name"] == "求财"
    assert matters["wealth"]["symbols"] == ["生门", "戊"]
    assert matters["missing_person"]["symbols"] == ["六合"]
    assert all(row["basis"] for row in matters.values())


def test_all_matter_requests_satisfy_engine_schema(monkeypatch) -> None:
    schema = json.loads(
        (Path(__file__).resolve().parents[1] / "engine/internal/agent/qimen_schema.json").read_text(
            encoding="utf-8"
        )
    )["params"]
    requests = []

    def fake_engine_data(method: str, params: dict) -> dict:
        assert method == "qimen.chart"
        requests.append(params)
        return {}

    monkeypatch.setattr(paipan, "engine_data", fake_engine_data)
    for matter in load_routing_matter_table():
        paipan.qimen_chart("2026-09-07T12:00:00+08:00", matter=matter)
    assert len(requests) == len(load_routing_matter_table())
    for params in requests:
        validate(params, schema)


def test_qimen_chart_maps_matter_before_engine_call(monkeypatch) -> None:
    calls = []

    def fake_engine_data(method: str, params: dict) -> dict:
        calls.append((method, params))
        return {"ri_gan_gong": "震", "shi_gan_gong": "兑"}

    monkeypatch.setattr(paipan, "engine_data", fake_engine_data)
    result = paipan.qimen_chart(
        "2026-09-07T12:00:00+08:00",
        matter="wealth",
        birth_date="1984-02-15",
        scope="hour",
        school="zhuanpan",
    )

    assert len(calls) == 1
    method, params = calls[0]
    assert method == "qimen.chart"
    assert "matter" not in params
    assert params["yong_shen"] == ["生门", "戊"]
    assert params["birth_date"] == "1984-02-15"
    assert result["matter"]["matter"] == "wealth"
    assert result["matter"]["name"] == "求财"
    assert result["chart"]["ri_gan_gong"] == "震"


def test_qimen_chart_rejects_duplicate_yong_shen_sources(monkeypatch) -> None:
    monkeypatch.setattr(paipan, "engine_data", lambda *_: (_ for _ in ()).throw(AssertionError("engine called")))
    try:
        paipan.qimen_chart("2026-09-07T12:00:00+08:00", matter="wealth", yong_shen=["生门"])
    except ValueError as exc:
        assert "互斥" in str(exc)
    else:
        raise AssertionError("matter and yong_shen must conflict")


def test_qimen_chart_rejects_empty_explicit_yong_shen(monkeypatch) -> None:
    monkeypatch.setattr(
        paipan,
        "engine_data",
        lambda *_: (_ for _ in ()).throw(AssertionError("engine called")),
    )
    with pytest.raises(ValueError, match="yong_shen 不能为空"):
        paipan.qimen_chart("2026-09-07T12:00:00+08:00", yong_shen=[])


def test_python_tool_schema_rejects_empty_or_duplicate_yong_shen() -> None:
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    qimen_snapshot = next(
        item["function"]["parameters"]
        for item in schema["tools"]
        if item["function"]["name"] == "qimen_snapshot"
    )["properties"]["yong_shen"]
    assert qimen_snapshot["minItems"] == 1
    assert qimen_snapshot["uniqueItems"] is True


def test_qimen_chart_rejects_duplicate_explicit_yong_shen(monkeypatch) -> None:
    monkeypatch.setattr(
        paipan,
        "engine_data",
        lambda *_: (_ for _ in ()).throw(AssertionError("engine called")),
    )
    with pytest.raises(ValueError, match="yong_shen 符号重复"):
        paipan.qimen_chart(
            "2026-09-07T12:00:00+08:00", yong_shen=["生门", "生门"]
        )


def test_qimen_chart_rejects_string_yong_shen(monkeypatch) -> None:
    monkeypatch.setattr(
        paipan,
        "engine_data",
        lambda *_: (_ for _ in ()).throw(AssertionError("engine called")),
    )
    with pytest.raises(ValueError, match="yong_shen 必须是字符串数组"):
        paipan.qimen_chart("2026-09-07T12:00:00+08:00", yong_shen="生门")


def test_matter_table_rejects_duplicate_symbols(tmp_path, monkeypatch) -> None:
    import qimen_matters

    matters = tmp_path / "matters.csv"
    matters.write_text(
        "matter,name,symbols,basis\n"
        "wealth,求财,生门、生门、戊,test\n",
        encoding="utf-8",
    )
    monkeypatch.setattr(qimen_matters, "_MATTER_TABLE", None)
    monkeypatch.setattr(qimen_matters, "MATTERS_PATH", matters)
    with pytest.raises(TableError, match="重复符号"):
        qimen_matters.load_matter_table()


import json as _json


def _mcp_response(payload: str) -> bytes:
    return f"event: message\ndata: {payload}\n\n".encode()


def _mcp_tools_call_result(obj) -> bytes:
    """构造 tools/call 的 SSE 响应，text 为序列化 JSON。"""
    result = {"jsonrpc": "2.0", "id": 1, "result": {
        "content": [{"type": "text", "text": _json.dumps(obj, ensure_ascii=False)}],
    }}
    return _mcp_response(_json.dumps(result, ensure_ascii=False))


class _FakeResp:
    def __init__(self, payload: bytes):
        self._payload = payload

    def read(self):
        return self._payload

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return False


def test_mcp_call_reports_malformed_engine_payload(monkeypatch) -> None:
    import engine_client
    from engine_client import MCPError

    monkeypatch.setattr(engine_client.urllib.request, "urlopen",
                        lambda *_, **__: _FakeResp(b"not-json"))
    with pytest.raises(MCPError, match="无法解析 MCP 响应"):
        engine_client.call("qimen.chart", {"solar_time": "x"})


def test_mcp_call_reports_missing_result(monkeypatch) -> None:
    import engine_client
    from engine_client import MCPError

    # tools/call 返回空 content → call 报错
    monkeypatch.setattr(engine_client.urllib.request, "urlopen",
                        lambda *_, **__: _FakeResp(_mcp_response('{"jsonrpc":"2.0","id":1,"result":{"content":[]}}')))
    with pytest.raises(MCPError):
        engine_client.call("qimen.chart", {"solar_time": "x"})


def test_mcp_retries_server_error_but_not_client_error(monkeypatch) -> None:
    import engine_client
    from engine_client import MCPError

    calls = []

    def urlopen(*_, **__):
        if not calls:
            calls.append("server-error")
            raise HTTPError("https://example.test", 500, "server error", {}, None)
        calls.append("ok")
        return _FakeResp(_mcp_tools_call_result({"ok": True}))

    monkeypatch.setattr(engine_client.urllib.request, "urlopen", urlopen)
    result = engine_client.call("qimen.chart", {"solar_time": "x"})
    assert result == {"data": {"ok": True}}
    assert calls == ["server-error", "ok"]

    client_error = HTTPError("https://example.test", 403, "forbidden", {}, None)
    monkeypatch.setattr(engine_client.urllib.request, "urlopen",
                        lambda *_, **__: (_ for _ in ()).throw(client_error))
    with pytest.raises(MCPError, match="HTTP 403"):
        engine_client.call("qimen.chart", {"solar_time": "x"})


def test_mcp_endpoint_reads_environment_on_each_call(monkeypatch) -> None:
    import engine_client

    seen = []

    def urlopen(request, *_, **__):
        seen.append(request.full_url)
        return _FakeResp(_mcp_tools_call_result({"ok": True}))

    monkeypatch.setenv("LIKI_MCP_URL", "https://dynamic.example/mcp")
    monkeypatch.setattr(engine_client.urllib.request, "urlopen", urlopen)
    assert engine_client.call("qimen.chart", {"solar_time": "x"}) == {"data": {"ok": True}}
    assert seen == ["https://dynamic.example/mcp"]
