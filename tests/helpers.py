"""Shared fixtures and constants for repository contract tests."""

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SKILL_ROOT = ROOT / "skills" / "liki"
DOMAIN_NAMES = ("bazi", "divination", "fengshui", "naming")
SLOGAN = "懂命理，用灵机"


def skill_dir(name: str = "liki") -> Path:
    if name in DOMAIN_NAMES:
        return SKILL_ROOT / name
    return SKILL_ROOT


def skill_version(name: str = "liki") -> str:
    return (skill_dir(name) / "VERSION").read_text(encoding="utf-8").strip()
