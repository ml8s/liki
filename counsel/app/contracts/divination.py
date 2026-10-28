"""Divination tool contracts.

本模块是 divination 域工具契约的单一事实源：input/output schema、描述、枚举
与约束全部由此投影。
"""


from typing import Annotated, Any, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator

from app.contracts.schema import ToolContract, unique_list
from app.guards import require_exactly_one, require_mutually_exclusive


class AuditError(BaseModel):
    """One structured answer-validation error."""

    model_config = ConfigDict(extra="forbid")
    field: Annotated[str, Field(min_length=1)] = Field(description="未通过校验的字段。")
    reason: Annotated[str, Field(min_length=1)] = Field(description="失败原因。")
    ref: str | None = Field(default=None, description="涉及的具体引用（如 snapshot:…）。")


# ── liuyao_snapshot ──
TOOL_LIUYAO_SNAPSHOT_DESCRIPTION = "创建六爻 immutable snapshot：起卦（casting.mode 选 auto 随机起卦 / coins 传六组铜钱记录 / yaos 传六个爻值）、排盘、用神因子、专题、应期候选一次完成。后续追问必须复用该 snapshot。"

class LiuyaoAutoCasting(BaseModel):
    """auto：服务端安全随机三枚铜钱起六次。"""
    model_config = ConfigDict(extra="forbid")
    mode: Literal["auto"] = Field(description="安全随机起卦。")


class LiuyaoCoinsCasting(BaseModel):
    """coins：复现历史卦时传入六组三枚正/反记录。"""
    model_config = ConfigDict(extra="forbid")
    mode: Literal["coins"] = Field(description="铜钱记录起卦。")
    rounds: Annotated[
        list[Annotated[list[Literal["正", "反"]], Field(min_length=3, max_length=3)]],
        Field(min_length=6, max_length=6),
    ] = Field(description="六组三枚正/反，顺序为初爻到上爻。")


class LiuyaoYaosCasting(BaseModel):
    """yaos：直接传入六个爻值（复现已知卦）。"""
    model_config = ConfigDict(extra="forbid")
    mode: Literal["yaos"] = Field(description="爻值起卦。")
    yaos: Annotated[
        list[Annotated[int, Field(ge=6, le=9)]],
        Field(min_length=6, max_length=6),
    ] = Field(description="六个爻值 6-9，初爻到上爻。")


LiuyaoCastingInput = LiuyaoAutoCasting | LiuyaoCoinsCasting | LiuyaoYaosCasting

class LiuyaoSnapshotInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    question: Annotated[str, Field(min_length=2)] = Field(
        description="本次要判断的一个具体目标",
    )
    longitude: Annotated[float, Field(ge=-180, le=180)] = Field(
        description="问卦地点经度（-180..180）；必须显式传入，不做默认（精度要求）。",
    )
    standard_time: Annotated[str, Field(json_schema_extra={"format": "date-time"})] = Field(
        description="起卦地标准时间 RFC3339（无 DST）；复现历史卦时传入。",
    )
    casting: LiuyaoCastingInput = Field(
        description="起卦方式：auto=服务端安全随机起卦（缺省）；coins=传入六组三枚正/反铜钱记录；yaos=直接传入六个爻值。",
        default={"mode": "auto"},
    )
    matter: Literal["general", "self", "other", "career", "wealth", "relationship", "study", "home", "travel", "family", "competition", "children", "lost_item", "legal"] = Field(
        default=None,
        description="普通问事类别（决定用神）；必须与 yong_shen 二选一。与 topic 的关系：topic 省略时按 matter 推导，显式传入仅用于偏离 matter 的专题视角。",
    )
    yong_shen: Literal["父母", "兄弟", "官鬼", "妻财", "子孙", "世爻", "应爻"] = Field(
        default=None,
        description="高级显式用神；必须与 matter 二选一",
    )
    perspective: Literal["male", "female"] = Field(
        default=None,
        description="问卦人性别视角（非命主性别，不参与排盘）；matter=relationship 必填，决定感情类用神：男问取妻财，女问取官鬼。",
    )
    topic: Literal["career", "children", "competition", "family", "general", "health_context", "home", "legal", "lost_item", "other", "relationship", "self", "study", "travel", "wealth"] = Field(
        default=None,
        description="可选专题视角；省略时与 matter 相同。",
    )

    @model_validator(mode="after")
    def _enforce_matter_matrix(self) -> "LiuyaoSnapshotInput":
        """matter/yong_shen 二选一与 perspective 条件是领域规则，发布 schema
        由 casting 联合与文档表达；此处保留运行时闭环。"""
        require_exactly_one(matter=self.matter, yong_shen=self.yong_shen)
        if self.perspective is not None and self.matter != "relationship":
            raise ValueError("perspective is only valid with matter=relationship")
        if self.matter == "relationship" and self.perspective is None:
            raise ValueError("matter=relationship requires perspective")
        return self


