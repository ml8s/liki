"""Divination tool contracts.

六爻 / 奇门契约按领域拆分；旧导入路径由包级 `__init__.py` 保持兼容。
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator

from app.contracts.divination.common import (
    DivinationFollowupPolicy,
)
from app.contracts.natal import NatalSubjectInput
from app.contracts.schema import ToolContract, unique_list


class QimenHourZhuanpanMethod(BaseModel):
    """时家奇门 + 转盘法；定局法可选，缺省拆补。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["hour"] = Field(description="时家奇门。")
    school: Literal["zhuanpan"] = Field(description="转盘法。")
    dingju_method: Literal["chaibu", "zhirun", "maoshan"] = Field(description="定局法。")


class QimenHourLuoshuMethod(BaseModel):
    """时家奇门 + 洛书飞盘法；定局法可选，缺省拆补。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["hour"] = Field(description="时家奇门。")
    school: Literal["luoshu_feipan"] = Field(description="洛书飞盘法。")
    dingju_method: Literal["chaibu", "zhirun"] = Field(description="定局法。")


class QimenHourMingfaMethod(BaseModel):
    """时家奇门 + 鸣法飞盘法；盘式固定定局。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["hour"] = Field(description="时家奇门。")
    school: Literal["mingfa_feipan"] = Field(description="鸣法飞盘；定局由盘式固定。")


class QimenDayZhuanpanMethod(BaseModel):
    """日家奇门 + 转盘法；定局法可选，缺省拆补。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["day"] = Field(description="日家奇门。")
    school: Literal["zhuanpan"] = Field(description="转盘法。")
    dingju_method: Literal["chaibu", "zhirun"] = Field(description="定局法。")


class QimenDayLuoshuMethod(BaseModel):
    """日家奇门 + 洛书飞盘法；定局法可选，缺省拆补。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["day"] = Field(description="日家奇门。")
    school: Literal["luoshu_feipan"] = Field(description="洛书飞盘法。")
    dingju_method: Literal["chaibu", "zhirun"] = Field(description="定局法。")


class QimenMonthMethod(BaseModel):
    """月家周期奇门；转盘或洛书飞盘。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["month"] = Field(description="月家周期奇门。")
    school: Literal["zhuanpan", "luoshu_feipan"] = Field(description="盘式流派。")


class QimenYearMethod(BaseModel):
    """年家周期奇门；转盘或洛书飞盘。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["year"] = Field(description="年家周期奇门。")
    school: Literal["zhuanpan", "luoshu_feipan"] = Field(description="盘式流派。")


class QimenQuarterTenMethod(BaseModel):
    """刻家奇门：十分钟三元刻家。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["quarter"] = Field(description="刻家奇门。")
    school: Literal["zhuanpan"] = Field(description="转盘法。")
    quarter_rule: Literal["ten_minute_sanyuan"] = Field(description="十分钟三元刻家。")


class QimenQuarterTwelveMethod(BaseModel):
    """刻家奇门：十二分钟十分局刻家。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["quarter"] = Field(description="刻家奇门。")
    school: Literal["zhuanpan"] = Field(description="转盘法。")
    quarter_rule: Literal["twelve_minute_ten_division"] = Field(description="十二分钟十分局刻家。")
    base_dingju_method: Literal["chaibu", "zhirun", "maoshan"] = Field(description="基准时家定局法。")
    dun_source: Literal["hour_branch", "solar_term"] = Field(description="分遁口径。")
    hour_boundary: Literal["late_zi", "zi_zheng"] = Field(description="时辰界。")


class QimenJinhanMethod(BaseModel):
    """日家金函玉镜专用盘。"""

    model_config = ConfigDict(extra="forbid")
    scope: Literal["day"] = Field(description="日家奇门。")
    school: Literal["jinhan_yujing"] = Field(description="金函玉镜专用盘。")


QimenMethod = (
    QimenHourZhuanpanMethod
    | QimenHourLuoshuMethod
    | QimenHourMingfaMethod
    | QimenDayZhuanpanMethod
    | QimenDayLuoshuMethod
    | QimenMonthMethod
    | QimenYearMethod
    | QimenQuarterTenMethod
    | QimenQuarterTwelveMethod
    | QimenJinhanMethod
)


