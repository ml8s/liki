"""natal_chart canonical contract and fail-closed structural validation.

Pydantic v2 model owns the structural shape; domain cross-field invariants
(palace ordering, daxian palace uniqueness) live in model_validator.
"""

from __future__ import annotations

from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, TypeAdapter, ValidationError, model_validator

from app.natal.tools.chart_integrity import DIGEST_FIELD, validate_natal_chart_digest
from app.natal.tools.errors import ChartSchemaError
from app.natal.tools.factor_constants import load_constants

_CONSTANTS = load_constants()
REQUIRED_FIELDS = ("subject", "bazi", "ziwei", "ziwei_daxian", DIGEST_FIELD)
PILLARS = tuple(_CONSTANTS["四柱"])
PALACES = tuple(_CONSTANTS["紫微宫位"])
DAXIAN_COUNT = _CONSTANTS["大限段数"]


class _SubjectModel(BaseModel):
    model_config = ConfigDict(extra="allow")
    standard_time: Annotated[str, Field(min_length=1)]
    longitude: Annotated[float, Field(ge=-180, le=180)]
    gender: Literal["male", "female"]


class _BaziPillarData(BaseModel):
    model_config = ConfigDict(extra="allow")
    gan: Annotated[str, Field(min_length=1)]
    zhi: Annotated[str, Field(min_length=1)]


class _BaziModel(BaseModel):
    model_config = ConfigDict(extra="allow")
    nian: _BaziPillarData
    yue: _BaziPillarData
    ri: _BaziPillarData
    shi: _BaziPillarData


class _StarModel(BaseModel):
    model_config = ConfigDict(extra="allow")
    ming: Annotated[str, Field(min_length=1)]


class _PalaceModel(BaseModel):
    model_config = ConfigDict(extra="allow")
    ming: str
    xing_yao: list[_StarModel] = Field(default_factory=list)


class _ZiweiModel(BaseModel):
    model_config = ConfigDict(extra="allow")
    shi_er_gong: Annotated[list[_PalaceModel], Field(min_length=12, max_length=12)]


class _DaxianStepModel(BaseModel):
    model_config = ConfigDict(extra="allow")
    ming_gong: Annotated[str, Field(min_length=1)]
    start_year: int
    end_year: int
    qi_sui: int
    zhi_sui: int


class NatalChartModel(BaseModel):
    """Structural contract for build_natal_chart output."""

    model_config = ConfigDict(extra="allow")
    subject: _SubjectModel
    bazi: _BaziModel
    ziwei: _ZiweiModel
    ziwei_daxian: Annotated[list[_DaxianStepModel], Field(min_length=DAXIAN_COUNT, max_length=DAXIAN_COUNT)]
    natal_chart_digest: str

    @model_validator(mode="after")
    def _check_domain_invariants(self) -> "NatalChartModel":
        actual = tuple(p.ming for p in self.ziwei.shi_er_gong)
        if actual != PALACES:
            raise ValueError(
                f"ziwei.shi_er_gong 必须按 engine 宫位闭集从命宫顺布: {PALACES}。"
            )
        daxian_palaces = [s.ming_gong for s in self.ziwei_daxian]
        if len(set(daxian_palaces)) != DAXIAN_COUNT:
            raise ValueError(f"ziwei_daxian 的 {DAXIAN_COUNT} 个宫位必须唯一。")
        return self


_adapter: TypeAdapter[NatalChartModel] = TypeAdapter(NatalChartModel)


def validate_natal_chart(natal_chart: object, action: str = "natal_chart") -> None:
    """只接受 build_natal_chart 的 canonical subject + bazi + ziwei 结构。"""
    if not isinstance(natal_chart, dict):
        raise ChartSchemaError(f"{action} natal_chart 必须是 build_natal_chart 返回的 object。")
    missing = [key for key in REQUIRED_FIELDS if key not in natal_chart]
    if missing:
        raise ChartSchemaError(f"{action} natal_chart 缺少字段: {', '.join(missing)}。")
    try:
        _adapter.validate_python(natal_chart)
    except ValidationError as exc:
        raise ChartSchemaError(f"{action} natal_chart 结构校验失败: {exc}") from exc
    validate_natal_chart_digest(natal_chart, action=action)
