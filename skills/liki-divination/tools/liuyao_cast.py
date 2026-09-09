"""六爻起卦编排；原始硬币一律交给 engine 归一化。"""
from __future__ import annotations

from typing import Any

from divination_rpc import engine_data


def qigua(
    *,
    mode: str = "auto",
    rounds: list[list[str]] | None = None,
    yaos: list[int] | None = None,
    seed: int | None = None,
) -> dict:
    """调用 engine 的 liuyao.qigua；本函数不换算硬币、不判断动爻。"""
    if mode not in {"auto", "coins", "yaos"}:
        raise ValueError("mode must be auto, coins, or yaos")
    params: dict[str, Any] = {"mode": mode}
    if mode == "auto":
        if rounds is not None or yaos is not None:
            raise ValueError("auto mode does not accept rounds or yaos")
        if seed is not None:
            params["seed"] = seed
    elif mode == "coins":
        if rounds is None:
            raise ValueError("coins mode requires rounds")
        if yaos is not None or seed is not None:
            raise ValueError("coins mode accepts rounds only")
        if not isinstance(rounds, list) or len(rounds) != 6:
            raise ValueError("rounds must contain exactly 6 rounds")
        for index, round_coins in enumerate(rounds, 1):
            if not isinstance(round_coins, list) or len(round_coins) != 3:
                raise ValueError(f"round {index} must contain exactly 3 coins")
        params["rounds"] = rounds
    else:
        if yaos is None:
            raise ValueError("yaos mode requires yaos")
        if rounds is not None or seed is not None:
            raise ValueError("yaos mode accepts yaos only")
        if not isinstance(yaos, list) or len(yaos) != 6:
            raise ValueError("yaos must contain exactly 6 values")
        params["yaos"] = yaos

    result = engine_data("liuyao.qigua", params)
    casting = result.get("casting")
    if not isinstance(casting, dict):
        raise ValueError("engine returned liuyao.qigua without casting receipt")
    return {"casting": casting}