class LiuyaoSnapshotOutputQuestion(BaseModel):
    """Question returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    text: Annotated[str, Field(min_length=2)]
    matter: str | None
    explicit_yong_shen: str | None
    perspective: str | None


class LiuyaoSnapshotOutputCastingRoundsItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    position: Annotated[int, Field(ge=1, le=6)]
    coins: Annotated[list[Literal["正", "反"]], Field(min_length=3, max_length=3)]
    value: Annotated[int, Field(ge=6, le=9)]
    label: Literal["老阴", "少阳", "少阴", "老阳"]
    changing: bool


DongYaoPositions = Annotated[
    list[Annotated[int, Field(ge=1, le=6)]],
    Field(min_length=0, max_length=6),
]


class LiuyaoSnapshotOutputCoinsCasting(BaseModel):
    """Coins casting returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    mode: Literal["coins"] = Field(description="起卦方式。")
    order: Literal["bottom_up"] = Field(description="爻序。")
    rounds: list[LiuyaoSnapshotOutputCastingRoundsItem] = Field(description="六组铜钱记录。")
    yaos: Annotated[list[Annotated[int, Field(ge=6, le=9)]], Field(min_length=6, max_length=6)] = Field(description="六个爻值。")
    dong_yao: DongYaoPositions = Field(description="动爻位置列表，1-6。")
    casting_id: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(description="canonical deterministic casting identity")


class LiuyaoSnapshotOutputYaosCasting(BaseModel):
    """Direct line-value casting returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    mode: Literal["yaos"] = Field(description="起卦方式。")
    order: Literal["bottom_up"] = Field(description="爻序。")
    yaos: Annotated[list[Annotated[int, Field(ge=6, le=9)]], Field(min_length=6, max_length=6)] = Field(description="六个爻值。")
    dong_yao: DongYaoPositions = Field(description="动爻位置列表，1-6。")
    casting_id: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(description="canonical deterministic casting identity")


LiuyaoSnapshotOutputCasting = LiuyaoSnapshotOutputCoinsCasting | LiuyaoSnapshotOutputYaosCasting


class LiuyaoSnapshotOutputBoard(BaseModel):
    """Board returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    name: str
    ben_gua: str
    bian_gua: str | None = Field(
        default=None,
    )
    palace: str
    palace_wuxing: str = Field(
        default=None,
    )
    lines: Annotated[list[Any], Field(min_length=6, max_length=6)]
    bian_lines: Annotated[list[Any], Field(max_length=6)]
    ri_gan: str
    ri_zhi: str
    yue_gan: str
    yue_zhi: str
    xun_kong: Annotated[list[Any], Field(min_length=2, max_length=2)]


class LiuyaoSnapshotOutputFocus(BaseModel):
    """Focus returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    yong_shen: dict[str, Any]
    yong_line: dict[str, Any] | None = Field(
        default=None,
    )
    world: dict[str, Any] | None = Field(
        default=None,
    )
    other: dict[str, Any] | None = Field(
        default=None,
    )


class DivinationConflict(BaseModel):
    """引擎判定的并列/矛盾信号；解读必须显式处理，不得静默择一。"""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)] = Field(description="并列或矛盾信号 ID。")
    reason: Annotated[str, Field(min_length=1)] = Field(description="为什么必须解释而不是静默择一。")
    facts: list[Annotated[str, Field(min_length=1)]] = Field(
        default_factory=list,
        description="构成矛盾的事实。",
    )


class LiuyaoSnapshotOutputEvidence(BaseModel):
    """Evidence returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    primary: list[Any]
    secondary: list[Any]
    reference: list[Any]
    conflicts: list[DivinationConflict]
    ignored_scope: list[Any]