class QimenMatterFocus(BaseModel):
    """按普通事象推导奇门用神。"""
    model_config = ConfigDict(extra="forbid")
    kind: Literal["matter"] = Field(description="焦点类型：普通事象。")
    matter: Literal[
        "career", "health", "hiding", "legal", "missing_person",
        "relationship", "study", "travel", "wealth",
    ] = Field(description="普通事象。")


class QimenYongshenFocus(BaseModel):
    """显式指定奇门用神符号。"""
    model_config = ConfigDict(extra="forbid")
    kind: Literal["yong_shen"] = Field(description="焦点类型：显式用神。")
    yong_shen: Annotated[unique_list[str], Field(min_length=1)] = Field(
        description="盘面符号：九星/八门/八神/天干等。",
    )


QimenFocus = QimenMatterFocus | QimenYongshenFocus

QimenDomain = Literal[
    "career", "health", "hiding", "legal", "missing_person",
    "relationship", "study", "travel", "wealth", "lost_property",
    "thief_capture", "capture_escape", "thief_profile", "timing",
    "palace_relation",
]


class QimenMatterFocusFact(QimenMatterFocus):
    """Matter focus resolved through the closed routing table."""

    name: str = Field(description="事象名称。")
    symbols: list[str] = Field(description="事象映射的用神符号。")
    basis: str = Field(description="映射依据。")


QimenResolvedFocus = QimenMatterFocusFact | QimenYongshenFocus


class QimenAskInput(BaseModel):
    """一次完成奇门排盘与受控解读所需的输入。"""

    model_config = ConfigDict(extra="forbid")
    question: str = Field(
        description="本次要判断的一个具体目标",
    )
    standard_time: Annotated[str, Field(json_schema_extra={"format": "date-time"})] = Field(
        description="起局地标准时间 RFC3339（无 DST）；来自用户明确事实或既有工具结果，节气换局按此时刻判定。",
    )
    method: QimenMethod = Field(
        description="奇门方法事实；用户未指定高级方法时按 schema 描述选择常规组合，不臆造高级参数。",
    )
    longitude: Annotated[float, Field(ge=-180, le=180)] = Field(
        description="起局地经度（-180..180）；必须显式传入，城市需先转换为经度。",
    )
    focus: QimenFocus | None = Field(
        default=None,
        description="普通事象或显式用神；quarter / specialized rule 可按规则省略。",
    )
    rule: Literal["lost_property", "thief_capture", "missing_person", "capture_escape", "thief_profile"] = Field(
        default=None,
        description="可选专占规则；由因子契约校验适用盘型",
    )
    subject: NatalSubjectInput | None = Field(
        default=None,
        description="可选命主；提供时用于年命干落宫；金函玉镜不接受。",
    )

    @model_validator(mode="after")
    def _enforce_case_context(self) -> "QimenAskInput":
        """Validate facts that are outside the cohesive qimen method object."""
        is_jinhan = isinstance(self.method, QimenJinhanMethod)
        if is_jinhan:
            forbidden = [
                name for name, value in (
                    ("focus", self.focus),
                    ("rule", self.rule), ("subject", self.subject),
                ) if value is not None
            ]
            if forbidden:
                raise ValueError(f"jinhan_yujing does not accept {', '.join(forbidden)}")
        elif self.rule is None and self.method.scope != "quarter":
            if self.focus is None:
                raise ValueError("ordinary qimen requires focus")

        if self.rule in ("lost_property", "thief_capture", "capture_escape", "thief_profile"):
            if self.focus is not None:
                raise ValueError(f"qimen rule {self.rule} does not accept matter or yong_shen")
        if self.rule == "missing_person" and (
            self.focus is None
            or getattr(self.focus, "matter", None) != "missing_person"
        ):
            raise ValueError("qimen rule missing_person requires matter=missing_person")
        return self


class QimenQiJu(BaseModel):
    """Resolved qi ju facts."""
    model_config = ConfigDict(extra="forbid")
    longitude: Annotated[float, Field(ge=-180, le=180, description="起局经度。")]
    standard_time: Annotated[
        str,
        Field(
            json_schema_extra={"format": "date-time"},
            description="起局地标准时间。",
        ),
    ]


