"""问卦文档运行时校验；公开契约唯一源是 Pydantic model。"""
from __future__ import annotations

from app.contracts.divination import (
    DIVINATION_TOOLS,
    LiuyaoQueryOutput,
    LiuyaoSnapshotOutput,
    QimenQueryOutput,
    QimenSnapshotOutput,
)
from pydantic import ValidationError

DOCUMENT_MODELS = {
    "liuyao_snapshot": LiuyaoSnapshotOutput,
    "qimen_snapshot": QimenSnapshotOutput,
    "liuyao_query": LiuyaoQueryOutput,
    "qimen_query": QimenQueryOutput,
}
_CONTRACT_SCHEMAS = {tool.name: tool.output_schema for tool in DIVINATION_TOOLS}


def load_contract(name: str) -> dict:
    """Return the already-published Pydantic schema for inspection/tests."""
    try:
        return _CONTRACT_SCHEMAS[name]
    except KeyError as exc:
        raise ValueError(f"unknown contract: {name}") from exc


def validate_document(name: str, document: dict) -> None:
    """校验失败抛 ValueError；成功返回 None。"""
    model = DOCUMENT_MODELS.get(name)
    if model is None:
        raise ValueError(f"unknown contract: {name}")
    try:
        model.model_validate(document)
    except ValidationError as exc:
        details = "; ".join(
            f"{'.'.join(str(part) for part in error['loc']) or '$'}: {error['msg']}"
            for error in exc.errors()
        )
        raise ValueError(f"{name} contract failed: {details}") from exc