class LiuyaoSnapshotOutputFacts(BaseModel):
    """Facts returned by this tool."""
    model_config = ConfigDict(extra="allow")
    hidden_lines: Any
    branch_relation_facts: Any
    day_clash_facts: Any
    moving_transformations: Any
    san_he_candidates: Any
    force_chain: Any
    yong_shen_candidates: Any


class LiuyaoSnapshotOutputPolicy(BaseModel):
    """Policy returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    may_output: Annotated[list[str], Field(min_length=1)] = Field(description="允许输出的结论字段。")
    must_cover_conflicts: bool
    forbidden: Annotated[list[str], Field(min_length=0)] = Field(description="禁止输出的结论类型。")
    immutable: Literal[True]
    no_recast_without_new_event: Literal[True]


class LiuyaoSnapshotOutputTimingPlanCandidatesItem(BaseModel):
    """One ranked timing candidate; engine fact fields flow through extra."""
    model_config = ConfigDict(extra="allow")
    id: Annotated[str, Field(min_length=1)] = Field(description="应期候选 id（timing 前缀引用）。")
    mechanism: str | None = Field(default=None, description="应期机制。")
    priority: Literal["high", "medium", "low"] = Field(description="上下文优先级，不是概率。")
    rank_score: int = Field(description="排序得分。")
    rank: int = Field(ge=1, description="排序序号。")
    horizon: str = Field(description="适用时间尺度。")
    conclusion_scope: Literal["conditional_timing_candidate_not_date"] = Field(
        description="结论边界：条件性应期候选，不是确定日期。",
    )
    required_check: str = Field(description="采信前必须复核的条件。")


class LiuyaoSnapshotOutputTimingPlanPolicy(BaseModel):
    model_config = ConfigDict(extra="forbid")
    may_output: list[str]
    forbidden: list[str]
    advice_scope: Literal["traditional_reference_not_professional_advice"]


class LiuyaoSnapshotOutputTimingPlan(BaseModel):
    """Timing plan returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    blocked: Literal[False]
    topic: str | None
    candidates: list[LiuyaoSnapshotOutputTimingPlanCandidatesItem]
    boundary: Literal["传统六爻应期候选；不构成医疗、法律、财务或其他专业建议。"]
    policy: LiuyaoSnapshotOutputTimingPlanPolicy


class LiuyaoSnapshotTimingCandidate(BaseModel):
    """One engine timing candidate reference; engine fact fields flow through extra."""

    model_config = ConfigDict(extra="allow")
    id: Annotated[str, Field(min_length=1)] = Field(description="应期候选 id（timing 前缀引用）。")
    mechanism: str | None = Field(default=None, description="应期机制。")


