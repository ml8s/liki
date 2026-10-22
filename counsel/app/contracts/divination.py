"""Divination tool contracts.

Generated once from the legacy hand-written manifests, then maintained as
ordinary source: this module is the single source of truth for these tool
contracts (input schema, output schema, descriptions, enums, constraints).
"""

from __future__ import annotations

from typing import Annotated, Any, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator

from app.contracts.schema import ToolContract, unique_list

# ── liuyao_snapshot ──
TOOL_LIUYAO_SNAPSHOT_DESCRIPTION = "创建六爻 immutable snapshot：auto/coins/yaos 起卦、排盘、用神因子、专题、应期候选一次完成。后续追问必须复用该 snapshot。"

class LiuyaoSnapshotInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    question: str = Field(
        description="本次要判断的一个具体目标",
    )
    longitude: Annotated[float, Field(ge=-180, le=180)] = Field(
        description="问卦地点经度（-180..180）；必须显式传入，不做默认（精度要求）。",
    )
    timezone: str = Field(
        description="起卦地 IANA 时区。",
    )
    instant: Annotated[str, Field(json_schema_extra={"format": "date-time"})] = Field(
        default=None,
        description="起卦瞬间 UTC（RFC3339，Z）。省略表示当前起卦；复现历史卦时传入。",
    )
    mode: Literal["auto", "coins", "yaos"] = Field(
        default="auto",
        description="起卦方式",
    )
    rounds: Annotated[list[Annotated[list[Literal["正", "反"]], Field(min_length=3, max_length=3)]], Field(min_length=6, max_length=6)] = Field(
        default=None,
        description="coins 模式必填；六组三枚正/反，顺序为初爻到上爻",
    )
    yaos: Annotated[list[Annotated[int, Field(ge=6, le=9)]], Field(min_length=6, max_length=6)] = Field(
        default=None,
        description="yaos 模式必填；六个爻值 6-9，初爻到上爻",
    )
    matter: Literal["general", "self", "other", "career", "wealth", "relationship", "study", "home", "travel", "family", "competition", "children", "lost_item", "legal"] = Field(
        default=None,
        description="普通问事类别；必须与 yong_shen 二选一",
    )
    yong_shen: Literal["父母", "兄弟", "官鬼", "妻财", "子孙", "世爻", "应爻"] = Field(
        default=None,
        description="高级显式用神；必须与 matter 二选一",
    )
    perspective: Literal["male", "female"] = Field(
        default=None,
        description="relationship 必填；男问取妻财，女问取官鬼",
    )
    topic: Literal["career", "children", "competition", "family", "general", "health_context", "home", "legal", "lost_item", "other", "relationship", "self", "study", "travel", "wealth"] = Field(
        default=None,
        description="可选专题；省略时按 matter 推导",
    )

    @model_validator(mode="after")
    def _enforce_casting_matrix(self) -> "LiuyaoSnapshotInput":
        """Cross-field rules the legacy schema expressed as ``allOf``."""
        if (self.matter is None) == (self.yong_shen is None):
            raise ValueError("provide exactly one of matter or yong_shen")
        if self.mode == "coins":
            if self.rounds is None:
                raise ValueError("coins mode requires rounds")
            if self.yaos is not None:
                raise ValueError("coins mode accepts rounds only")
        elif self.mode == "yaos":
            if self.yaos is None:
                raise ValueError("yaos mode requires yaos")
            if self.rounds is not None:
                raise ValueError("yaos mode accepts yaos only")
        elif self.rounds is not None or self.yaos is not None:
            raise ValueError("auto mode does not accept rounds or yaos")
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
    solar_time: str | None = Field(
        default=None,
        description="历史字段：旧快照记录的真太阳时；新快照不再生成。",
    )


class LiuyaoSnapshotOutputCastingRoundsItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    position: Annotated[int, Field(ge=1, le=6)]
    coins: Annotated[list[Literal["正", "反"]], Field(min_length=3, max_length=3)]
    value: Annotated[int, Field(ge=6, le=9)]
    label: Literal["老阴", "少阳", "少阴", "老阳"]
    changing: bool


class LiuyaoSnapshotOutputCasting(BaseModel):
    """Casting returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    mode: Literal["coins", "yaos"]
    rounds: list[LiuyaoSnapshotOutputCastingRoundsItem] = Field(
        default=None,
    )
    yaos: Annotated[list[Annotated[int, Field(ge=6, le=9)]], Field(min_length=6, max_length=6)]
    dong_yao: list[Any]
    casting_id: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(
        description="canonical deterministic casting identity",
    )
    order: Literal["bottom_up"]


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


class LiuyaoSnapshotOutputEvidence(BaseModel):
    """Evidence returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    primary: list[Any]
    secondary: list[Any]
    reference: list[Any]
    conflicts: list[Any]
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
    may_output: list[Any]
    must_cover_conflicts: bool
    forbidden: list[Any]
    immutable: Literal[True]
    no_recast_without_new_event: Literal[True]


