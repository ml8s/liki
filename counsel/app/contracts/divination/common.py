"""Divination tool contracts.

六爻 / 奇门契约按领域拆分；旧导入路径由包级 `__init__.py` 保持兼容。
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field


class DivinationConflict(BaseModel):
    """引擎判定的并列/矛盾信号；解读必须显式处理，不得静默择一。"""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)] = Field(description="并列或矛盾信号 ID。")
    reason: Annotated[str, Field(min_length=1)] = Field(description="为什么必须解释而不是静默择一。")
    facts: list[Annotated[str, Field(min_length=1)]] = Field(
        default_factory=list,
        description="构成矛盾的事实。",
    )



class DivinationSameMatterFollowupPolicy(BaseModel):
    """Allowed clarification dimensions for the same matter."""

    model_config = ConfigDict(extra="forbid")
    clarification: Literal["allowed_from_reading"] = Field(description="可基于 Reading 澄清同一事项。")
    timing: Literal["allowed_from_reading"] = Field(description="可基于 Reading 讨论应期。")
    obstacle: Literal["allowed_from_reading"] = Field(description="可基于 Reading 讨论阻碍。")


class DivinationFollowupPolicy(BaseModel):
    """Explicit follow-up boundary returned with every reading."""

    model_config = ConfigDict(extra="forbid")
    same_matter: DivinationSameMatterFollowupPolicy = Field(description="同一事项追问范围。")
    new_domain: Literal["requires_new_case"] = Field(description="新领域需新起问测。")
    another_matter: Literal["requires_new_case"] = Field(description="另一事项需新起问测。")
    recast_same_matter: Literal["forbidden"] = Field(description="同一事项禁止重起盘卦。")
    beyond_returned_evidence: Literal["answer_insufficient"] = Field(description="超出 Reading 证据时不得作答。")