class LiuyaoSnapshotOutputSafetyAdvisory(BaseModel):
    """Safety advisory returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    status: Literal["allow", "advisory"]
    blocking: Literal[False]
    category: str | None
    severity: Literal["none", "normal", "high"]
    support_first: bool
    message: str
    guidance: str


class LiuyaoSnapshotOutput(BaseModel):
    model_config = ConfigDict(extra="forbid", title="六爻 immutable snapshot 契约", json_schema_extra={"$id": "liki:liuyao-snapshot"})
    question: LiuyaoSnapshotOutputQuestion = Field(
        description="Question returned by this tool.",
    )
    casting: LiuyaoSnapshotOutputCasting = Field(
        description="Casting returned by this tool.",
    )
    board: LiuyaoSnapshotOutputBoard = Field(
        description="Board returned by this tool.",
    )
    focus: LiuyaoSnapshotOutputFocus = Field(
        description="Focus returned by this tool.",
    )
    evidence: LiuyaoSnapshotOutputEvidence = Field(
        description="Evidence returned by this tool.",
    )
    facts: LiuyaoSnapshotOutputFacts = Field(
        description="Facts returned by this tool.",
    )
    timing_candidates: list[LiuyaoSnapshotTimingCandidate] = Field(
        description="Timing candidates returned by this tool.",
    )
    policy: LiuyaoSnapshotOutputPolicy = Field(
        description="Policy returned by this tool.",
    )
    method: Literal["liuyao"] = Field(
        description="Method returned by this tool.",
    )
    snapshot_digest: Annotated[str, Field(min_length=16)] = Field(
        description="Snapshot digest returned by this tool.",
    )
    matter: str | None = Field(
        description="Matter returned by this tool.",
    )
    topic: str | None = Field(
        description="Topic returned by this tool.",
    )
    topic_guidance: dict[str, Any] | None = Field(
        description="Topic guidance returned by this tool.",
    )
    timing_plan: LiuyaoSnapshotOutputTimingPlan = Field(
        description="Timing plan returned by this tool.",
    )
    condition_rules: dict[str, Any] = Field(
        description="Condition rules returned by this tool.",
    )
    safety_advisory: LiuyaoSnapshotOutputSafetyAdvisory = Field(
        description="Safety advisory returned by this tool.",
    )

TOOL_LIUYAO_SNAPSHOT = ToolContract(
    name="liuyao_snapshot",
    description=TOOL_LIUYAO_SNAPSHOT_DESCRIPTION,
    input=LiuyaoSnapshotInput,
    output=LiuyaoSnapshotOutput,
)


# ── liuyao_query ──
TOOL_LIUYAO_QUERY_DESCRIPTION = "基于同一个 immutable liuyao snapshot 追问；校验 digest 后返回结构化条件性 answer，不重排。"

class LiuyaoQueryInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    snapshot: LiuyaoSnapshotOutput = Field(
        description="liuyao_snapshot 返回的完整 snapshot",
    )
    message: Annotated[str, Field(min_length=2)] = Field(
        description="本次追问内容",
    )


class LiuyaoQueryOutputAudit(BaseModel):
    """Audit returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    accepted: Literal[True]
    errors: list[AuditError]
    checked: list[str]
    unchecked: list[str]


