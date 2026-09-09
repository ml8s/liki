"""Deprecated: use divination_rpc.call / engine_data.

Kept only as a temporary compatibility wrapper for local callers.
"""
from __future__ import annotations

from divination_rpc import RPCError, call, engine_data

__all__ = ["RPCError", "call", "engine_data"]
