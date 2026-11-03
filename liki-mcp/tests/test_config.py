from app.config import load_config


def test_forbidden_tools_cannot_be_exposed(monkeypatch):
    monkeypatch.setenv("LIKI_EXPOSE_CONFIG", "/tmp/bad-expose.json")
    from pathlib import Path
    Path("/tmp/bad-expose.json").write_text(
        '{"expose": {"engine": ["qimen_chart"]}}',
        encoding="utf-8",
    )
    try:
        load_config()
        raise AssertionError("forbidden exposure was accepted")
    except ValueError as error:
        assert "qimen_chart" in str(error)
