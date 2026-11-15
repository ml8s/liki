"""Liuyao-only divination contracts."""

from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator

from app.contracts.divination.common import (
    DivinationConflict,
    DivinationFollowupPolicy,
)
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


class LiuyaoMatterFocus(BaseModel):
    """按问事类别推导用神。"""
    model_config = ConfigDict(extra="forbid")
    kind: Literal["matter"] = Field(description="焦点类型：问事类别。")
    matter: Literal[
        "general", "self", "other", "career", "wealth",
        "study", "home", "travel", "family", "competition", "children",
        "lost_item", "legal",
    ] = Field(description="问事类别；决定用神。")

class LiuyaoRelationshipFocus(BaseModel):
    """按问卦人性别视角推导感情用神。"""
    model_config = ConfigDict(extra="forbid")
    kind: Literal["matter"] = Field(description="焦点类型：问事类别。")
    matter: Literal["relationship"] = Field(description="问事类别。")
    perspective: Literal["male", "female"] = Field(
        description="问卦人性别视角；matter=relationship 必填。",
    )

class LiuyaoYongshenFocus(BaseModel):
    """使用显式六亲用神。"""
    model_config = ConfigDict(extra="forbid")
    kind: Literal["yong_shen"] = Field(description="焦点类型：显式用神。")
    yong_shen: Literal["父母", "兄弟", "官鬼", "妻财", "子孙", "世爻", "应爻"] = Field(
        description="高级显式用神。",
    )


LiuyaoFocus = LiuyaoMatterFocus | LiuyaoRelationshipFocus | LiuyaoYongshenFocus


class LiuyaoAutoOrigin(BaseModel):
    model_config = ConfigDict(extra="forbid")
    mode: Literal["auto"] = Field(description="安全随机起卦。")


class LiuyaoCoinsOrigin(BaseModel):
    model_config = ConfigDict(extra="forbid")
    mode: Literal["coins"] = Field(description="铜钱记录起卦。")
    rounds: Annotated[
        list[Annotated[list[Literal["正", "反"]], Field(min_length=3, max_length=3)]],
        Field(min_length=6, max_length=6),
    ] = Field(description="六组三枚正/反，顺序为初爻到上爻。")


class LiuyaoYaosOrigin(BaseModel):
    model_config = ConfigDict(extra="forbid")
    mode: Literal["yaos"] = Field(description="爻值起卦。")
    yaos: Annotated[
        list[Annotated[int, Field(ge=6, le=9)]],
        Field(min_length=6, max_length=6),
    ] = Field(description="六个爻值 6-9，初爻到上爻。")


LiuyaoOrigin = LiuyaoAutoOrigin | LiuyaoCoinsOrigin | LiuyaoYaosOrigin


def _liuyao_context_schema_extra(schema: dict) -> None:
    """Project focus discriminators and relationship perspective rules to JSON Schema."""
    matter_names = [
        "general", "self", "other", "career", "wealth",
        "study", "home", "travel", "family",
        "competition", "children", "lost_item", "legal",
    ]
    focus = {
        "oneOf": [
            {
                "additionalProperties": False,
                "properties": {
                    "kind": {"const": "matter", "description": "焦点类型：问事类别。", "type": "string"},
                    "matter": {
                        "description": "问事类别；决定用神。",
                        "enum": matter_names,
                        "type": "string",
                    },
                },
                "required": ["kind", "matter"],
                "type": "object",
            },
            {
                "additionalProperties": False,
                "properties": {
                    "kind": {"const": "matter", "description": "焦点类型：问事类别。", "type": "string"},
                    "matter": {"const": "relationship", "description": "问事类别。", "type": "string"},
                    "perspective": {
                        "description": "问卦人性别视角；matter=relationship 必填。",
                        "enum": ["male", "female"],
                        "type": "string",
                    },
                },
                "required": ["kind", "matter", "perspective"],
                "type": "object",
            },
            {
                "additionalProperties": False,
                "properties": {
                    "kind": {"const": "yong_shen", "description": "焦点类型：显式用神。", "type": "string"},
                    "yong_shen": {
                        "description": "高级显式用神。",
                        "enum": ["父母", "兄弟", "官鬼", "妻财", "子孙", "世爻", "应爻"],
                        "type": "string",
                    },
                },
                "required": ["kind", "yong_shen"],
                "type": "object",
            },
        ],
        "description": "占问焦点；chart 确定性解析为用神。",
    }
    schema.setdefault("properties", {})["focus"] = focus

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

