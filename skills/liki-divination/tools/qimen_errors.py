"""奇门工具层错误类型。"""
from __future__ import annotations


class LikiQimenToolError(Exception):
    """奇门工具链错误。"""


class RPCError(LikiQimenToolError):
    """JSON-RPC 网络或协议错误。"""


class TableError(LikiQimenToolError):
    """事象表或解释表结构错误。"""