class LiuyaoQueryOutputSafetyAdvisory(BaseModel):
    """Safety advisory returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    status: Literal["allow", "advisory"]
    blocking: Literal[False]
    category: str | None
    severity: Literal["none", "normal", "high"]
    support_first: bool
    message: str
    guidance: str


class LiuyaoQueryOutput(BaseModel):
    model_config = ConfigDict(extra="forbid", title="六爻结构化 answer 契约", json_schema_extra={"$id": "liki:liuyao-answer"})
    method: Literal["liuyao"] = Field(
        description="Method returned by this tool.",
    )
    snapshot_digest: Annotated[str, Field(min_length=16)] = Field(
        description="Snapshot digest returned by this tool.",
    )
    message_digest: Annotated[str, Field(min_length=16)] = Field(
        description="Message digest returned by this tool.",
    )
    headline: Annotated[str, Field(min_length=4, max_length=160)] = Field(
        description="Headline returned by this tool.",
    )
    verdict: Annotated[str, Field(min_length=8)] = Field(
        description="Verdict returned by this tool.",
    )
    confidence: Literal["low", "medium", "high"] = Field(
        description="Confidence returned by this tool.",
    )
    conflicts: list[DivinationConflict] = Field(
        description="Conflicts returned by this tool.",
    )
    timing_refs: list[str] = Field(
        description="Timing refs returned by this tool.",
    )
    action: Annotated[str, Field(min_length=4)] = Field(
        description="Action returned by this tool.",
    )
    boundary: Annotated[str, Field(min_length=8)] = Field(
        description="Boundary returned by this tool.",
    )
    disclaimer: Annotated[str, Field(min_length=8)] = Field(
        description="Disclaimer returned by this tool.",
    )
    audit: LiuyaoQueryOutputAudit = Field(
        description="Audit returned by this tool.",
    )
    primary_evidence_refs: Annotated[list[str], Field(min_length=1)] = Field(
        description="Primary evidence refs returned by this tool.",
    )
    secondary_evidence_refs: list[str] = Field(
        description="Secondary evidence refs returned by this tool.",
    )
    safety_advisory: LiuyaoQueryOutputSafetyAdvisory = Field(
        description="Safety advisory returned by this tool.",
    )

TOOL_LIUYAO_QUERY = ToolContract(
    name="liuyao_query",
    description=TOOL_LIUYAO_QUERY_DESCRIPTION,
    input=LiuyaoQueryInput,
    output=LiuyaoQueryOutput,
)


# ── qimen_snapshot ──
TOOL_QIMEN_SNAPSHOT_DESCRIPTION = "创建奇门 immutable snapshot：自动补时间、换算真太阳时、排盘、因子投影和可选专占查询一次完成。"

class QimenHourZhuanpanMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["hour"] = Field(description="时家奇门。")
    school: Literal["zhuanpan"] = Field(description="转盘法。")
    dingju_method: Literal["chaibu", "zhirun", "maoshan"] = Field(description="定局法。")


class QimenHourLuoshuMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["hour"] = Field(description="时家奇门。")
    school: Literal["luoshu_feipan"] = Field(description="洛书飞盘法。")
    dingju_method: Literal["chaibu", "zhirun"] = Field(description="定局法。")


class QimenHourMingfaMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["hour"] = Field(description="时家奇门。")
    school: Literal["mingfa_feipan"] = Field(description="鸣法飞盘；定局由盘式固定。")


class QimenDayZhuanpanMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["day"] = Field(description="日家奇门。")
    school: Literal["zhuanpan"] = Field(description="转盘法。")
    dingju_method: Literal["chaibu", "zhirun"] = Field(description="定局法。")


class QimenDayLuoshuMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["day"] = Field(description="日家奇门。")
    school: Literal["luoshu_feipan"] = Field(description="洛书飞盘法。")
    dingju_method: Literal["chaibu", "zhirun"] = Field(description="定局法。")


class QimenMonthMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["month"] = Field(description="月家周期奇门。")
    school: Literal["zhuanpan", "luoshu_feipan"] = Field(description="盘式流派。")


class QimenYearMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["year"] = Field(description="年家周期奇门。")
    school: Literal["zhuanpan", "luoshu_feipan"] = Field(description="盘式流派。")


class QimenQuarterTenMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["quarter"] = Field(description="刻家奇门。")
    school: Literal["zhuanpan"] = Field(description="转盘法。")
    quarter_rule: Literal["ten_minute_sanyuan"] = Field(description="十分钟三元刻家。")


class QimenQuarterTwelveMethod(BaseModel):
    model_config = ConfigDict(extra="forbid")
    scope: Literal["quarter"] = Field(description="刻家奇门。")
    school: Literal["zhuanpan"] = Field(description="转盘法。")
    quarter_rule: Literal["twelve_minute_ten_division"] = Field(description="十二分钟十分局刻家。")
    base_dingju_method: Literal["chaibu", "zhirun", "maoshan"] = Field(description="基准时家定局法。")
    dun_source: Literal["hour_branch", "solar_term"] = Field(description="分遁口径。")
    hour_boundary: Literal["late_zi", "zi_zheng"] = Field(description="时辰界。")


class QimenJinhanMethod(BaseModel):
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


class QimenSnapshotInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    question: str = Field(
        description="本次要判断的一个具体目标",
    )
    standard_time: Annotated[str, Field(json_schema_extra={"format": "date-time"})] = Field(
        description="起局地标准时间 RFC3339（无 DST）；节气换局按此时刻判定。",
    )
    method: QimenMethod = Field(
        description="奇门方法事实；按 scope、school 和条件字段的合法组合传入。",
    )
    city: Annotated[str | None, Field(min_length=1)] = Field(
        default=None,
        description="起局地点；与 longitude 二选一",
    )
    longitude: Annotated[float | None, Field(ge=-180, le=180)] = Field(
        default=None,
        description="起局经度；与 city 二选一",
    )
    matter: Literal["career", "health", "hiding", "legal", "missing_person", "relationship", "study", "travel", "wealth"] = Field(
        default=None,
        description="普通事象；必须与 yong_shen 二选一",
    )
    yong_shen: Annotated[unique_list[str], Field(min_length=1)] = Field(
        default=None,
        description="高级显式用神符号列表（盘面符号：九星/八门/八神/天干等，如 天芮、开门）。奇门用神是开放符号集，与六爻的六亲闭集不同；必须与 matter 二选一。",
    )
    rule: Literal["lost_property", "thief_capture", "missing_person", "capture_escape", "thief_profile"] = Field(
        default=None,
        description="可选专占规则；由因子契约校验适用盘型",
    )
    birth_standard_time: Annotated[str, Field(json_schema_extra={"format": "date-time"})] = Field(
        default=None,
        description="出生地标准时间 RFC3339（无 DST），用于年命干落宫；金函玉镜不接受",
    )

    @model_validator(mode="after")
    def _enforce_snapshot_context(self) -> "QimenSnapshotInput":
        """Validate facts that are outside the cohesive qimen method object."""
        require_exactly_one(city=self.city, longitude=self.longitude)
        require_mutually_exclusive(matter=self.matter, yong_shen=self.yong_shen)

        is_jinhan = isinstance(self.method, QimenJinhanMethod)
        if is_jinhan:
            forbidden = [
                name for name, value in (
                    ("matter", self.matter), ("yong_shen", self.yong_shen),
                    ("rule", self.rule), ("birth_standard_time", self.birth_standard_time),
                ) if value is not None
            ]
            if forbidden:
                raise ValueError(f"jinhan_yujing does not accept {', '.join(forbidden)}")
        elif self.rule is None and self.method.scope != "quarter":
            if self.matter is None and self.yong_shen is None:
                raise ValueError("ordinary qimen requires exactly one of matter or yong_shen")

        if self.rule in ("lost_property", "thief_capture", "capture_escape", "thief_profile"):
            if self.matter is not None or self.yong_shen is not None:
                raise ValueError(f"qimen rule {self.rule} does not accept matter or yong_shen")
        if self.rule == "missing_person" and self.matter != "missing_person":
            raise ValueError("qimen rule missing_person requires matter=missing_person")
        return self


class QimenSnapshotOutputInput(BaseModel):
    """Input returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    city: str | None
    longitude: float
    standard_time: str


