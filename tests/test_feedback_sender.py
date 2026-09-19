"""Tests for the deterministic feedback sender bundled with every skill."""

import importlib.util
import json
import sys
import urllib.error
from io import StringIO
from unittest import mock

from _helpers import SKILL_ROOT, skill_dir, skill_version


def load_sender(skill="liki"):
    return load_sender_from(skill_dir(skill) / "feedback.py", skill)


def load_sender_from(path, name="feedback"):
    spec = importlib.util.spec_from_file_location(f"feedback_{name}", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def payload() -> dict:
    return {
        "meta": {
            "skill": "liki",
            "skill_version": skill_version(),
            "engine_version": skill_version(),
        },
        "agent": {"name": "codex-cli", "version": "0.21.6"},
        "llm": {"provider": "openai", "model": "gpt-5.1"},
        "problem": {
            "type": "clarity",
            "severity": "medium",
            "tool": "ziwei.liuyue",
            "summary": "resolved_period needs a clearer interpretation entry",
        },
    }


def test_unified_skill_has_one_sender():
    senders = list(SKILL_ROOT.rglob("feedback.py"))
    assert [p.relative_to(SKILL_ROOT) for p in senders] == [__import__("pathlib").Path("feedback.py")]


def test_sender_builds_valid_payload_and_uses_endpoint(monkeypatch):
    sender = load_sender()
    sent = {}

    def fake_urlopen(request, timeout):
        sent["request"] = request
        sent["timeout"] = timeout
        response = mock.MagicMock()
        response.status = 204
        response.__enter__.return_value = response
        response.__exit__.return_value = False
        return response

    monkeypatch.setattr(sender.urllib.request, "urlopen", fake_urlopen)
    monkeypatch.setenv("LIKI_FEEDBACK_URL", "https://liki.test/api/feedback")
    monkeypatch.setattr(sys, "stdin", StringIO(json.dumps(payload())))
    status = sender.main([])
    assert status == 0
    assert sent["timeout"] == 2
    request = sent["request"]
    assert request.full_url == "https://liki.test/api/feedback"
    body = json.loads(request.data.decode("utf-8"))
    source = payload()
    assert body == {
        "schema_version": "feedback-v1",
        "meta": source["meta"] | {"source": "skill-agent"},
        "agent": source["agent"],
        "llm": source["llm"],
        "problem": source["problem"],
    }


def test_sender_honors_disabled(monkeypatch):
    sender = load_sender()
    monkeypatch.setenv("LIKI_FEEDBACK_DISABLED", "1")
    monkeypatch.setenv("LIKI_FEEDBACK_URL", "https://liki.test/api/feedback")

    def fail(*args, **kwargs):
        raise AssertionError("disabled sender must not send")

    monkeypatch.setattr(sender.urllib.request, "urlopen", fail)
    assert sender.main([]) == 0


def test_sender_host_overrides_win_over_payload(monkeypatch, tmp_path):
    sender = load_sender()
    payload_path = tmp_path / "payload.json"
    source = payload()
    payload_path.write_text(json.dumps(source), encoding="utf-8")
    sent = {}

    def fake_urlopen(request, timeout):
        sent["request"] = request
        response = mock.MagicMock()
        response.status = 204
        response.__enter__.return_value = response
        response.__exit__.return_value = False
        return response

    monkeypatch.setattr(sender.urllib.request, "urlopen", fake_urlopen)
    monkeypatch.setenv("LIKI_FEEDBACK_URL", "https://liki.test/api/feedback")
    monkeypatch.setenv("LIKI_FEEDBACK_SKILL", "liki")
    monkeypatch.setenv("LIKI_FEEDBACK_SKILL_VERSION", "2026.01.01.0")
    monkeypatch.setenv("LIKI_ENGINE_VERSION", "2026.01.01.0")
    monkeypatch.setenv("LIKI_FEEDBACK_SESSION_HASH", "sha256:" + "0" * 64)
    assert sender.main(["--payload-file", str(payload_path)]) == 0
    body = json.loads(sent["request"].data.decode("utf-8"))
    assert body["meta"] == {
        "source": "skill-agent",
        "skill": "liki",
        "skill_version": "2026.01.01.0",
        "engine_version": "2026.01.01.0",
        "session_hash": "sha256:" + "0" * 64,
    }


def test_sender_unknown_defaults_fill_partial_host_facts(monkeypatch):
    sender = load_sender()
    source = payload()
    source.pop("agent")
    sent = {}

    def fake_urlopen(request, timeout):
        sent["request"] = request
        response = mock.MagicMock()
        response.status = 204
        response.__enter__.return_value = response
        response.__exit__.return_value = False
        return response

    monkeypatch.setattr(sender.urllib.request, "urlopen", fake_urlopen)
    monkeypatch.setenv("LIKI_FEEDBACK_URL", "https://liki.test/api/feedback")
    monkeypatch.setattr(sys, "stdin", StringIO(json.dumps(source)))
    assert sender.main([]) == 0
    body = json.loads(sent["request"].data.decode("utf-8"))
    assert body["agent"] == {"name": "unknown", "version": "unknown"}


def test_sender_ignores_context_file_env(monkeypatch, tmp_path):
    context_path = tmp_path / "context.json"
    context_path.write_text(json.dumps({"agent": {"name": "host-agent"}}), encoding="utf-8")
    sender = load_sender()
    source = payload()
    sent = {}

    def fake_urlopen(request, timeout):
        sent["request"] = request
        response = mock.MagicMock()
        response.status = 204
        response.__enter__.return_value = response
        response.__exit__.return_value = False
        return response

    monkeypatch.setattr(sender.urllib.request, "urlopen", fake_urlopen)
    monkeypatch.setenv("LIKI_FEEDBACK_URL", "https://liki.test/api/feedback")
    monkeypatch.setenv("LIKI_FEEDBACK_CONTEXT", str(context_path))
    monkeypatch.setattr(sys, "stdin", StringIO(json.dumps(source)))
    assert sender.main([]) == 0
    body = json.loads(sent["request"].data.decode("utf-8"))
    assert body["agent"] == source["agent"]


def test_sender_rejects_unknown_private_field_without_posting(monkeypatch):
    sender = load_sender()
    bad = payload()
    bad["conversation"] = "private user text"

    def fail(*args, **kwargs):
        raise AssertionError("invalid payload must not send")

    monkeypatch.setattr(sender.urllib.request, "urlopen", fail)
    assert sender.main([]) == 0


def test_sender_ignores_network_failure(monkeypatch):
    sender = load_sender()
    monkeypatch.setattr(
        sender.urllib.request,
        "urlopen",
        mock.Mock(side_effect=urllib.error.URLError("down")),
    )
    assert sender.main([]) == 0
