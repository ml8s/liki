"""Backward-compatible surface for the naming domain pipeline."""

from .composition import COMPOSE_DEFAULT_NAMES, COMPOSE_MAX_CANDIDATES, COMPOSE_MAX_NAMES, compose_names, lookup_char
from .evaluation import evaluate_names
from .models import Character, SurnameRecord
from .selection import pick_chars_by_selected_yongshen
from .surnames import match_surnames

__all__ = [
    "COMPOSE_DEFAULT_NAMES",
    "COMPOSE_MAX_CANDIDATES",
    "COMPOSE_MAX_NAMES",
    "Character",
    "SurnameRecord",
    "compose_names",
    "lookup_char",
    "evaluate_names",
    "match_surnames",
    "pick_chars_by_selected_yongshen",
]
