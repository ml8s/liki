"""Divination tool contracts.

本模块是 divination 域工具契约的单一事实源：input/output schema、描述、枚举
与约束全部由此投影。
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator

from app.contracts.natal import NatalSubjectInput
from app.contracts.schema import ToolContract, unique_list

LiuyaoTopic = Literal[
    "career", "children", "competition", "family", "general", "health_context",
    "home", "legal", "lost_item", "other", "relationship", "self", "study",
    "travel", "wealth",
]
LiuyaoDomain = Literal[
    "general", "self", "other_person", "career", "wealth", "relationship",
    "study", "home", "travel", "family", "competition", "children",
    "lost_item", "legal", "obstacle", "timing",
]


# ── liuyao ask inputs ──
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

class LiuyaoMatterFocus(BaseModel):
    """按问事类别推导用神。"""
    model_config = ConfigDict(extra="forbid")
    kind: Literal["matter"] = Field(description="焦点类型：问事类别。")
    matter: Literal[
        "general", "self", "other", "career", "wealth", "relationship",
        "study", "home", "travel", "family", "competition", "children",
        "lost_item", "legal",
    ] = Field(description="问事类别；决定用神。")
    perspective: Literal["male", "female"] | None = Field(
        default=None,
        description="问卦人性别视角；matter=relationship 必填。",
    )

    @model_validator(mode="after")
    def _enforce_perspective(self) -> "LiuyaoMatterFocus":
        if self.matter == "relationship" and self.perspective is None:
            raise ValueError("matter=relationship requires perspective")
        if self.perspective is not None and self.matter != "relationship":
            raise ValueError("perspective is only valid with matter=relationship")
        return self


class LiuyaoYongshenFocus(BaseModel):
    """使用显式六亲用神。"""
    model_config = ConfigDict(extra="forbid")
    kind: Literal["yong_shen"] = Field(description="焦点类型：显式用神。")
    yong_shen: Literal["父母", "兄弟", "官鬼", "妻财", "子孙", "世爻", "应爻"] = Field(
        description="高级显式用神。",
    )


LiuyaoFocus = LiuyaoMatterFocus | LiuyaoYongshenFocus

class LiuyaoAskInput(BaseModel):
    """一次完成六爻排盘与受控解读所需的输入。"""

    model_config = ConfigDict(extra="forbid")
    question: Annotated[str, Field(min_length=2)] = Field(
        description="本次要判断的一个具体目标",
    )
    longitude: Annotated[float, Field(ge=-180, le=180)] = Field(
        description="起卦地经度（-180..180）；必须显式传入，不做默认（精度要求）。",
    )
    standard_time: Annotated[str, Field(json_schema_extra={"format": "date-time"})] = Field(
        description="起卦地标准时间 RFC3339（无 DST），必填。",
    )
    casting: LiuyaoCastingInput = Field(
        description="起卦方式；按 schema 选择 auto / coins / yaos，用户手动记录按原顺序组装。",
        default={"mode": "auto"},
    )
    focus: LiuyaoFocus = Field(
        description="占问焦点：matter 自动推导用神，或显式指定六亲用神。",
    )
    topic: LiuyaoTopic = Field(
        default=None,
        description="可选专题视角；省略时与 matter 相同。",
    )

class LiuyaoCastingRoundsItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    position: Annotated[int, Field(ge=1, le=6, description="铜钱记录爻位。")]
    coins: Annotated[list[Literal["正", "反"]], Field(min_length=3, max_length=3, description="三枚铜钱正反面。")]
    yao_value: Annotated[int, Field(ge=6, le=9, description="本爻爻值。")]
    yao_label: Literal["老阴", "少阳", "少阴", "老阳"] = Field(description="爻值名称。")
    is_dong_yao: bool = Field(description="是否动爻。")


DongYaoPositions = Annotated[
    list[Annotated[int, Field(ge=1, le=6)]],
    Field(min_length=0, max_length=6),
]


class LiuyaoCoinsCasting(BaseModel):
    """Coins casting returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    mode: Literal["coins"] = Field(description="起卦方式。")
    order: Literal["bottom_up"] = Field(description="爻序。")
    coin_convention: dict[Literal["正", "反"], int] | None = Field(
        default=None,
        description="铜钱记分约定；auto 实际落为 coins 时返回。",
    )
    rounds: list[LiuyaoCastingRoundsItem] = Field(description="六组铜钱记录。")
    yaos: Annotated[list[Annotated[int, Field(ge=6, le=9)]], Field(min_length=6, max_length=6)] = Field(description="六个爻值。")
    dong_yao: DongYaoPositions = Field(description="动爻位置列表，1-6。")


