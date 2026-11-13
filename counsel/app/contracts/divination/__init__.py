"""Backward-compatible public surface for divination contracts."""

from .common import DivinationConflict, DivinationFollowupPolicy, DivinationSameMatterFollowupPolicy
from .liuyao import *  # noqa: F401,F403
from .liuyao import TOOL_LIUYAO_ASK
from .qimen import *  # noqa: F401,F403
from .qimen import TOOL_QIMEN_ASK

DIVINATION_TOOLS = (TOOL_LIUYAO_ASK, TOOL_QIMEN_ASK)

__all__ = [
    "DIVINATION_TOOLS",
    "DivinationConflict",
    "DivinationFollowupPolicy",
    "DivinationSameMatterFollowupPolicy",
    "TOOL_LIUYAO_ASK",
    "TOOL_QIMEN_ASK",
]