class QimenEvidenceReference(BaseModel):
    """A curated reference into qimen factors."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="稳定 evidence ID。")]
    fact_ref: Annotated[str, Field(min_length=1, description="factors 内的事实引用。")]
    layer: Literal["context", "structure", "relation", "focus", "timing"] = Field(
        description="证据层级。",
    )


class QimenTimingCandidate(BaseModel):
    """A qimen timing candidate with a stable published ID."""

    model_config = ConfigDict(extra="allow")
    id: Annotated[str, Field(min_length=1, description="稳定应期候选 ID。")]
    ying_qi_kind: str | None = Field(default=None, description="应期机制；无明确机制时为 null。")
    palace: str | None = Field(default=None, description="候选相关宫位。")
    branch: str | None = Field(default=None, description="候选相关地支。")
    related_to: list[dict[str, str]] | None = Field(default=None, description="候选关联主体。")


class QimenSpecialEvidence(BaseModel):
    model_config = ConfigDict(extra="forbid")
    field: str = Field(description="命中的因子字段。")
    item_key: str | None = Field(default=None, description="数组因子中的取值键。")
    operator: str = Field(description="条件操作符。")
    expected: object | None = Field(default=None, description="规则期望值。")
    actual: object | None = Field(default=None, description="盘面实际值。")
    matched: object | None = Field(default=None, description="命中的候选对象。")


class QimenSpecialAssertion(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: str = Field(description="专占断语 ID。")
    name: str = Field(description="专占断语名称。")
    conclusion: str = Field(description="条件性结论。")
    basis: str = Field(description="断语依据。")
    evidence: list[QimenSpecialEvidence] = Field(description="命中的结构化证据。")


class QimenSpecialAssertions(BaseModel):
    """Special returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    rule: Literal["lost_property", "thief_capture", "missing_person", "capture_escape", "thief_profile"]
    assertions: list[QimenSpecialAssertion]


class QimenPolicy(BaseModel):
    """Policy returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    immutable: Literal[True]
    no_rechart_without_new_event: Literal[True]
    candidate_not_absolute_verdict: Literal[True]
    forbid_downscaling_obvious_big_failure: Literal[True]


class QimenSafetyAdvisory(BaseModel):
    """Safety advisory returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    status: Literal["allow", "advisory"] = Field(description="安全提示级别。")
    category: str | None = Field(default=None, description="风险类别。")
    support_first: bool = Field(default=False, description="是否优先提供支持资源。")
    message: str = Field(default="", description="安全提示内容。")
    guidance: str = Field(default="", description="处理指引。")