class LiuyaoYaosCasting(BaseModel):
    """Direct line-value casting returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    mode: Literal["yaos"] = Field(description="起卦方式。")
    order: Literal["bottom_up"] = Field(description="爻序。")
    yaos: Annotated[list[Annotated[int, Field(ge=6, le=9)]], Field(min_length=6, max_length=6)] = Field(description="六个爻值。")
    dong_yao: DongYaoPositions = Field(description="动爻位置列表，1-6。")


LiuyaoCasting = LiuyaoCoinsCasting | LiuyaoYaosCasting


class LiuyaoLineFlags(BaseModel):
    """Deterministic line state flags."""

    model_config = ConfigDict(extra="forbid")
    moving: bool = Field(description="本爻是否发动。")
    yue_po: bool = Field(description="是否月破。")
    xun_kong: bool = Field(description="是否旬空。")
    mu_ku: bool = Field(description="是否入墓库。")
    mu_ku_branch: str | None = Field(default=None, description="墓库地支。")
    mu_ku_element: str | None = Field(default=None, description="墓库五行。")
    dong_sheng: bool = Field(description="动爻是否生它爻。")
    dong_ke: bool = Field(description="动爻是否克它爻。")


class LiuyaoLine(BaseModel):
    """One projected line and its deterministic derived states."""

    model_config = ConfigDict(extra="forbid")
    position: Annotated[int, Field(ge=0, le=6, description="爻位；0 表示未定位。")]
    yao_value: Annotated[int, Field(ge=0, le=9)] = Field(description="爻值；0 表示无变爻。")
    gan_zhi: str = Field(min_length=0, description="该爻干支。")
    wuxing: str = Field(description="该爻五行。")
    liu_qin: str = Field(description="六亲。")
    shi_ying: Literal["世", "应", ""] = Field(default="", description="世爻或应爻标记。")
    liu_shou: str = Field(description="六神。")
    ri_chen_relations: dict[str, list[str]] | None = Field(default=None, description="与日辰的关系。")
    wang_shuai: str | None = Field(default=None, description="旺衰状态。")
    chang_sheng_yue: str | None = Field(default=None, description="月令长生状态。")
    flags: LiuyaoLineFlags = Field(description="爻位状态标记。")


class LiuyaoPattern(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="格局 ID。")]
    pattern_kind: str = Field(description="格局类型。")
    sub_type: str = Field(description="格局子类型。")
    position: int = Field(description="关联爻位。")
    is_true: bool = Field(description="结构是否成立。")


class LiuyaoGuaCi(BaseModel):
    model_config = ConfigDict(extra="forbid")
    name: str = Field(description="卦名。")
    text: str = Field(description="卦辞原文。")
    yao_ci: list[str] = Field(description="各爻爻辞。")


class LiuyaoChart(BaseModel):
    """Structural Liuyao chart."""

    model_config = ConfigDict(extra="forbid")
    gua_name: str = Field(description="本卦名称。")
    bian_gua: str | None = Field(default=None, description="变卦名称。")
    palace: str = Field(description="卦宫。")
    palace_wuxing: str = Field(description="卦宫五行。")
    lines: Annotated[list[LiuyaoLine], Field(min_length=6, max_length=6, description="初爻到上爻。")]
    bian_lines: Annotated[list[LiuyaoLine], Field(max_length=6, description="动爻变爻。")]
    ri_gan: str = Field(description="日辰天干。")
    ri_zhi: str = Field(description="日辰地支。")
    yue_gan: str = Field(description="月建天干。")
    yue_zhi: str = Field(description="月建地支。")
    xun_kong: Annotated[list[str], Field(min_length=2, max_length=2, description="旬空地支。")]
    patterns: list[LiuyaoPattern] = Field(default_factory=list, description="盘面格局。")
    gua_ci: LiuyaoGuaCi | None = Field(default=None, description="卦爻辞。")


class LiuyaoFuShen(BaseModel):
    model_config = ConfigDict(extra="forbid")
    position: Annotated[int, Field(ge=1, le=6, description="伏神所在爻位。")]
    liu_qin: str = Field(description="伏神六亲。")
    zhi: str = Field(description="伏神地支。")


class LiuyaoSelectedYongshen(BaseModel):
    model_config = ConfigDict(extra="forbid")
    name: str = Field(description="用神六亲。")
    position: Annotated[int, Field(ge=0, le=6, description="用神爻位；0 表示伏藏未定位。")]
    is_hidden: bool = Field(default=False, description="是否伏藏。")
    wang_shuai: str = Field(description="用神旺衰。")
    yue_po: bool = Field(default=False, description="是否月破。")
    xun_kong: bool = Field(default=False, description="是否旬空。")
    mu_ku: bool = Field(default=False, description="是否入墓库。")
    liu_shou: str | None = Field(default=None, description="六神。")
    chang_sheng: str | None = Field(default=None, description="长生状态。")
    fu_shen: LiuyaoFuShen | None = Field(default=None, description="伏神来源。")
    fu_shen_candidates: list[LiuyaoFuShen] = Field(default_factory=list, description="候选伏神。")


class LiuyaoSelectedYongshenContext(BaseModel):
    """Resolved selected-yongshen context for the requested focus."""

    model_config = ConfigDict(extra="forbid")
    selected_yongshen: LiuyaoSelectedYongshen = Field(description="最终用神状态。")
    yong_line: LiuyaoLine | None = Field(default=None, description="用神所在爻。")
    world_line: LiuyaoLine | None = Field(default=None, description="世爻。")
    other_line: LiuyaoLine | None = Field(default=None, description="应爻。")


class DivinationConflict(BaseModel):
    """引擎判定的并列/矛盾信号；解读必须显式处理，不得静默择一。"""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)] = Field(description="并列或矛盾信号 ID。")
    reason: Annotated[str, Field(min_length=1)] = Field(description="为什么必须解释而不是静默择一。")
    facts: list[Annotated[str, Field(min_length=1)]] = Field(
        default_factory=list,
        description="构成矛盾的事实。",
    )


class LiuyaoHiddenLine(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    position: Annotated[int, Field(ge=1, le=6)]
    liu_qin: str
    gan: str
    zhi: str
    wuxing: str
    flying_position: Annotated[int, Field(ge=1, le=6)]
    flying_liu_qin: str
    flying_gan: str
    flying_zhi: str
    flying_wuxing: str
    fei_fu_relation: str
    xun_kong: bool = False
    yue_po: bool = False
    mu_ku: bool = False
    exit_condition: str | None = None


class LiuyaoBranchRelationFact(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    left_position: Annotated[int, Field(ge=1, le=6)]
    right_position: Annotated[int, Field(ge=1, le=6)]
    relations: list[str]
    targets: list[str] = Field(default_factory=list)
    basis: list[str] = Field(default_factory=list)


class LiuyaoDayClashFact(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    position: Annotated[int, Field(ge=1, le=6)]
    kind: str
    basis: list[str] = Field(default_factory=list)
    condition: str | None = None


class LiuyaoMovingRelation(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    position: Annotated[int, Field(ge=1, le=6)]
    relations: list[str]


class LiuyaoMovingTransformation(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    position: Annotated[int, Field(ge=1, le=6)]
    from_branch: str
    to_branch: str
    return_relation: str | None = None
    advance_retreat: str | None = None
    changed_void: bool = False
    changed_break: bool = False
    basis: list[str] = Field(default_factory=list)


class LiuyaoSanHeCandidate(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    positions: list[Annotated[int, Field(ge=1, le=6)]]
    branches: list[str]
    wuxing: str
    complete: bool
    moving_count: int
    qualification: str
    line_complete: bool
    day_present: bool
    month_present: bool
    missing: list[str] = Field(default_factory=list)
    static_count: int
    void_count: int
    break_count: int
    activated: bool
    targets: list[str] = Field(default_factory=list)
    conclusion_scope: str


class LiuyaoForceEntry(BaseModel):
    model_config = ConfigDict(extra="forbid")
    role: Literal["用神", "原神", "忌神", "仇神"]
    wuxing: str
    positions: list[Annotated[int, Field(ge=1, le=6)]] = Field(default_factory=list)
    branches: list[str] = Field(default_factory=list)
    relations: list[str] = Field(default_factory=list)
    states: list[str] = Field(default_factory=list)
    note: str | None = None


class LiuyaoForceChain(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    yong_shen: str
    yong_element: str
    position: Annotated[int, Field(ge=0, le=6)]
    is_hidden: bool
    entries: list[LiuyaoForceEntry]


class LiuyaoYongshenCandidate(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    position: Annotated[int, Field(ge=1, le=6)]
    branch: str
    is_hidden: bool
    selected: bool
    wang_shuai: str | None = None
    reason: str
    basis: list[str] = Field(default_factory=list)


class LiuyaoReferenceOnlyBranchRelation(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    reason: str
    left_position: Annotated[int, Field(ge=1, le=6)]
    right_position: Annotated[int, Field(ge=1, le=6)]
    relations: list[str]
    targets: list[str] = Field(default_factory=list)
    basis: list[str] = Field(default_factory=list)


class LiuyaoReferenceOnlySanHeCandidate(BaseModel):
    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)]
    reason: str
    positions: list[Annotated[int, Field(ge=1, le=6)]]
    branches: list[str]
    wuxing: str
    complete: bool
    moving_count: int
    qualification: str
    line_complete: bool
    day_present: bool
    month_present: bool
    missing: list[str] = Field(default_factory=list)
    static_count: int
    void_count: int
    break_count: int
    activated: bool
    targets: list[str] = Field(default_factory=list)
    conclusion_scope: str


LiuyaoReferenceOnlyFact = (
    LiuyaoReferenceOnlyBranchRelation | LiuyaoReferenceOnlySanHeCandidate
)


class LiuyaoYongshenState(BaseModel):
    model_config = ConfigDict(extra="forbid")
    position: Annotated[int, Field(ge=1, le=6)]
    liu_qin: str
    wang_shuai: str
    yue_po: bool = False
    xun_kong: bool = False
    mu_ku: bool = False
    chang_sheng: str | None = None


class LiuyaoLiuShouFact(BaseModel):
    model_config = ConfigDict(extra="forbid")
    position: Annotated[int, Field(ge=1, le=6)]
    liu_shou: str


class LiuyaoNameFact(BaseModel):
    model_config = ConfigDict(extra="forbid")
    gua_name: str


LiuyaoEvidenceFact = (
    LiuyaoLine
    | LiuyaoSelectedYongshen
    | LiuyaoHiddenLine
    | LiuyaoBranchRelationFact
    | LiuyaoDayClashFact
    | LiuyaoMovingRelation
    | LiuyaoMovingTransformation
    | LiuyaoSanHeCandidate
    | LiuyaoForceChain
    | LiuyaoFuShen
    | LiuyaoPattern
    | LiuyaoGuaCi
    | LiuyaoYongshenState
    | LiuyaoLiuShouFact
    | LiuyaoNameFact
)
class LiuyaoEvidenceReference(BaseModel):
    """A judgment-layer reference to a chart fact; fact bodies are not duplicated."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1)] = Field(description="稳定 evidence ID。")
    source_ref: Annotated[str, Field(min_length=1)] = Field(description="事实或结构来源引用。")
    fact: LiuyaoEvidenceFact = Field(description="自包含事实内容。")
    conclusion_scope: str = Field(description="结论边界。")


