"""问卦内部 case 校验；公开契约唯一源是 Pydantic model。"""

from pydantic import ValidationError

from app.contracts.divination import LiuyaoCase, QimenCase

DOCUMENT_MODELS = {
    "liuyao_case": LiuyaoCase,
    "qimen_case": QimenCase,
}


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