class LiuyaoCoinConvention(BaseModel):
    """Engine's fixed coin scoring convention."""

    model_config = ConfigDict(extra="forbid")
    heads: Literal[3] = Field(serialization_alias="正", validation_alias="正", description="正面记 3。")
    tails: Literal[2] = Field(serialization_alias="反", validation_alias="反", description="反面记 2。")


class LiuyaoCoinsCastingReceipt(BaseModel):
    """Coins casting returned by liuyao_qigua."""

    model_config = ConfigDict(extra="forbid")
    mode: Literal["coins"] = Field(description="起卦方式。")
    order: Literal["bottom_up"] = Field(description="爻序。")
    coin_convention: LiuyaoCoinConvention | None = Field(
        default=None,
        description="铜钱记分约定。",
    )
    rounds: list[LiuyaoCastingRoundsItem] = Field(description="六组铜钱记录。")
    yaos: Annotated[list[Annotated[int, Field(ge=6, le=9)]], Field(min_length=6, max_length=6)] = Field(description="六个爻值。")
    dong_yao: DongYaoPositions = Field(description="动爻位置列表，1-6。")


class LiuyaoYaosCastingReceipt(BaseModel):
    """Direct line-value casting returned by liuyao_qigua."""

    model_config = ConfigDict(extra="forbid")
    mode: Literal["yaos"] = Field(description="起卦方式。")
    order: Literal["bottom_up"] = Field(description="爻序。")
    yaos: Annotated[list[Annotated[int, Field(ge=6, le=9)]], Field(min_length=6, max_length=6)] = Field(description="六个爻值。")
    dong_yao: DongYaoPositions = Field(description="动爻位置列表，1-6。")


LiuyaoCasting = LiuyaoCoinsCastingReceipt | LiuyaoYaosCastingReceipt


class LiuyaoCastingContext(BaseModel):
    """Immutable casting context bound at qigua time."""

    model_config = ConfigDict(extra="forbid", json_schema_extra=_liuyao_context_schema_extra)
    standard_time: Annotated[str, Field(json_schema_extra={"format": "date-time"})] = Field(
        description="起卦地标准时间 RFC3339（无 DST）。",
    )
    longitude: Annotated[float, Field(ge=-180, le=180)] = Field(
        description="起卦地经度（-180..180）。",
    )
    focus: LiuyaoFocus = Field(description="占问焦点；chart 确定性解析为用神。")


class LiuyaoCastingEvent(BaseModel):
    """Immutable question, origin, casting, and casting context."""

    model_config = ConfigDict(extra="forbid")
    question: Annotated[str, Field(min_length=2)] = Field(
        description="起卦时绑定的唯一固定问题。",
    )
    origin: LiuyaoOrigin = Field(description="起卦创建方式。")
    casting: LiuyaoCasting = Field(description="一次起卦的完整源事实。")
    casting_context: LiuyaoCastingContext = Field(
        description="起卦地标准时间、经度与用户确认的占问焦点。",
    )

    @model_validator(mode="after")
    def _origin_matches_casting(self) -> "LiuyaoCastingEvent":
        expected = "yaos" if self.origin.mode == "yaos" else "coins"
        if self.casting.mode != expected:
            raise ValueError(
                f"origin.mode={self.origin.mode} requires casting.mode={expected}",
            )
        return self