class LiuyaoEvidence(BaseModel):
    """Curated judgment evidence."""

    model_config = ConfigDict(extra="forbid")
    primary: list[LiuyaoEvidenceReference]
    secondary: list[LiuyaoEvidenceReference]
    reference: list[LiuyaoEvidenceReference]
    conflicts: list[DivinationConflict]




class LiuyaoChartFacts(BaseModel):
    """Canonical Liuyao chart facts."""

    model_config = ConfigDict(extra="forbid")
    moving_relations: list[LiuyaoMovingRelation]
    patterns: list[LiuyaoPattern]
    hidden_lines: list[LiuyaoHiddenLine]
    branch_relation_facts: list[LiuyaoBranchRelationFact]
    day_clash_facts: list[LiuyaoDayClashFact]
    moving_transformations: list[LiuyaoMovingTransformation]
    san_he_candidates: list[LiuyaoSanHeCandidate]
    yong_shen_force_chain: LiuyaoForceChain | None
    yong_shen_candidates: list[LiuyaoYongshenCandidate]
    reference_only_facts: list[LiuyaoReferenceOnlyFact]


class LiuyaoPolicy(BaseModel):
    """Unified case, evidence, timing, and condition guardrails."""

    model_config = ConfigDict(extra="forbid")
    immutable: Literal[True]
    no_recast_without_new_event: Literal[True]
    may_output: Annotated[list[str], Field(min_length=1)]
    must_cover_conflicts: bool
    forbidden: Annotated[list[str], Field(min_length=0)]
    timing_may_output: Annotated[list[str], Field(min_length=1)]
    timing_forbidden: Annotated[list[str], Field(min_length=0)]
    timing_advice_scope: Literal["traditional_reference_not_professional_advice"]
    must_restore_hidden_conditions: Literal[True]
    forbid_absolute_old_phrases: Literal[True]
    reference_may_not_override_primary: Literal[True]


