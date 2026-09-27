#!/usr/bin/env python3
"""Deterministic OpenAI Responses API fixture for Liki system E2E.

This is a test adapter for the official ADK OpenAI-compatible model client. It
is deliberately small and standards-shaped: production still uses the official
AG-UI, MCP, ADK, and OpenAI Responses protocol implementations.
"""
from __future__ import annotations

import json
import os
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any

HOST = os.getenv("FIXTURE_ADDR", "0.0.0.0")
PORT = int(os.getenv("FIXTURE_PORT", "8090"))
FINAL_ANSWER = os.getenv("FIXTURE_ANSWER", "E2E_FULL_CAPABILITY_ANSWER")

lock = threading.Lock()
state: dict[str, Any] = {
    "requests": 0,
    "saw_transfer_tool": False,
    "saw_transfer_call": False,
    "saw_bazi_tool": False,
    "saw_bazi_call": False,
    "saw_profile_context": False,
    "function_calls": [],
    "last_request": None,
}


def walk(value: Any):
    if isinstance(value, dict):
        yield value
        for item in value.values():
            yield from walk(item)
    elif isinstance(value, list):
        for item in value:
            yield from walk(item)


def tool_names(document: dict[str, Any]) -> set[str]:
    names: set[str] = set()
    for tool in document.get("tools") or []:
        if isinstance(tool, dict):
            if isinstance(tool.get("name"), str):
                names.add(tool["name"])
            function = tool.get("function")
            if isinstance(function, dict) and isinstance(function.get("name"), str):
                names.add(function["name"])
    return names


def has_function_output(document: dict[str, Any], name: str) -> bool:
    for item in walk(document.get("input")):
        if item.get("type") != "function_call_output":
            continue
        if item.get("name") == name or item.get("call_id") == f"call_{name}":
            return True
    return False


def response_body(output: list[dict[str, Any]]) -> dict[str, Any]:
    return {
        "id": f"resp_e2e_{state['requests']}",
        "model": "liki-e2e-full",
        "status": "completed",
        "output": output,
        "usage": {"input_tokens": 1, "output_tokens": 1, "total_tokens": 2},
    }


def function_call(name: str, arguments: dict[str, Any]) -> list[dict[str, Any]]:
    return [{
        "type": "function_call",
        "call_id": f"call_{name}",
        "name": name,
        "arguments": json.dumps(arguments, ensure_ascii=False, separators=(",", ":")),
    }]


def final_output() -> list[dict[str, Any]]:
    return [{
        "type": "message",
        "role": "assistant",
        "content": [{"type": "output_text", "text": FINAL_ANSWER}],
    }]


def choose_response(document: dict[str, Any]) -> tuple[list[dict[str, Any]], str | None, str | None]:
    names = tool_names(document)
    serialized = json.dumps(document, ensure_ascii=False)

    if "bazi_chart" in names and not has_function_output(document, "bazi_chart"):
        return (
            function_call("bazi_chart", {
                "solar_time": "1990-05-20T12:00:00+08:00",
                "gender": "male",
            }),
            "bazi_chart",
            "engine-tool",
        )
    if "bazi_chart" in names and has_function_output(document, "bazi_chart"):
        return final_output(), None, None
    if "transfer_to_agent" in names and not has_function_output(document, "transfer_to_agent"):
        return (
            function_call("transfer_to_agent", {"agent_name": "bazi"}),
            "transfer_to_agent",
            "bazi",
        )
    return final_output(), None, None


class Handler(BaseHTTPRequestHandler):
    server_version = "LikiE2EResponses/1"

    def log_message(self, _format: str, *args: Any) -> None:
        return

    def send_json(self, status: int, document: dict[str, Any]) -> None:
        raw = json.dumps(document, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def do_GET(self) -> None:
        if self.path == "/healthz":
            self.send_json(200, {"status": "ok"})
            return
        if self.path == "/stats":
            with lock:
                self.send_json(200, dict(state))
            return
        if self.path == "/last-request":
            with lock:
                self.send_json(200, state.get("last_request") or {})
            return
        self.send_json(404, {"error": "not_found"})

    def do_POST(self) -> None:
        if self.path != "/v1/responses":
            self.send_json(404, {"error": "not_found"})
            return
        try:
            length = int(self.headers.get("Content-Length", "0"))
            document = json.loads(self.rfile.read(length))
        except (ValueError, json.JSONDecodeError) as error:
            self.send_json(400, {"error": {"message": str(error)}})
            return

        output, call_name, call_kind = choose_response(document)
        serialized_document = json.dumps(document, ensure_ascii=False)
        with lock:
            state["requests"] += 1
            state["last_request"] = document
            state["saw_transfer_tool"] = state["saw_transfer_tool"] or "transfer_to_agent" in tool_names(document)
            state["saw_bazi_tool"] = state["saw_bazi_tool"] or "bazi_chart" in tool_names(document)
            state["saw_profile_context"] = state["saw_profile_context"] or (
                "1990-05-20" in serialized_document and "solar" in serialized_document
            )
            if call_name:
                state["function_calls"].append({"name": call_name, "kind": call_kind})
            if call_name == "transfer_to_agent":
                state["saw_transfer_call"] = True
            if call_name == "bazi_chart":
                state["saw_bazi_call"] = True

        body = response_body(output)
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.end_headers()
        events = [
            {"type": "response.created", "response": {"id": body["id"], "model": body["model"], "status": "in_progress"}},
        ]
        if not call_name:
            events.append({"type": "response.output_text.delta", "delta": FINAL_ANSWER})
        events.append({"type": "response.completed", "response": body})
        for event in events:
            self.wfile.write(f"data: {json.dumps(event, ensure_ascii=False, separators=(',', ':'))}\n\n".encode("utf-8"))
        self.wfile.write(b"data: [DONE]\n\n")


if __name__ == "__main__":
    server = ThreadingHTTPServer((HOST, PORT), Handler)
    print(f"liki E2E OpenAI Responses fixture listening on {HOST}:{PORT}", flush=True)
    server.serve_forever()