class LiuyaoAskInput(BaseModel):
    """解读一个已存在的六爻卦象事件所需的输入。"""

    model_config = ConfigDict(extra="forbid")
    event: LiuyaoCastingEvent = Field(
        description="liuyao_qigua 返回并由 chart 原样回传的完整卦象事件。",
    )
    topic: LiuyaoTopic | None = Field(
        default=None,
        description="可选专题视角；省略时与 matter 相同。",
    )


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


class LiuyaoFuShenView(BaseModel):
    """Fu-shen receipt used by a hidden yong-shen."""

    model_config = ConfigDict(extra="forbid")
    position: Annotated[int, Field(ge=1, le=6, description="伏神所在爻位。")]
    liu_qin: str = Field(description="伏神六亲。")
    zhi: str = Field(description="伏神地支。")


class LiuyaoYongShenView(BaseModel):
    """Yong-shen state published with the unified chart view."""

    model_config = ConfigDict(extra="forbid")
    name: str = Field(description="用神类别。")
    position: Annotated[int, Field(ge=0, le=6, description="用神爻位；0 表示本卦不现。")]
    is_hidden: bool = Field(default=False, description="本卦不现时是否取伏神。")
    wang_shuai: str = Field(description="用神旺衰。")
    yue_po: bool = Field(default=False, description="用神是否月破。")
    xun_kong: bool = Field(default=False, description="用神是否旬空。")
    mu_ku: bool = Field(default=False, description="用神是否入墓。")
    liu_shou: str | None = Field(default=None, description="用神临的六兽；伏神不继承飞神六兽。")
    chang_sheng: str | None = Field(default=None, description="用神五行在月支的十二长生。")
    fu_shen: LiuyaoFuShenView | None = Field(default=None, description="用神不现时选用的伏神。")
    fu_shen_candidates: list[LiuyaoFuShenView] = Field(
        default_factory=list, description="可用伏神候选。"
    )


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
    yong_shen: LiuyaoYongShenView = Field(description="最终用神状态。")
    gua_ci: LiuyaoGuaCi | None = Field(default=None, description="卦爻辞。")


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
    | LiuyaoYongShenView
    | LiuyaoHiddenLine
    | LiuyaoBranchRelationFact
    | LiuyaoDayClashFact
    | LiuyaoMovingRelation
    | LiuyaoMovingTransformation
    | LiuyaoSanHeCandidate
    | LiuyaoForceChain
    | LiuyaoFuShenView
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