class LiuyaoTopicGuidance(BaseModel):
    """Topic-specific judgment guidance; evidence availability lives in evidence."""

    model_config = ConfigDict(extra="forbid")
    name: str
    expected_yongshen: str = Field(description="该专题期望取的用神。")
    force_roles: dict[Literal["yuan_shen", "ji_shen", "chou_shen"], str]
    primary_factors: list[str]
    reference_factors: list[str]
    timing_focus: list[str]
    common_misjudgments: list[str]
    boundary: str
    conclusion_scope: str


class LiuyaoConditionCandidate(BaseModel):
    """A condition-restoration candidate evaluated against this case."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="条件候选 ID。")]
    statement: str = Field(description="传统断语原文。")
    restored: str = Field(description="还原后的条件表达。")
    applies: bool = Field(description="该条件候选是否适用。")
    state: Literal["strong", "weak", "conditional", "not_applicable"] = Field(
        description="当前盘面的条件状态。",
    )
    conclusion: str = Field(description="条件满足时的候选结论。")
    requires: list[str] = Field(description="采信前必须复核的条件。")
    forbidden: list[str] = Field(description="禁止的绝对化结论。")
    conclusion_scope: Literal["conditional_rule_restoration_not_absolute_verdict"] = Field(
        description="条件还原候选，不是绝对断语。",
    )


class LiuyaoReadingConditionCandidate(BaseModel):
    """Applied condition candidate published in a Reading."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="条件候选 ID。")]
    statement: str = Field(description="传统断语原文。")
    restored: str = Field(description="还原后的条件表达。")
    state: Literal["strong", "weak", "conditional"] = Field(
        description="当前盘面的条件状态。",
    )
    conclusion: str = Field(description="条件满足时的候选结论。")
    requires: list[str] = Field(description="采信前必须复核的条件。")
    forbidden: list[str] = Field(description="禁止的绝对化结论。")
    conclusion_scope: Literal["conditional_rule_restoration_not_absolute_verdict"] = Field(
        description="条件还原候选，不是绝对断语。",
    )