class QimenSnapshotOutputMethodContext(BaseModel):
    """Method context returned by this tool."""
    model_config = ConfigDict(extra="allow")
    scope: str
    school: str


class QimenSnapshotOutputSpecialAssertionsItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: str
    name: str
    conclusion: str
    basis: str
    evidence: list[Any]


class QimenSnapshotOutputSpecial(BaseModel):
    """Special returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    rule: Literal["lost_property", "thief_capture", "missing_person", "capture_escape", "thief_profile"]
    assertions: list[QimenSnapshotOutputSpecialAssertionsItem]


class QimenSnapshotOutputPolicy(BaseModel):
    """Policy returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    immutable: Literal[True]
    no_rechart_without_new_event: Literal[True]


class QimenSnapshotOutputSafetyAdvisory(BaseModel):
    """Safety advisory returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    status: Literal["allow", "advisory"]
    blocking: Literal[False]
    category: str | None
    severity: Literal["none", "normal", "high"]
    support_first: bool
    message: str
    guidance: str


class QimenSnapshotOutputJudgmentsResultsItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: str
    name: str
    statement: str
    conclusion: str
    basis: str
    applies: bool
    forbidden: list[str]
    conclusion_scope: str


class QimenSnapshotOutputJudgments(BaseModel):
    """Judgments returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    results: list[QimenSnapshotOutputJudgmentsResultsItem]
    policy: dict[str, Any]


class QimenSnapshotOutput(BaseModel):
    model_config = ConfigDict(extra="forbid", title="奇门 immutable snapshot 契约", json_schema_extra={"$id": "liki:qimen-snapshot"})
    method: Literal["qimen"] = Field(
        description="Method returned by this tool.",
    )
    snapshot_digest: Annotated[str, Field(min_length=16)] = Field(
        description="Snapshot digest returned by this tool.",
    )
    question: Annotated[str, Field(min_length=2)] = Field(
        description="Question returned by this tool.",
    )
    input: QimenSnapshotOutputInput = Field(
        description="Input returned by this tool.",
    )
    matter: dict[str, Any] | None = Field(
        description="Matter returned by this tool.",
    )
    method_context: QimenSnapshotOutputMethodContext = Field(
        description="Method context returned by this tool.",
    )
    factors: dict[str, Any] = Field(
        description="Factors returned by this tool.",
    )
    special: QimenSnapshotOutputSpecial | None = Field(
        description="Special returned by this tool.",
    )
    policy: QimenSnapshotOutputPolicy = Field(
        description="Policy returned by this tool.",
    )
    snapshot_kind: Literal["standard", "jinhan"] = Field(
        description="Snapshot kind returned by this tool.",
    )
    safety_advisory: QimenSnapshotOutputSafetyAdvisory = Field(
        description="Safety advisory returned by this tool.",
    )
    judgments: QimenSnapshotOutputJudgments = Field(
        description="Judgments returned by this tool.",
    )