class LiuyaoAnalysisTimingCandidate(BaseModel):
    """Raw engine timing candidate before policy projection."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="应期候选 ID。")]
    mechanism: str = Field(description="应期机制。")
    position: Annotated[int, Field(ge=0, le=6)] = Field(description="关联爻位；0 表示卦级候选。")
    trigger_branch: str | None = Field(default=None, description="触发地支。")
    window: str | None = Field(default=None, description="时间窗口类型。")
    basis: list[str] = Field(default_factory=list, description="确定性依据。")


class LiuyaoTombSchool(BaseModel):
    """Explicit tomb-school convention used by this chart."""

    model_config = ConfigDict(extra="forbid")
    earth_branch: str = Field(description="土爻墓库地支。")
    school: str = Field(description="墓库流派。")


class LiuyaoAnalysis(BaseModel):
    """Judgment-layer facts derived beyond the displayed chart."""

    model_config = ConfigDict(extra="forbid")
    moving_relations: list[LiuyaoMovingRelation] = Field(default_factory=list)
    hidden_lines: list[LiuyaoHiddenLine] = Field(default_factory=list)
    branch_relation_facts: list[LiuyaoBranchRelationFact] = Field(default_factory=list)
    day_clash_facts: list[LiuyaoDayClashFact] = Field(default_factory=list)
    moving_transformations: list[LiuyaoMovingTransformation] = Field(default_factory=list)
    san_he_candidates: list[LiuyaoSanHeCandidate] = Field(default_factory=list)
    timing_candidates: list[LiuyaoAnalysisTimingCandidate] = Field(default_factory=list)
    yong_shen_force_chain: LiuyaoForceChain | None
    yong_shen_candidates: list[LiuyaoYongshenCandidate] = Field(default_factory=list)
    conflicts: list[DivinationConflict] = Field(default_factory=list)
    tomb_school: LiuyaoTombSchool
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
    case_version: Literal["liuyao-case-v3"] = Field(
        description="Immutable case contract version.",
    )
    event: LiuyaoCastingEvent = Field(
        description="Immutable casting event shared by chart and ask.",
    )
    chart: LiuyaoChart = Field(
        description="Structural Liuyao chart.",
    )
    evidence: LiuyaoEvidence = Field(
        description="Evidence returned by this tool.",
    )
    analysis: LiuyaoAnalysis = Field(
        description="Judgment-layer analysis facts.",
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

class LiuyaoReadingScope(BaseModel):
    """Controlled business scope for one divination reading."""

    model_config = ConfigDict(extra="forbid")
    focus_kind: Literal["matter", "yong_shen"] = Field(description="问测焦点类型。")
    focus_value: str | None = Field(default=None, description="问测焦点值。")
    topic: LiuyaoTopic | None = Field(default=None, description="专题视角。")
    included_domains: unique_list[LiuyaoDomain] = Field(default_factory=list, description="本次 Reading 覆盖的领域。")
    excluded_domains: unique_list[LiuyaoDomain] = Field(default_factory=list, description="本次 Reading 不覆盖的领域。")
    followup_policy: DivinationFollowupPolicy = Field(description="追问范围与重新问测边界。")



class LiuyaoReadingEvidence(BaseModel):
    """Self-contained evidence returned with a Liuyao reading."""

    model_config = ConfigDict(extra="forbid")
    id: Annotated[str, Field(min_length=1, description="稳定 evidence ID。")]
    source_kind: Literal[
        "yongshen_context", "chart", "analysis", "pattern",
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
        title="六爻问测解读契约",
        json_schema_extra={
            "$id": "liki:liuyao-reading",
            "description": "六爻问测返回的受控 Reading。",
        },
    )
    """六爻问测返回的受控 Reading。"""

    technique: Literal["liuyao"] = Field(description="占测技术。")
    case_version: Literal["liuyao-case-v3"] = Field(description="问卦档案契约版本。")
    scope: LiuyaoReadingScope = Field(description="本次 Reading 的受控范围。")
    event: LiuyaoCastingEvent = Field(description="本次使用的不可变卦象事件。")
    chart: LiuyaoChart = Field(description="六爻盘面受控视图。")
    evidence: list[LiuyaoReadingEvidence] = Field(min_length=1, description="解读依据。")
    conflicts: list[DivinationConflict] = Field(description="并列或矛盾信号。")
    timing_candidates: list[LiuyaoTimingCandidate] = Field(description="应期候选。")
    condition_candidates: list[LiuyaoReadingConditionCandidate] = Field(description="条件还原候选。")
    interpretation: LiuyaoInterpretation = Field(description="结构化解读结论。")
    safety_advisory: LiuyaoSafetyAdvisory = Field(description="安全提示。")


TOOL_LIUYAO_ASK = ToolContract(
    name="liuyao_ask",
    description="解读一个已存在的六爻卦象事件。必须传入 liuyao_qigua 返回并在 chart 中原样回传的完整 event；返回专业盘面、证据、应期候选与受控解读。本方法不创建新卦，event 缺失时直接失败。需要时空方位决策时改用 qimen_ask。",
    input=LiuyaoAskInput,
    output=LiuyaoAskOutput,
    annotations={"readOnlyHint": True, "idempotentHint": True},
)