class LiuyaoTimingCandidate(BaseModel):
    """A projected timing candidate with a stable published ID."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="稳定应期候选 ID。")]
    mechanism: str = Field(description="应期机制。")
    position: int | None = Field(default=None, description="关联爻位。")
    trigger_branch: str | None = Field(default=None, description="触发地支。")
    window: str | None = Field(default=None, description="时间窗口类型。")
    basis: list[str] = Field(default_factory=list, description="引擎依据。")
    horizon: str = Field(description="适用时间尺度。")
    conclusion_scope: Literal["conditional_timing_candidate_not_date"] = Field(description="结论边界。")
    required_check: str = Field(description="采信前必须复核的条件。")


class LiuyaoTimingPlan(BaseModel):
    """Policy-ordered timing candidates; array order is presentation order."""

    model_config = ConfigDict(extra="forbid")
    candidates: list[LiuyaoTimingCandidate]
    boundary: Literal["传统六爻应期候选；不构成医疗、法律、财务或其他专业建议。"]


class LiuyaoSafetyAdvisory(BaseModel):
    """Safety advisory returned by this tool."""
    model_config = ConfigDict(extra="forbid")
    status: Literal["allow", "advisory"] = Field(description="安全提示级别。")
    category: str | None = Field(default=None, description="风险类别。")
    support_first: bool = Field(default=False, description="是否优先提供支持资源。")
    message: str = Field(default="", description="安全提示内容。")
    guidance: str = Field(default="", description="处理指引。")


class LiuyaoCase(BaseModel):
    model_config = ConfigDict(extra="forbid", title="六爻 immutable case 契约", json_schema_extra={"$id": "liki:liuyao-case"})
    question: Annotated[str, Field(min_length=2)] = Field(
        description="Question returned by this tool.",
    )
    focus: LiuyaoFocus = Field(
        description="Requested query focus.",
    )
    casting: LiuyaoCasting = Field(
        description="Casting returned by this tool.",
    )
    liuyao_chart: LiuyaoChart = Field(
        description="Structural Liuyao chart.",
    )
    selected_yongshen_context: LiuyaoSelectedYongshenContext = Field(
        description="Resolved yong-shen context.",
    )
    evidence: LiuyaoEvidence = Field(
        description="Evidence returned by this tool.",
    )
    chart_facts: LiuyaoChartFacts = Field(
        description="Canonical chart facts.",
    )
    policy: LiuyaoPolicy = Field(
        description="Policy returned by this tool.",
    )
    technique: Literal["liuyao"] = Field(
        description="Divination technique.",
    )
    topic: str | None = Field(
        description="Topic returned by this tool.",
    )
    topic_guidance: LiuyaoTopicGuidance | None = Field(
        description="Topic-specific judgment guidance.",
    )
    timing_plan: LiuyaoTimingPlan = Field(
        description="Policy-ordered timing candidates.",
    )
    condition_candidates: list[LiuyaoConditionCandidate] = Field(
        description="Condition-restoration candidates evaluated for this case.",
    )
    safety_advisory: LiuyaoSafetyAdvisory = Field(
        description="Safety advisory returned by this tool.",
    )


# ── qimen ask inputs ──
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


class LiuyaoReadingScope(BaseModel):
    """Controlled business scope for one divination reading."""

    model_config = ConfigDict(extra="forbid")
    focus_kind: Literal["matter", "yong_shen"] = Field(description="问测焦点类型。")
    focus_value: str | None = Field(default=None, description="问测焦点值。")
    topic: LiuyaoTopic | None = Field(default=None, description="专题视角。")
    included_domains: unique_list[LiuyaoDomain] = Field(default_factory=list, description="本次 Reading 覆盖的领域。")
    excluded_domains: unique_list[LiuyaoDomain] = Field(default_factory=list, description="本次 Reading 不覆盖的领域。")
    followup_policy: DivinationFollowupPolicy = Field(description="追问范围与重新问测边界。")


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


class LiuyaoReadingEvidence(BaseModel):
    """Self-contained evidence returned with a one-shot Liuyao reading."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="稳定 evidence ID。")]
    source_kind: Literal[
        "yongshen_context", "chart", "chart_facts", "pattern",
        "moving_relation", "timing", "condition",
    ] = Field(description="证据来源类别。")
    layer: Literal["primary", "secondary", "reference", "conflict"] = Field(description="证据层级。")
    fact: dict[str, object] = Field(description="自包含事实内容。")