TOOL_QIMEN_SNAPSHOT = ToolContract(
    name="qimen_snapshot",
    description=TOOL_QIMEN_SNAPSHOT_DESCRIPTION,
    input=QimenSnapshotInput,
    output=QimenSnapshotOutput,
)


# ── qimen_query ──
TOOL_QIMEN_QUERY_DESCRIPTION = "基于同一个 immutable qimen snapshot 追问；校验 digest 后返回结构化条件性 answer，不重排。"

class QimenQueryInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    snapshot: QimenSnapshotOutput = Field(
        description="qimen_snapshot 返回的完整 snapshot",
    )
    message: Annotated[str, Field(min_length=2)] = Field(
        description="本次追问内容",
    )


class QimenQueryOutputAudit(BaseModel):
    """Audit returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    accepted: Literal[True]
    errors: list[AuditError]
    checked: list[str]
    unchecked: list[str]


class QimenQueryOutputSafetyAdvisory(BaseModel):
    """Safety advisory returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    status: Literal["allow", "advisory"]
    blocking: Literal[False]
    category: str | None
    severity: Literal["none", "normal", "high"]
    support_first: bool
    message: str
    guidance: str


class QimenQueryOutput(BaseModel):
    model_config = ConfigDict(extra="forbid", title="奇门结构化 answer 契约", json_schema_extra={"$id": "liki:qimen-answer"})
    method: Literal["qimen"] = Field(
        description="Method returned by this tool.",
    )
    snapshot_digest: Annotated[str, Field(min_length=16)] = Field(
        description="Snapshot digest returned by this tool.",
    )
    message_digest: Annotated[str, Field(min_length=16)] = Field(
        description="Message digest returned by this tool.",
    )
    headline: Annotated[str, Field(min_length=4, max_length=160)] = Field(
        description="Headline returned by this tool.",
    )
    verdict: Annotated[str, Field(min_length=8)] = Field(
        description="Verdict returned by this tool.",
    )
    confidence: Literal["low", "medium", "high"] = Field(
        description="Confidence returned by this tool.",
    )
    conflicts: list[DivinationConflict] = Field(
        description="Conflicts returned by this tool.",
    )
    timing_refs: list[Annotated[str, Field(pattern="^timing-")]] = Field(
        description="Timing refs returned by this tool.",
    )
    action: Annotated[str, Field(min_length=4)] = Field(
        description="Action returned by this tool.",
    )
    boundary: Annotated[str, Field(min_length=8)] = Field(
        description="Boundary returned by this tool.",
    )
    disclaimer: Annotated[str, Field(min_length=8)] = Field(
        description="Disclaimer returned by this tool.",
    )
    audit: QimenQueryOutputAudit = Field(
        description="Audit returned by this tool.",
    )
    evidence_refs: Annotated[list[Annotated[str, Field(pattern="^snapshot:")]], Field(min_length=1)] = Field(
        description="Evidence refs returned by this tool.",
    )
    assertion_refs: list[Annotated[str, Field(pattern="^assertion:")]] = Field(
        description="Assertion refs returned by this tool.",
    )
    safety_advisory: QimenQueryOutputSafetyAdvisory = Field(
        description="Safety advisory returned by this tool.",
    )

TOOL_QIMEN_QUERY = ToolContract(
    name="qimen_query",
    description=TOOL_QIMEN_QUERY_DESCRIPTION,
    input=QimenQueryInput,
    output=QimenQueryOutput,
)



DIVINATION_TOOLS = (
    TOOL_LIUYAO_SNAPSHOT,
    TOOL_LIUYAO_QUERY,
    TOOL_QIMEN_SNAPSHOT,
    TOOL_QIMEN_QUERY,
)
