#!/usr/bin/env python3
"""liki-divination 问卦路由 / 奇门 / 六爻工具的 CLI 适配器。"""
from __future__ import annotations

import json
import os
import sys

from qimen_read import read as qimen_read
from qimen_report import template as qimen_report_template
from qimen_report import validate_report as qimen_report_validate
from qimen_session import append_followup as qimen_session_append
from qimen_session import create_session as qimen_session_create
from qimen_session import validate_session as qimen_session_validate
from divination_router import route_question
from divination_contracts import validate_document
from huangli_days import days as huangli_days
from liuyao_cast import qigua as liuyao_qigua
from liuyao_audit import audit_report as liuyao_audit
from liuyao_report import template as liuyao_report_template
from liuyao_report import validate_report as liuyao_report_validate
from liuyao_session import append_followup, create_session, validate_session
from liuyao_topic_guidance import get_topic, project_topic_guidance
from liuyao_timing import rank_timing_candidates as liuyao_timing_plan
from liuyao_conditions import evaluate as liuyao_condition_evaluate
from liuyao_conditions import list_rules as liuyao_condition_list
from liuyao_read import read as liuyao_read


def _configure_windows_stdio() -> None:
    if os.name != "nt":
        return
    for stream in (sys.stdin, sys.stdout):
        try:
            stream.reconfigure(encoding="utf-8")
        except (AttributeError, OSError):
            pass
    try:
        sys.stderr.reconfigure(encoding="utf-8", errors="backslashreplace")
    except (AttributeError, OSError):
        pass


def _emit(payload: dict) -> None:
    print(json.dumps(payload, ensure_ascii=True))


_DISPATCH = {
    "divination_route": lambda args: route_question(
        args["question"],
        category=args.get("category"),
        specified_method=args.get("specified_method"),
    ),
    "qimen_read": lambda args: qimen_read(
        question=args["question"],
        city=args.get("city"),
        longitude=args.get("longitude"),
        time=args.get("time"),
        matter=args.get("matter"),
        yong_shen=args.get("yong_shen"),
        rule=args.get("rule"),
        scope=args.get("scope"),
        school=args.get("school"),
        dingju_method=args.get("dingju_method"),
        quarter_rule=args.get("quarter_rule"),
        base_dingju_method=args.get("base_dingju_method"),
        dun_source=args.get("dun_source"),
        hour_boundary=args.get("hour_boundary"),
        birth_date=args.get("birth_date"),
    ),
    "qimen_report": lambda args: (
        qimen_report_template(args["read_result"])
        if args["action"] == "template"
        else qimen_report_validate(args["report"], args["read_result"])
    ),
    "qimen_session": lambda args: (
        qimen_session_create(
            read_result=args["read_result"],
            headline=args.get("headline", ""),
            verdict=args.get("verdict", ""),
            confidence=args.get("confidence", "medium"),
            report=args.get("report"),
        )
        if args["action"] == "create"
        else qimen_session_append(
            args["session"],
            kind=args["kind"],
            text=args["text"],
            answer=args.get("answer", ""),
            outcome=args.get("outcome"),
        )
        if args["action"] == "append"
        else qimen_session_validate(args["session"])
    ),
    "liuyao_qigua": lambda args: liuyao_qigua(
        mode=args.get("mode", "auto"),
        rounds=args.get("rounds"),
        yaos=args.get("yaos"),
        seed=args.get("seed"),
    ),
    "liuyao_audit": lambda args: liuyao_audit(
        report=args["report"],
        chart=args["chart"],
    ),
    "liuyao_report": lambda args: (
        liuyao_report_template(args["snapshot"])
        if args["action"] == "template"
        else liuyao_report_validate(args["report"], args["snapshot"])
    ),
    "liuyao_session": lambda args: (
        create_session(
            question=args["question"],
            snapshot=args["snapshot"],
            headline=args.get("headline", ""),
            verdict=args.get("verdict", ""),
            confidence=args.get("confidence", "medium"),
            report=args.get("report"),
        )
        if args["action"] == "create"
        else append_followup(
            args["session"],
            kind=args["kind"],
            text=args["text"],
            answer=args.get("answer", ""),
            outcome=args.get("outcome"),
        )
        if args["action"] == "append"
        else validate_session(args["session"])
    ),
    "liuyao_topic_guidance": lambda args: (
        get_topic(args["topic"])
        if args.get("question") is None
        else project_topic_guidance(args["snapshot"], args["topic"])
    ),
    "liuyao_timing": lambda args: rank_timing_candidates(
        args["snapshot"],
        topic=args.get("topic"),
        limit=args.get("limit", 6),
    ),
    "liuyao_conditions": lambda args: (
        liuyao_condition_list()
        if args.get("action") == "list"
        else liuyao_condition_evaluate(args["snapshot"], topic=args.get("topic"))
    ),
    "huangli_days": lambda args: huangli_days(
        question=args["question"],
        event=args.get("event"),
        start_date=args.get("start_date"),
        end_date=args.get("end_date"),
        days=args.get("days"),
    ),
    "liuyao_read": lambda args: liuyao_read(
        casting=args.get("casting"),
        yaos=args.get("yaos"),
        solar_time=args.get("solar_time"),
        matter=args.get("matter"),
        yong_shen=args.get("yong_shen"),
        perspective=args.get("perspective"),
        question=args.get("question"),
        topic=args.get("topic"),
    ),
}


def _dispatch(fn: str, args: dict):
    handler = _DISPATCH.get(fn)
    if handler is None:
        raise ValueError(f"unknown tool: {fn}")
    return handler(args)


def main() -> int:
    raw = sys.stdin.read().strip()
    if not raw:
        _emit({"ok": False, "error": "empty stdin"})
        return 0
    try:
        request = json.loads(raw)
        args = request.get("args", {})
        if not isinstance(args, dict):
            raise ValueError("args must be an object")
        _emit({"ok": True, "data": _dispatch(request["fn"], args)})
    except KeyError as error:
        _emit({"ok": False, "error": f"missing arg: {error}"})
    except Exception as error:  # noqa: BLE001 — 工具链错误统一转为 JSON error
        _emit({"ok": False, "error": f"{type(error).__name__}: {error}"})
    return 0


if __name__ == "__main__":
    _configure_windows_stdio()
    sys.exit(main())