class QimenJudgmentAssertion(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: str = Field(description="判断候选 ID。")
    name: str = Field(description="判断名称。")
    statement: str = Field(description="传统断语原文。")
    conclusion: str = Field(description="条件性结论。")
    basis: str = Field(description="判断依据。")
    applies: bool = Field(description="该判断是否命中。")
    forbidden: list[str] = Field(description="禁止的绝对化表述。")
    conclusion_scope: str = Field(description="结论边界。")


class QimenReadingJudgmentAssertion(BaseModel):
    """Hit judgment candidate published in a Reading."""

    model_config = ConfigDict(extra="forbid")
    id: str = Field(description="判断候选 ID。")
    name: str = Field(description="判断名称。")
    statement: str = Field(description="传统断语原文。")
    conclusion: str = Field(description="条件性结论。")
    basis: str = Field(description="判断依据。")
    forbidden: list[str] = Field(description="禁止的绝对化表述。")
    conclusion_scope: str = Field(description="结论边界。")


class QimenJudgmentAssertions(BaseModel):
    """Judgments returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    results: list[QimenJudgmentAssertion]


class QimenCase(BaseModel):
    model_config = ConfigDict(extra="forbid", title="奇门 immutable case 契约", json_schema_extra={"$id": "liki:qimen-case"})
    technique: Literal["qimen"] = Field(
        description="Divination technique.",
    )
    question: Annotated[str, Field(min_length=2)] = Field(
        description="Question returned by this tool.",
    )
    qi_ju: QimenQiJu = Field(
        description="Resolved qi ju facts.",
    )
    focus: QimenResolvedFocus | None = Field(
        description="Resolved query focus.",
    )
    method: QimenMethod = Field(
        description="Concrete qimen method.",
    )
    evidence: list[QimenEvidenceReference] = Field(
        description="Curated evidence references into factors.",
    )
    timing_candidates: list[QimenTimingCandidate] = Field(
        description="Stable timing candidates referenced by qimen_ask.",
    )
    factors: dict[str, object] = Field(
        description="Qimen factors; method and timing candidates are separate.",
    )
    special: QimenSpecialAssertions | None = Field(
        description="Special returned by this tool.",
    )
    policy: QimenPolicy = Field(
        description="Policy returned by this tool.",
    )
    safety_advisory: QimenSafetyAdvisory = Field(
        description="Safety advisory returned by this tool.",
    )
    judgments: QimenJudgmentAssertions = Field(
        description="Judgments returned by this tool.",
    )


class QimenReadingScope(BaseModel):
    """Controlled business scope for one qimen reading."""

    model_config = ConfigDict(extra="forbid")
    focus_kind: Literal["matter", "yong_shen"] = Field(description="问测焦点类型。")
    focus_value: str | None = Field(default=None, description="问测焦点值。")
    rule: Literal[
        "lost_property", "thief_capture", "missing_person",
        "capture_escape", "thief_profile",
    ] | None = Field(default=None, description="专占规则；无专占时为 null。")
    included_domains: unique_list[QimenDomain] = Field(default_factory=list, description="本次 Reading 覆盖的领域。")
    excluded_domains: unique_list[QimenDomain] = Field(default_factory=list, description="本次 Reading 不覆盖的领域。")
    followup_policy: DivinationFollowupPolicy = Field(description="追问范围与重新问测边界。")



class QimenReadingEvidence(BaseModel):
    """Self-contained evidence returned with a one-shot Qimen reading."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="稳定 evidence ID。")]
    fact_ref: str = Field(description="因子或断语来源引用。")
    layer: Literal["context", "structure", "relation", "focus", "assertion"] = Field(description="证据层级。")
    fact: object = Field(description="自包含事实内容。")


class QimenInterpretation(BaseModel):
    model_config = ConfigDict(extra="forbid")
    headline: Annotated[str, Field(min_length=4, max_length=160, description="一句话解读。")]
    verdict: Annotated[str, Field(min_length=8, description="条件性判断。")]
    action: Annotated[str, Field(min_length=4, description="建议动作或核查点。")]
    boundary: Annotated[str, Field(min_length=8, description="解读边界。")]
    disclaimer: Annotated[str, Field(min_length=8, description="结果边界声明。")]


class QimenAskOutput(BaseModel):
    model_config = ConfigDict(
        extra="forbid",
        title="奇门一次性问测解读契约",
        json_schema_extra={
            "$id": "liki:qimen-reading",
            "description": "奇门一次性问测返回的受控 Reading。",
        },
    )
    """奇门一次性问测返回的受控 Reading。"""

    technique: Literal["qimen"] = Field(description="占测技术。")
    question: Annotated[str, Field(min_length=2, description="本次问测问题。")]
    scope: QimenReadingScope = Field(description="本次 Reading 的受控范围。")
    qi_ju: QimenQiJu = Field(description="起局事实。")
    focus: QimenResolvedFocus | None = Field(default=None, description="解析后的问测焦点。")
    method: QimenMethod = Field(description="所用奇门方法。")
    evidence: list[QimenReadingEvidence] = Field(min_length=1, description="解读依据。")
    timing_candidates: list[QimenTimingCandidate] = Field(description="应期候选。")
    special_assertions: list[QimenSpecialAssertion] = Field(description="专占断语候选。")
    judgment_assertions: list[QimenReadingJudgmentAssertion] = Field(description="通用判断候选。")
    interpretation: QimenInterpretation = Field(description="结构化解读结论。")
    safety_advisory: QimenSafetyAdvisory = Field(description="安全提示。")

TOOL_QIMEN_ASK = ToolContract(
    name="qimen_ask",
    description="一次完成奇门起局、因子投影和结构化解读，适用于需要时空方位信息的决策类问题（出行、谈判、寻物、择时方位等）。longitude 必填，城市需先转换；subject 与金函玉镜不兼容，非法组合在解读前失败。同一事禁止重新排局，后续追问基于本次 Reading 解释；起卦问事（得卦象而非方位盘）改用 liuyao_ask。",
    input=QimenAskInput,
    output=QimenAskOutput,
)