class LiuyaoInterpretation(BaseModel):
    model_config = ConfigDict(extra="forbid")
    headline: Annotated[str, Field(min_length=4, max_length=160, description="一句话解读。")]
    verdict: Annotated[str, Field(min_length=8, description="条件性判断。")]
    action: Annotated[str, Field(min_length=4, description="建议动作或核查点。")]
    boundary: Annotated[str, Field(min_length=8, description="解读边界。")]
    disclaimer: Annotated[str, Field(min_length=8, description="结果边界声明。")]


class LiuyaoAskOutput(BaseModel):
    model_config = ConfigDict(
        extra="forbid",
        title="六爻一次性问测解读契约",
        json_schema_extra={
            "$id": "liki:liuyao-reading",
            "description": "六爻一次性问测返回的受控 Reading。",
        },
    )
    """六爻一次性问测返回的受控 Reading。"""

    technique: Literal["liuyao"] = Field(description="占测技术。")
    question: Annotated[str, Field(min_length=2, description="本次问测问题。")]
    scope: LiuyaoReadingScope = Field(description="本次 Reading 的受控范围。")
    casting: LiuyaoCasting = Field(description="本次使用的卦象源事实。")
    chart_view: LiuyaoChart = Field(description="六爻盘面受控视图。")
    selected_yongshen_context: LiuyaoSelectedYongshenContext = Field(description="用神上下文。")
    evidence: list[LiuyaoReadingEvidence] = Field(min_length=1, description="解读依据。")
    conflicts: list[DivinationConflict] = Field(description="并列或矛盾信号。")
    timing_candidates: list[LiuyaoTimingCandidate] = Field(description="应期候选。")
    condition_candidates: list[LiuyaoReadingConditionCandidate] = Field(description="条件还原候选。")
    interpretation: LiuyaoInterpretation = Field(description="结构化解读结论。")
    safety_advisory: LiuyaoSafetyAdvisory = Field(description="安全提示。")


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


LiuyaoAskInput = LiuyaoAskInput

TOOL_LIUYAO_ASK = ToolContract(
    name="liuyao_ask",
    description="一次完成六爻起卦、排盘和结构化解读，适用于以卦象断事的问询（吉凶、成败、得失）。casting=auto 时会新起一卦，同一事禁止自动重试或重复调用；后续追问基于本次 Reading 解释，卦象事实按 schema 原样使用。需要时空方位决策时改用 qimen_ask。",
    input=LiuyaoAskInput,
    output=LiuyaoAskOutput,
    annotations={"readOnlyHint": False, "idempotentHint": False},
)

TOOL_QIMEN_ASK = ToolContract(
    name="qimen_ask",
    description="一次完成奇门起局、因子投影和结构化解读，适用于需要时空方位信息的决策类问题（出行、谈判、寻物、择时方位等）。同一事禁止重新排局，后续追问基于本次 Reading 解释；起卦问事（得卦象而非方位盘）改用 liuyao_ask。",
    input=QimenAskInput,
    output=QimenAskOutput,
)


DIVINATION_TOOLS = (
    TOOL_LIUYAO_ASK,
    TOOL_QIMEN_ASK,
)
