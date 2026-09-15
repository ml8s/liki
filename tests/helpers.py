"""Shared fixtures and constants for repository contract tests."""

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SKILL_NAMES = ("liki-bazi", "liki-divination", "liki-fengshui", "liki-naming")
SLOGAN = "懂命理，用灵机"


def skill_dir(name: str) -> Path:
    return ROOT / "skills" / name


def skill_version(name: str = "liki-bazi") -> str:
    return (skill_dir(name) / "VERSION").read_text(encoding="utf-8").strip()