class LiuyaoSnapshotOutputTimingPlanCandidatesItem(BaseModel):
    model_config = ConfigDict(extra="allow")
    id: Any


class LiuyaoSnapshotOutputTimingPlanPolicy(BaseModel):
    model_config = ConfigDict(extra="forbid")
    may_output: list[str]
    forbidden: list[str]
    advice_scope: Literal["traditional_reference_not_professional_advice"]


class LiuyaoSnapshotOutputTimingPlan(BaseModel):
    """Timing plan returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    schema_version: Literal["liuyao-timing-plan-v2"]
    blocked: Literal[False]
    topic: str | None
    candidates: list[LiuyaoSnapshotOutputTimingPlanCandidatesItem]
    boundary: Literal["传统六爻应期候选；不构成医疗、法律、财务或其他专业建议。"]
    policy: LiuyaoSnapshotOutputTimingPlanPolicy


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
    model_config = ConfigDict(extra="forbid", title="六爻 immutable snapshot 契约", json_schema_extra={"$id": "liki:liuyao-snapshot-v7"})
    schema_version: Literal["liuyao-snapshot-v7"] = Field(
        description="Schema version returned by this tool.",
    )
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
    timing_candidates: list[Any] = Field(
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

class LiuyaoQueryInputSnapshot(BaseModel):
    """liuyao_snapshot 返回的完整 snapshot"""
    model_config = ConfigDict(extra="allow")
    method: Literal["liuyao"]
    schema_version: Literal["liuyao-snapshot-v7"]
    snapshot_digest: str


class LiuyaoQueryInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    snapshot: LiuyaoQueryInputSnapshot = Field(
        description="liuyao_snapshot 返回的完整 snapshot",
    )
    message: Annotated[str, Field(min_length=2)] = Field(
        description="本次追问内容",
    )


class LiuyaoQueryOutputAudit(BaseModel):
    """Audit returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    accepted: Literal[True]
    errors: list[Any]
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
    model_config = ConfigDict(extra="forbid", title="六爻结构化 answer 契约", json_schema_extra={"$id": "liki:liuyao-answer-v2"})
    schema_version: Literal["liuyao-answer-v2"] = Field(
        description="Schema version returned by this tool.",
    )
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
    conflicts: list[dict[str, Any]] = Field(
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

class QimenSnapshotInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    question: str = Field(
        description="本次要判断的一个具体目标",
    )
    timezone: str = Field(
        description="起局地 IANA 时区。",
    )
    city: Annotated[str | None, Field(min_length=1)] = Field(
        default=None,
        description="起局地点；与 longitude 二选一",
    )
    longitude: Annotated[float | None, Field(ge=-180, le=180)] = Field(
        default=None,
        description="起局经度；与 city 二选一",
    )
    instant: Annotated[str, Field(json_schema_extra={"format": "date-time"})] = Field(
        default=None,
        description="起局瞬间 UTC（RFC3339，Z）；缺省使用服务端当前时间",
    )
    matter: Literal["career", "health", "hiding", "legal", "missing_person", "relationship", "study", "travel", "wealth"] = Field(
        default=None,
        description="普通事象；必须与 yong_shen 二选一",
    )
    yong_shen: Annotated[unique_list[str], Field(min_length=1)] = Field(
        default=None,
        description="高级显式用神；必须与 matter 二选一",
    )
    rule: Literal["lost_property", "thief_capture", "missing_person", "capture_escape", "thief_profile"] = Field(
        default=None,
        description="可选专占规则；由因子契约校验适用盘型",
    )
    scope: Literal["hour", "day", "quarter", "month", "year"] = Field(
        default=None,
        description="奇门范围；缺省 hour",
    )
    school: Literal["zhuanpan", "luoshu_feipan", "mingfa_feipan", "jinhan_yujing"] = Field(
        default=None,
        description="奇门流派；缺省 zhuanpan",
    )
    dingju_method: Literal["chaibu", "zhirun", "maoshan"] = Field(
        default=None,
        description="定局方法；仅支持范围可用",
    )
    quarter_rule: Literal["ten_minute_sanyuan", "twelve_minute_ten_division"] = Field(
        default=None,
        description="刻家规则；仅 scope=quarter 使用",
    )
    base_dingju_method: Literal["chaibu", "zhirun", "maoshan"] = Field(
        default=None,
        description="十二分钟十分局基准定局法",
    )
    dun_source: Literal["hour_branch", "solar_term"] = Field(
        default=None,
        description="十二分钟十分局遁甲来源",
    )
    hour_boundary: Literal["late_zi", "zi_zheng"] = Field(
        default=None,
        description="子时边界口径",
    )
    birth_date: str = Field(
        default=None,
        description="出生日期或时刻，用于年命干落宫",
    )

    @model_validator(mode="after")
    def _enforce_method_catalog(self) -> "QimenSnapshotInput":
        """Cross-field rules the legacy schema expressed as ``allOf``.

        Pydantic cannot project ``if/then`` into the published schema, so the
        equivalent matrix runs here instead. The engine catalog is the backstop.
        """
        scope = self.scope or "hour"
        school = self.school or "zhuanpan"
        advanced = ("base_dingju_method", "dun_source", "hour_boundary")

        def forbidden(*names: str) -> list[str]:
            return [name for name in names if getattr(self, name) is not None]

        if (self.city is None) == (self.longitude is None):
            raise ValueError("provide exactly one of city or longitude")
        if self.matter is not None and self.yong_shen is not None:
            raise ValueError("matter 与 yong_shen 互斥")
        if school == "jinhan_yujing":
            if scope != "day":
                raise ValueError("school=jinhan_yujing requires scope=day")
            if (bad := forbidden(
                *advanced, "quarter_rule", "dingju_method", "matter",
                "yong_shen", "rule", "birth_date",
            )):
                raise ValueError(
                    f"school=jinhan_yujing does not accept {', '.join(bad)}"
                )
        elif self.rule is None and scope != "quarter":
            if self.matter is None and self.yong_shen is None:
                raise ValueError(
                    "ordinary qimen requires exactly one of matter or yong_shen"
                )
        if self.rule in ("lost_property", "thief_capture", "capture_escape", "thief_profile"):
            if self.matter is not None or self.yong_shen is not None:
                raise ValueError(f"qimen rule {self.rule} does not accept matter or yong_shen")
        if self.rule == "missing_person" and self.matter != "missing_person":
            raise ValueError("qimen rule missing_person requires matter=missing_person")

        if school == "mingfa_feipan":
            if scope != "hour":
                raise ValueError("school=mingfa_feipan requires scope=hour")
            if self.dingju_method != "chaibu":
                raise ValueError("school=mingfa_feipan requires dingju_method=chaibu")
            if (bad := forbidden(*advanced, "quarter_rule")):
                raise ValueError(
                    f"school=mingfa_feipan does not accept {', '.join(bad)}"
                )
        if scope == "quarter":
            if self.school not in (None, "zhuanpan"):
                raise ValueError("scope=quarter requires school=zhuanpan")
            if self.quarter_rule is None:
                raise ValueError("scope=quarter requires quarter_rule")
            if self.dingju_method is not None:
                raise ValueError("scope=quarter does not accept dingju_method")
        if self.quarter_rule == "ten_minute_sanyuan":
            if (bad := forbidden(*advanced)):
                raise ValueError(
                    f"quarter_rule=ten_minute_sanyuan does not accept {', '.join(bad)}"
                )
        if self.quarter_rule == "twelve_minute_ten_division":
            if (missing := [name for name in advanced if getattr(self, name) is None]):
                raise ValueError(
                    "quarter_rule=twelve_minute_ten_division requires "
                    f"{', '.join(missing)}"
                )
        if scope in ("month", "year"):
            if (bad := forbidden(*advanced, "quarter_rule", "dingju_method")):
                raise ValueError(f"scope={scope} does not accept {', '.join(bad)}")
        if self.dingju_method == "maoshan":
            if self.scope != "hour" or self.school != "zhuanpan":
                raise ValueError("dingju_method=maoshan requires scope=hour, school=zhuanpan")
        return self


class QimenSnapshotOutputInput(BaseModel):
    """Input returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    city: str | None
    longitude: float
    instant: str
    timezone: str


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
    schema_version: str
    results: list[QimenSnapshotOutputJudgmentsResultsItem]
    policy: dict[str, Any]


class QimenSnapshotOutput(BaseModel):
    model_config = ConfigDict(extra="forbid", title="奇门 immutable snapshot 契约", json_schema_extra={"$id": "liki:qimen-snapshot-v5"})
    schema_version: Literal["qimen-snapshot-v5"] = Field(
        description="Schema version returned by this tool.",
    )
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

class QimenQueryInputSnapshot(BaseModel):
    """qimen_snapshot 返回的完整 snapshot"""
    model_config = ConfigDict(extra="allow")
    method: Literal["qimen"]
    schema_version: Literal["qimen-snapshot-v5"]
    snapshot_digest: str


class QimenQueryInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    snapshot: QimenQueryInputSnapshot = Field(
        description="qimen_snapshot 返回的完整 snapshot",
    )
    message: Annotated[str, Field(min_length=2)] = Field(
        description="本次追问内容",
    )


class QimenQueryOutputAudit(BaseModel):
    """Audit returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    accepted: Literal[True]
    errors: list[Any]
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
    model_config = ConfigDict(extra="forbid", title="奇门结构化 answer 契约", json_schema_extra={"$id": "liki:qimen-answer-v2"})
    schema_version: Literal["qimen-answer-v2"] = Field(
        description="Schema version returned by this tool.",
    )
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
    conflicts: list[dict[str, Any]] = Field(
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
