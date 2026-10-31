package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/liuyao"
)

type LiuyaoCoinRoundsInput []LiuyaoCoinRoundInput

type LiuyaoCoinRoundInput struct {
	Position LiuyaoPosition      `json:"position" jsonschema:"爻位，初爻到上爻 1-6。"`
	Coins    [3]LiuyaoCoinSymbol `json:"coins" jsonschema:"三枚铜钱记录，正/反按调用顺序保留。"`
	Value    LiuyaoYaoValue      `json:"value" jsonschema:"三币合计爻值 6-9。"`
	Label    LiuyaoYaoLabel      `json:"label" jsonschema:"老阴/少阳/少阴/老阳。"`
	Changing bool                `json:"changing" jsonschema:"是否动爻。"`
}

type LiuyaoYaoValuesInput []LiuyaoYaoValue

type LiuyaoQiguaInput struct {
	Mode   LiuyaoQiguaMode       `json:"mode,omitempty" jsonschema:"起卦模式：auto 安全随机，coins 传六组铜钱，yaos 传六个爻值。"`
	Rounds LiuyaoCoinRoundsInput `json:"rounds,omitempty" jsonschema:"coins 模式：六组三枚铜钱，初爻到上爻。"`
	Yaos   LiuyaoYaoValuesInput  `json:"yaos,omitempty" jsonschema:"yaos 模式：六个爻值，初爻到上爻。"`
}

type LiuyaoQiguaAutoForm struct {
	Mode LiuyaoQiguaMode `json:"mode" jsonschema:"起卦模式；auto 使用安全随机。"`
}

type LiuyaoQiguaCoinsForm struct {
	Mode   LiuyaoQiguaMode       `json:"mode" jsonschema:"起卦模式；coins 要求六组铜钱记录。"`
	Rounds LiuyaoCoinRoundsInput `json:"rounds" jsonschema:"六组三枚铜钱，初爻到上爻。"`
}

type LiuyaoQiguaYaosForm struct {
	Mode LiuyaoQiguaMode      `json:"mode" jsonschema:"起卦模式；yaos 要求六个爻值。"`
	Yaos LiuyaoYaoValuesInput `json:"yaos" jsonschema:"六个爻值，初爻到上爻。"`
}

func liuyaoQiguaFormsSchema() (json.RawMessage, error) {
	modes := []string{"auto", "coins", "yaos"}
	forms := make([]json.RawMessage, 0, len(modes))
	for _, mode := range modes {
		var raw json.RawMessage
		var err error
		switch mode {
		case "auto":
			_, raw, err = inputSchemaFor[LiuyaoQiguaAutoForm]()
		case "coins":
			_, raw, err = inputSchemaFor[LiuyaoQiguaCoinsForm]()
		case "yaos":
			_, raw, err = inputSchemaFor[LiuyaoQiguaYaosForm]()
		}
		if err != nil {
			return nil, err
		}
		raw, err = withSchemaConst(raw, "mode", mode)
		if err != nil {
			return nil, err
		}
		forms = append(forms, raw)
	}
	return json.Marshal(map[string]any{
		"type":  "object",
		"oneOf": forms,
	})
}

type LiuyaoCoinConventionOutput struct {
	Heads LiuyaoHeadsValue `json:"正" jsonschema:"正面记 3。"`
	Tails LiuyaoTailsValue `json:"反" jsonschema:"反面记 2。"`
}

type LiuyaoCoinRoundsOutput []LiuyaoCoinRoundOutput

type LiuyaoCoinRoundOutput struct {
	Position LiuyaoPosition      `json:"position" jsonschema:"爻位，初爻到上爻 1-6。"`
	Coins    [3]LiuyaoCoinSymbol `json:"coins" jsonschema:"三枚铜钱记录，正/反按调用顺序保留。"`
	Value    LiuyaoYaoValue      `json:"value" jsonschema:"三币合计爻值 6-9。"`
	Label    LiuyaoYaoLabel      `json:"label" jsonschema:"老阴/少阳/少阴/老阳。"`
	Changing bool                `json:"changing" jsonschema:"是否动爻。"`
}

type LiuyaoYaoValuesOutput []LiuyaoYaoValue
type LiuyaoPositions []LiuyaoPosition

type LiuyaoCastingOutput struct {
	Mode           LiuyaoCastMode              `json:"mode" jsonschema:"实际使用的起卦模式。"`
	Order          LiuyaoOrderName             `json:"order" jsonschema:"爻序；当前固定为初爻到上爻。"`
	CoinConvention *LiuyaoCoinConventionOutput `json:"coin_convention,omitempty" jsonschema:"coins 模式的铜钱记分约定。"`
	Rounds         LiuyaoCoinRoundsOutput      `json:"rounds,omitempty" jsonschema:"coins 模式的六组审计记录。"`
	Yaos           LiuyaoYaoValuesOutput       `json:"yaos" jsonschema:"六个爻值，初爻到上爻。"`
	DongYao        LiuyaoPositions             `json:"dong_yao" jsonschema:"动爻位置列表，1-6。"`
	CastingID      LiuyaoCastingID             `json:"casting_id" jsonschema:"canonical SHA-256 casting digest。"`
}

type LiuyaoQiguaOutput struct {
	Casting LiuyaoCastingOutput `json:"casting" jsonschema:"完整可审计起卦收据；调用 liuyao.chart 时必须原样传回。"`
}

type LiuyaoLineOutput struct {
	Position      LiuyaoOptionalPosition `json:"position" jsonschema:"爻位；本卦为 1-6，未变化的变卦占位行为 0。"`
	Type          LiuyaoYaoValue         `json:"type" jsonschema:"爻值 6-9。"`
	Gan           LiuyaoGanName          `json:"gan" jsonschema:"纳甲天干。"`
	Zhi           LiuyaoZhiName          `json:"zhi" jsonschema:"纳甲地支。"`
	Wuxing        ElementName            `json:"wuxing" jsonschema:"爻五行。"`
	LiuQin        LiuyaoLiuQinName       `json:"liu_qin" jsonschema:"六亲。"`
	ShiYing       LiuyaoShiYing          `json:"shi_ying" jsonschema:"世、应或空。"`
	LiuShou       LiuyaoLiuShouName      `json:"liu_shou" jsonschema:"临卦六兽。"`
	YuePo         bool                   `json:"yue_po,omitempty" jsonschema:"是否月破。"`
	DongSelf      bool                   `json:"dong_self,omitempty" jsonschema:"本爻是否发动。"`
	DongSheng     bool                   `json:"dong_sheng,omitempty" jsonschema:"是否有动爻生此爻。"`
	DongKe        bool                   `json:"dong_ke,omitempty" jsonschema:"是否有动爻克此爻。"`
	XunKong       bool                   `json:"xun_kong,omitempty" jsonschema:"是否值日柱旬空。"`
	MuKu          bool                   `json:"mu_ku,omitempty" jsonschema:"是否入墓。"`
	MuKuBranch    string                 `json:"mu_ku_branch,omitempty" jsonschema:"本五行墓库地支。"`
	MuKuElement   string                 `json:"mu_ku_element,omitempty" jsonschema:"被入库的五行。"`
	MuKuTypes     []LiuyaoMuKuType       `json:"mu_ku_types,omitempty" jsonschema:"入墓来源：day/moving/transformed。"`
	ChangShengYue string                 `json:"chang_sheng_yue,omitempty" jsonschema:"本爻五行在月支的十二长生。"`
}

type LiuyaoFuShenOutput struct {
	Position LiuyaoPosition   `json:"position" jsonschema:"伏神爻位 1-6。"`
	LiuQin   LiuyaoLiuQinName `json:"liu_qin" jsonschema:"伏神六亲。"`
	Zhi      LiuyaoZhiName    `json:"zhi" jsonschema:"伏神地支。"`
}

// LiuyaoBianLineOutput preserves the legacy six-slot changed-line wire: a slot
// with no change remains an explicit zero object rather than null.
type LiuyaoBianLineOutput struct {
	Position      int      `json:"position" jsonschema:"爻位；本卦为 1-6，未变化的占位行为 0。"`
	Type          int      `json:"type" jsonschema:"爻值；未变化占位行为 0。"`
	Gan           string   `json:"gan" jsonschema:"纳甲天干；占位行为空。"`
	Zhi           string   `json:"zhi" jsonschema:"纳甲地支；占位行为空。"`
	Wuxing        string   `json:"wuxing" jsonschema:"爻五行；占位行为空。"`
	LiuQin        string   `json:"liu_qin" jsonschema:"六亲；占位行为空。"`
	ShiYing       string   `json:"shi_ying" jsonschema:"世、应或空。"`
	LiuShou       string   `json:"liu_shou" jsonschema:"临卦六兽；占位行为空。"`
	YuePo         bool     `json:"yue_po,omitempty" jsonschema:"是否月破。"`
	DongSelf      bool     `json:"dong_self,omitempty" jsonschema:"本爻是否发动。"`
	DongSheng     bool     `json:"dong_sheng,omitempty" jsonschema:"是否有动爻生此爻。"`
	DongKe        bool     `json:"dong_ke,omitempty" jsonschema:"是否有动爻克此爻。"`
	XunKong       bool     `json:"xun_kong,omitempty" jsonschema:"是否值日柱旬空。"`
	MuKu          bool     `json:"mu_ku,omitempty" jsonschema:"是否入墓。"`
	MuKuBranch    string   `json:"mu_ku_branch,omitempty" jsonschema:"本五行墓库地支。"`
	MuKuElement   string   `json:"mu_ku_element,omitempty" jsonschema:"被入库的五行。"`
	MuKuTypes     []string `json:"mu_ku_types,omitempty" jsonschema:"入墓来源：day/moving/transformed。"`
	ChangShengYue string   `json:"chang_sheng_yue,omitempty" jsonschema:"本爻五行在月支的十二长生。"`
}

type LiuyaoYongShenOutput struct {
	Name             LiuyaoYongShenName     `json:"name" jsonschema:"用神类别：六亲或世爻/应爻。"`
	Position         LiuyaoOptionalPosition `json:"position" jsonschema:"用神爻位；0 表示本卦不现。"`
	IsHidden         bool                   `json:"is_hidden,omitempty" jsonschema:"本卦不现时是否取伏神。"`
	WangShuai        LiuyaoWangShuai        `json:"wang_shuai" jsonschema:"用神旺衰。"`
	YuePo            bool                   `json:"yue_po,omitempty" jsonschema:"用神是否月破。"`
	XunKong          bool                   `json:"xun_kong,omitempty" jsonschema:"用神是否旬空。"`
	MuKu             bool                   `json:"mu_ku,omitempty" jsonschema:"用神是否入墓。"`
	LiuShou          LiuyaoLiuShouName      `json:"liu_shou,omitempty" jsonschema:"用神临的六兽；伏神不继承飞神六兽。"`
	ChangSheng       string                 `json:"chang_sheng,omitempty" jsonschema:"用神五行在月支的十二长生。"`
	FuShen           *LiuyaoFuShenOutput    `json:"fu_shen,omitempty" jsonschema:"用神不现时选用的伏神。"`
	FuShenCandidates []LiuyaoFuShenOutput   `json:"fu_shen_candidates,omitempty" jsonschema:"可用伏神候选。"`
}

type LiuyaoGuaCiOutput struct {
	Name  LiuyaoHexagramName `json:"name" jsonschema:"卦名。"`
	GuaCi string             `json:"gua_ci" jsonschema:"卦辞。"`
	YaoCi [6]string          `json:"yao_ci" jsonschema:"六爻爻辞，初爻到上爻。"`
}

type LiuyaoDayRelationOutput struct {
	Relations []string `json:"relations" jsonschema:"日建与该爻的关系：冲/合/生/扶/克/平；冲合可与其他关系并存。"`
}

type LiuyaoTimingOutput struct {
	ID            LiuyaoTimingID         `json:"id" jsonschema:"应期候选 ID。"`
	Mechanism     string                 `json:"mechanism" jsonschema:"应期机制。"`
	Position      LiuyaoOptionalPosition `json:"position,omitempty" jsonschema:"相关爻位；0 表示卦级候选。"`
	TriggerBranch LiuyaoZhiName          `json:"trigger_branch,omitempty" jsonschema:"触发地支。"`
	Window        string                 `json:"window,omitempty" jsonschema:"应期窗口。"`
	Basis         []string               `json:"basis,omitempty" jsonschema:"确定性依据。"`
}

type LiuyaoPatternOutput struct {
	Type     LiuyaoPatternType      `json:"type" jsonschema:"特殊格局类型。"`
	SubType  string                 `json:"sub_type" jsonschema:"格局子类型，如真空/假空。"`
	Position LiuyaoOptionalPosition `json:"position" jsonschema:"相关爻位；0 表示全卦。"`
	IsTrue   bool                   `json:"is_true" jsonschema:"结构是否有实质效力。"`
}

type LiuyaoDongRelationOutput struct {
	Position  LiuyaoPosition           `json:"position" jsonschema:"动爻位置 1-6。"`
	Relations []LiuyaoDongRelationType `json:"relations" jsonschema:"动爻与用神/原神/忌神的关系集合。"`
}

type LiuyaoHiddenLineOutput struct {
	Position       LiuyaoPosition   `json:"position" jsonschema:"伏神爻位。"`
	LiuQin         LiuyaoLiuQinName `json:"liu_qin" jsonschema:"伏神六亲。"`
	Gan            LiuyaoGanName    `json:"gan" jsonschema:"伏神天干。"`
	Zhi            LiuyaoZhiName    `json:"zhi" jsonschema:"伏神地支。"`
	Wuxing         ElementName      `json:"wuxing" jsonschema:"伏神五行。"`
	FlyingPosition LiuyaoPosition   `json:"flying_position" jsonschema:"飞神爻位。"`
	FlyingLiuQin   LiuyaoLiuQinName `json:"flying_liu_qin" jsonschema:"飞神六亲。"`
	FlyingGan      LiuyaoGanName    `json:"flying_gan" jsonschema:"飞神天干。"`
	FlyingZhi      LiuyaoZhiName    `json:"flying_zhi" jsonschema:"飞神地支。"`
	FlyingWuxing   ElementName      `json:"flying_wuxing" jsonschema:"飞神五行。"`
	FeiFuRelation  string           `json:"fei_fu_relation" jsonschema:"飞伏关系。"`
	XunKong        bool             `json:"xun_kong,omitempty" jsonschema:"伏神是否旬空。"`
	YuePo          bool             `json:"yue_po,omitempty" jsonschema:"伏神是否月破。"`
	MuKu           bool             `json:"mu_ku,omitempty" jsonschema:"伏神是否入墓。"`
	ExitCondition  string           `json:"exit_condition,omitempty" jsonschema:"伏神引拔条件。"`
}

type LiuyaoBranchRelationOutput struct {
	LeftPosition  LiuyaoPosition `json:"left_position" jsonschema:"左侧爻位。"`
	RightPosition LiuyaoPosition `json:"right_position" jsonschema:"右侧爻位。"`
	Relations     []string       `json:"relations" jsonschema:"爻间地支关系。"`
	Targets       []string       `json:"targets,omitempty" jsonschema:"关系指向的目标。"`
	Basis         []string       `json:"basis,omitempty" jsonschema:"确定性依据。"`
}

type LiuyaoTombOutput struct {
	EarthBranch LiuyaoTombBranch `json:"earth_branch" jsonschema:"土爻墓库地支。"`
	School      LiuyaoTombSchool `json:"school" jsonschema:"墓库流派；当前为 engine_default_chen。"`
}

type LiuyaoDayClashOutput struct {
	Position  LiuyaoPosition `json:"position" jsonschema:"爻位。"`
	Kind      string         `json:"kind" jsonschema:"日冲效果类型。"`
	Basis     []string       `json:"basis,omitempty" jsonschema:"确定性依据。"`
	Condition string         `json:"condition,omitempty" jsonschema:"成立条件或限定说明。"`
}

type LiuyaoMovingTransformationOutput struct {
	Position       LiuyaoPosition `json:"position" jsonschema:"动爻位置。"`
	FromBranch     LiuyaoZhiName  `json:"from_branch" jsonschema:"动爻本变前地支。"`
	ToBranch       LiuyaoZhiName  `json:"to_branch" jsonschema:"变爻地支。"`
	ReturnRelation string         `json:"return_relation,omitempty" jsonschema:"变爻返回本爻的关系。"`
	AdvanceRetreat string         `json:"advance_retreat,omitempty" jsonschema:"进神或退神。"`
	ChangedVoid    bool           `json:"changed_void,omitempty" jsonschema:"变爻是否旬空。"`
	ChangedBreak   bool           `json:"changed_break,omitempty" jsonschema:"变爻是否月破。"`
	Basis          []string       `json:"basis,omitempty" jsonschema:"机械变化依据。"`
}

type LiuyaoSanHeOutput struct {
	Positions       []LiuyaoPosition `json:"positions" jsonschema:"参与三合候选的爻位。"`
	Branches        []LiuyaoZhiName  `json:"branches" jsonschema:"参与三合候选的地支。"`
	Element         ElementName      `json:"element" jsonschema:"三合局五行。"`
	Complete        bool             `json:"complete" jsonschema:"卦爻与日月补支后三支是否齐。"`
	MovingCount     int              `json:"moving_count" jsonschema:"候选中发动爻数。"`
	Qualification   string           `json:"qualification" jsonschema:"成局资格限定。"`
	LineComplete    bool             `json:"line_complete" jsonschema:"卦爻本身是否三支齐。"`
	DayPresent      bool             `json:"day_present" jsonschema:"日建是否补支。"`
	MonthPresent    bool             `json:"month_present" jsonschema:"月建是否补支。"`
	Missing         []LiuyaoZhiName  `json:"missing,omitempty" jsonschema:"缺失的地支。"`
	StaticCount     int              `json:"static_count" jsonschema:"候选中安静爻数。"`
	VoidCount       int              `json:"void_count" jsonschema:"候选中旬空爻数。"`
	BreakCount      int              `json:"break_count" jsonschema:"候选中月破爻数。"`
	Activated       bool             `json:"activated" jsonschema:"是否满足当前确定性成局候选条件。"`
	Targets         []string         `json:"targets,omitempty" jsonschema:"候选指向的目标。"`
	ConclusionScope string           `json:"conclusion_scope" jsonschema:"结论边界；仅结构候选，不直接断结果。"`
}

type LiuyaoForceEntryOutput struct {
	Role      LiuyaoForceRole  `json:"role" jsonschema:"力量链角色。"`
	Element   ElementName      `json:"element" jsonschema:"角色五行。"`
	Positions []LiuyaoPosition `json:"positions,omitempty" jsonschema:"同五行爻位。"`
	Branches  []LiuyaoZhiName  `json:"branches,omitempty" jsonschema:"同五行地支。"`
	Relations []string         `json:"relations,omitempty" jsonschema:"日建关系。"`
	States    []string         `json:"states,omitempty" jsonschema:"旺衰、动静、空破、入墓状态。"`
	Note      string           `json:"note,omitempty" jsonschema:"角色说明。"`
}

type LiuyaoForceChainOutput struct {
	YongShen    LiuyaoYongShenName       `json:"yong_shen" jsonschema:"用神类别。"`
	YongElement ElementName              `json:"yong_element" jsonschema:"用神五行。"`
	Position    LiuyaoOptionalPosition   `json:"position" jsonschema:"用神爻位；0 表示伏神链。"`
	IsHidden    bool                     `json:"is_hidden" jsonschema:"力量链是否基于伏神。"`
	Entries     []LiuyaoForceEntryOutput `json:"entries" jsonschema:"生扶与克制的结构链。"`
}

type LiuyaoCandidateOutput struct {
	Position  LiuyaoOptionalPosition `json:"position" jsonschema:"候选爻位；0 表示不可见候选。"`
	Branch    LiuyaoZhiName          `json:"branch" jsonschema:"候选地支。"`
	IsHidden  bool                   `json:"is_hidden" jsonschema:"是否伏神候选。"`
	Selected  bool                   `json:"selected" jsonschema:"是否当前确定性选用。"`
	WangShuai LiuyaoWangShuai        `json:"wang_shuai,omitempty" jsonschema:"候选旺衰。"`
	Reason    string                 `json:"reason" jsonschema:"选择或列出原因。"`
	Basis     []string               `json:"basis,omitempty" jsonschema:"选择依据。"`
}

type LiuyaoConflictOutput struct {
	ID     LiuyaoConflictID `json:"id" jsonschema:"并列或矛盾信号 ID。"`
	Reason string           `json:"reason" jsonschema:"为什么必须解释而不是静默择一。"`
	Facts  []string         `json:"facts,omitempty" jsonschema:"构成矛盾的事实。"`
}

type LiuyaoChartOutput struct {
	Name                  LiuyaoHexagramName                 `json:"name" jsonschema:"本卦卦名。"`
	BenGua                LiuyaoHexagramName                 `json:"ben_gua" jsonschema:"本卦卦名。"`
	BianGua               *LiuyaoHexagramName                `json:"bian_gua,omitempty" jsonschema:"变卦卦名；无动爻时省略。"`
	Gong                  string                             `json:"gong" jsonschema:"本卦宫。"`
	PalaceWuxing          ElementName                        `json:"gong_wuxing" jsonschema:"宫五行。"`
	Lines                 [6]LiuyaoLineOutput                `json:"lines" jsonschema:"本卦六爻，初爻到上爻。"`
	BianLines             [6]LiuyaoBianLineOutput            `json:"bian_yao" jsonschema:"变卦六爻；未变化的占位为全零对象。"`
	RiGan                 LiuyaoGanName                      `json:"ri_chen_gan" jsonschema:"日干。"`
	RiZhi                 LiuyaoZhiName                      `json:"ri_chen_zhi" jsonschema:"日支。"`
	YueZhi                LiuyaoZhiName                      `json:"yue_jian_zhi" jsonschema:"月建地支。"`
	YueGan                LiuyaoGanName                      `json:"yue_jian_gan" jsonschema:"月建天干。"`
	YongShen              LiuyaoYongShenOutput               `json:"yong_shen" jsonschema:"用神分析。"`
	GuaCi                 LiuyaoGuaCiOutput                  `json:"gua_ci" jsonschema:"周易卦爻辞。"`
	WangShuai             [6]LiuyaoWangShuai                 `json:"wang_shuai" jsonschema:"六爻旺衰。"`
	DayRelations          [6]LiuyaoDayRelationOutput         `json:"ri_chen_relations" jsonschema:"日建与六爻关系。"`
	XunKong               [2]LiuyaoZhiName                   `json:"xun_kong" jsonschema:"日柱旬空地支。"`
	TimingCandidates      []LiuyaoTimingOutput               `json:"timing_candidates,omitempty" jsonschema:"可审计应期候选。"`
	Patterns              []LiuyaoPatternOutput              `json:"patterns,omitempty" jsonschema:"特殊格局。"`
	DongYaoRelations      []LiuyaoDongRelationOutput         `json:"dong_yao_relations,omitempty" jsonschema:"动爻与用神/原神/忌神关系。"`
	HiddenLines           []LiuyaoHiddenLineOutput           `json:"hidden_lines,omitempty" jsonschema:"本宫全部伏神。"`
	BranchRelationFacts   []LiuyaoBranchRelationOutput       `json:"branch_relation_facts,omitempty" jsonschema:"爻间地支关系。"`
	Casting               LiuyaoCastingOutput                `json:"casting" jsonschema:"完整起卦收据。"`
	TombSchool            LiuyaoTombOutput                   `json:"tomb_school" jsonschema:"土爻墓库流派。"`
	DayClashFacts         []LiuyaoDayClashOutput             `json:"day_clash_facts,omitempty" jsonschema:"日冲事实。"`
	MovingTransformations []LiuyaoMovingTransformationOutput `json:"moving_transformations,omitempty" jsonschema:"动爻机械变化。"`
	SanHeCandidates       []LiuyaoSanHeOutput                `json:"san_he_candidates,omitempty" jsonschema:"三合结构候选。"`
	ForceChain            *LiuyaoForceChainOutput            `json:"force_chain,omitempty" jsonschema:"用神生克力量链。"`
	YongShenCandidates    []LiuyaoCandidateOutput            `json:"yong_shen_candidates,omitempty" jsonschema:"用神候选。"`
	Conflicts             []LiuyaoConflictOutput             `json:"conflicts,omitempty" jsonschema:"并列或矛盾信号。"`
}

func LiuyaoQigua(_ context.Context, input LiuyaoQiguaInput) (LiuyaoQiguaOutput, error) {
	var casting liuyao.Casting
	var err error
	switch input.Mode {
	case "", "auto":
		if input.Rounds != nil || input.Yaos != nil {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao.qigua: auto mode does not accept rounds or yaos")
		}
		casting, err = liuyao.SecureQigua()
	case "coins":
		if input.Yaos != nil {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao.qigua: coins mode accepts rounds only")
		}
		if len(input.Rounds) != 6 {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao.qigua: coins requires exactly 6 rounds, got %d", len(input.Rounds))
		}
		rounds := make([][]string, 0, 6)
		for _, round := range input.Rounds {
			coins := make([]string, 0, 3)
			for _, coin := range round.Coins {
				coins = append(coins, string(coin))
			}
			rounds = append(rounds, coins)
		}
		casting, err = liuyao.NewCoinsCasting(rounds)
	case "yaos":
		if input.Rounds != nil {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao.qigua: yaos mode accepts yaos only")
		}
		if len(input.Yaos) != 6 {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao.qigua: yaos requires exactly 6 values, got %d", len(input.Yaos))
		}
		var values [6]int
		for i, value := range input.Yaos {
			values[i] = int(value)
		}
		casting, err = liuyao.NewValuesCasting(values)
	default:
		return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao.qigua: unsupported mode %q", string(input.Mode))
	}
	if err != nil {
		return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao.qigua: %w", err)
	}
	return LiuyaoQiguaOutput{Casting: liuyaoCastingWire(casting)}, nil
}

// LiuyaoChartInput takes standard_time and longitude. The chart domain derives
// day/month from LAST and solar-term events.
type LiuyaoChartInput struct {
	StandardTime StandardTime         `json:"standard_time" jsonschema:"占问地标准时间（无 DST），RFC3339 带 offset。"`
	Longitude    float64              `json:"longitude" jsonschema:"占问地经度（-180..180）；真太阳时校正用于时辰。"`
	YongShen     LiuyaoYongShenName   `json:"yong_shen,omitempty" jsonschema:"用神；缺省为世爻。"`
	Casting      *LiuyaoCastingOutput `json:"casting" jsonschema:"liuyao.qigua 返回的完整可审计起卦收据。"`
}

func LiuyaoChart(_ context.Context, input LiuyaoChartInput) (LiuyaoChartOutput, error) {
	if input.Casting == nil {
		return LiuyaoChartOutput{}, fmt.Errorf("liuyao.chart: casting is required")
	}
	casting, err := liuyaoCastingDomain(*input.Casting)
	if err != nil {
		return LiuyaoChartOutput{}, fmt.Errorf("liuyao.chart: invalid casting: %w", err)
	}
	standard, _, err := parseStandardTime(input.StandardTime)
	if err != nil {
		return LiuyaoChartOutput{}, fmt.Errorf("liuyao.chart: %w", err)
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return LiuyaoChartOutput{}, fmt.Errorf("liuyao.chart: longitude must be within [-180,180], got %v", input.Longitude)
	}
	yongShen := string(input.YongShen)
	if yongShen == "" {
		yongShen = "世爻"
	}
	ys, err := liuyao.ParseYongShen(yongShen)
	if err != nil {
		return LiuyaoChartOutput{}, fmt.Errorf("liuyao.chart: %w", err)
	}
	chart := liuyao.ComputeChartAt(standard, input.Longitude, ys, casting.Yaos)
	chart.Casting = &casting
	return liuyaoChartWire(chart), nil
}

func liuyaoCastingDomain(output LiuyaoCastingOutput) (liuyao.Casting, error) {
	if len(output.Yaos) != 6 {
		return liuyao.Casting{}, fmt.Errorf("yaos requires exactly 6 values, got %d", len(output.Yaos))
	}
	casting := liuyao.Casting{
		Mode:      string(output.Mode),
		Order:     string(output.Order),
		Yaos:      [6]int{},
		DongYao:   make([]int, 0, len(output.DongYao)),
		CastingID: string(output.CastingID),
	}
	for i, value := range output.Yaos {
		casting.Yaos[i] = int(value)
	}
	if output.CoinConvention != nil {
		casting.CoinConvention = map[string]int{
			"正": int(output.CoinConvention.Heads),
			"反": int(output.CoinConvention.Tails),
		}
	}
	for _, round := range output.Rounds {
		coins := make([]string, 0, 3)
		for _, coin := range round.Coins {
			coins = append(coins, string(coin))
		}
		casting.Rounds = append(casting.Rounds, liuyao.CoinRound{
			Position: int(round.Position), Coins: coins, Value: int(round.Value),
			Label: string(round.Label), Changing: round.Changing,
		})
	}
	for _, position := range output.DongYao {
		casting.DongYao = append(casting.DongYao, int(position))
	}
	if err := casting.Validate(); err != nil {
		return liuyao.Casting{}, err
	}
	return casting, nil
}

func liuyaoEncodedName(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	var name string
	if err := json.Unmarshal(raw, &name); err != nil {
		return ""
	}
	return name
}

func liuyaoLineWire(line liuyao.Line) LiuyaoLineOutput {
	output := LiuyaoLineOutput{
		Position: LiuyaoOptionalPosition(line.Position),
		Type:     LiuyaoYaoValue(line.Type),
		Gan:      LiuyaoGanName(line.Gan.String()),
		Zhi:      LiuyaoZhiName(line.Zhi.String()),
		Wuxing:   ElementName(line.Wuxing.String()),
		LiuQin:   LiuyaoLiuQinName(line.LiuQin.String()),
		ShiYing:  LiuyaoShiYing(line.ShiYing),
		LiuShou:  LiuyaoLiuShouName(line.LiuShou.String()),
		YuePo:    line.YuePo, DongSelf: line.DongSelf,
		DongSheng: line.DongSheng, DongKe: line.DongKe,
		XunKong: line.XunKong, MuKu: line.MuKu,
		MuKuBranch:    line.MuKuBranch,
		MuKuElement:   line.MuKuElement,
		ChangShengYue: line.ChangShengYue,
	}
	for _, value := range line.MuKuTypes {
		output.MuKuTypes = append(output.MuKuTypes, LiuyaoMuKuType(value))
	}
	return output
}

func liuyaoFuShenWire(value liuyao.FuShen) LiuyaoFuShenOutput {
	return LiuyaoFuShenOutput{
		Position: LiuyaoPosition(value.Position),
		LiuQin:   LiuyaoLiuQinName(value.LiuQin.String()),
		Zhi:      LiuyaoZhiName(value.Zhi),
	}
}

func liuyaoChartWire(chart liuyao.Chart) LiuyaoChartOutput {
	output := LiuyaoChartOutput{
		Name:         LiuyaoHexagramName(liuyaoEncodedName(chart.BenGua)),
		BenGua:       LiuyaoHexagramName(liuyaoEncodedName(chart.BenGua)),
		Gong:         chart.Palace,
		PalaceWuxing: ElementName(chart.PalaceWuxing.String()),
		RiGan:        LiuyaoGanName(chart.RiGan.String()),
		RiZhi:        LiuyaoZhiName(chart.RiZhi.String()),
		YueZhi:       LiuyaoZhiName(chart.YueZhi.String()),
		YueGan:       LiuyaoGanName(chart.YueGan.String()),
		YongShen: LiuyaoYongShenOutput{
			Name:       LiuyaoYongShenName(chart.YongShen.Name),
			Position:   LiuyaoOptionalPosition(chart.YongShen.Position),
			IsHidden:   chart.YongShen.IsHidden,
			WangShuai:  LiuyaoWangShuai(chart.YongShen.WangShuai),
			YuePo:      chart.YongShen.YuePo,
			XunKong:    chart.YongShen.XunKong,
			MuKu:       chart.YongShen.MuKu,
			ChangSheng: chart.YongShen.ChangSheng,
		},
		GuaCi: LiuyaoGuaCiOutput{
			Name:  LiuyaoHexagramName(chart.GuaCi.Name),
			GuaCi: chart.GuaCi.GuaCi,
			YaoCi: [6]string(chart.GuaCi.YaoCi),
		},
		XunKong: [2]LiuyaoZhiName{
			LiuyaoZhiName(chart.XunKong[0].String()),
			LiuyaoZhiName(chart.XunKong[1].String()),
		},
		Casting: liuyaoCastingWire(*chart.Casting),
		TombSchool: LiuyaoTombOutput{
			EarthBranch: LiuyaoTombBranch(chart.TombSchool.EarthBranch),
			School:      LiuyaoTombSchool(chart.TombSchool.School),
		},
	}
	if int(chart.BianGua) != 0 {
		name := LiuyaoHexagramName(liuyaoEncodedName(chart.BianGua))
		output.BianGua = &name
	}
	for i, line := range chart.Lines {
		output.Lines[i] = liuyaoLineWire(line)
	}
	for i, line := range chart.BianLines {
		if line.Position != 0 {
			output.BianLines[i] = LiuyaoBianLineOutput{
				Position:      line.Position,
				Type:          int(line.Type),
				Gan:           line.Gan.String(),
				Zhi:           line.Zhi.String(),
				Wuxing:        line.Wuxing.String(),
				LiuQin:        line.LiuQin.String(),
				ShiYing:       line.ShiYing,
				LiuShou:       line.LiuShou.String(),
				YuePo:         line.YuePo,
				DongSelf:      line.DongSelf,
				DongSheng:     line.DongSheng,
				DongKe:        line.DongKe,
				XunKong:       line.XunKong,
				MuKu:          line.MuKu,
				MuKuBranch:    line.MuKuBranch,
				MuKuElement:   line.MuKuElement,
				ChangShengYue: line.ChangShengYue,
			}
			output.BianLines[i].MuKuTypes = append(output.BianLines[i].MuKuTypes, line.MuKuTypes...)
		}
	}
	if chart.YongShen.FuShen != nil {
		fuShen := liuyaoFuShenWire(*chart.YongShen.FuShen)
		output.YongShen.FuShen = &fuShen
	}
	for _, value := range chart.YongShen.FuShenCandidates {
		output.YongShen.FuShenCandidates = append(output.YongShen.FuShenCandidates, liuyaoFuShenWire(value))
	}
	for i, value := range chart.WangShuai {
		output.WangShuai[i] = LiuyaoWangShuai(value.String())
	}
	for i, value := range chart.DayRelations {
		output.DayRelations[i] = LiuyaoDayRelationOutput{Relations: append([]string(nil), value.Relations...)}
	}
	for _, value := range chart.TimingCandidates {
		output.TimingCandidates = append(output.TimingCandidates, LiuyaoTimingOutput{
			ID:            LiuyaoTimingID(value.ID),
			Mechanism:     value.Mechanism,
			Position:      LiuyaoOptionalPosition(value.Position),
			TriggerBranch: LiuyaoZhiName(value.TriggerBranch),
			Window:        value.Window,
			Basis:         append([]string(nil), value.Basis...),
		})
	}
	for _, value := range chart.Patterns {
		output.Patterns = append(output.Patterns, LiuyaoPatternOutput{
			Type: LiuyaoPatternType(value.Type), SubType: value.SubType,
			Position: LiuyaoOptionalPosition(value.Position), IsTrue: value.IsTrue,
		})
	}
	for _, value := range chart.DongYaoRelations {
		relations := make([]LiuyaoDongRelationType, 0, len(value.Relations))
		for _, relation := range value.Relations {
			relations = append(relations, LiuyaoDongRelationType(relation))
		}
		output.DongYaoRelations = append(output.DongYaoRelations, LiuyaoDongRelationOutput{
			Position: LiuyaoPosition(value.Position), Relations: relations,
		})
	}
	for _, value := range chart.HiddenLines {
		output.HiddenLines = append(output.HiddenLines, LiuyaoHiddenLineOutput{
			Position: LiuyaoPosition(value.Position), LiuQin: LiuyaoLiuQinName(value.LiuQin),
			Gan: LiuyaoGanName(value.Gan), Zhi: LiuyaoZhiName(value.Zhi),
			Wuxing: ElementName(value.Wuxing), FlyingPosition: LiuyaoPosition(value.FlyingPosition),
			FlyingLiuQin: LiuyaoLiuQinName(value.FlyingLiuQin), FlyingGan: LiuyaoGanName(value.FlyingGan),
			FlyingZhi: LiuyaoZhiName(value.FlyingZhi), FlyingWuxing: ElementName(value.FlyingWuxing),
			FeiFuRelation: value.FeiFuRelation, XunKong: value.XunKong, YuePo: value.YuePo,
			MuKu: value.MuKu, ExitCondition: value.ExitCondition,
		})
	}
	for _, value := range chart.BranchRelationFacts {
		relations := append([]string(nil), value.Relations...)
		targets := append([]string(nil), value.Targets...)
		basis := append([]string(nil), value.Basis...)
		output.BranchRelationFacts = append(output.BranchRelationFacts, LiuyaoBranchRelationOutput{
			LeftPosition: LiuyaoPosition(value.LeftPosition), RightPosition: LiuyaoPosition(value.RightPosition),
			Relations: relations, Targets: targets, Basis: basis,
		})
	}
	for _, value := range chart.DayClashFacts {
		output.DayClashFacts = append(output.DayClashFacts, LiuyaoDayClashOutput{
			Position: LiuyaoPosition(value.Position), Kind: value.Kind,
			Basis: append([]string(nil), value.Basis...), Condition: value.Condition,
		})
	}
	for _, value := range chart.MovingTransformations {
		output.MovingTransformations = append(output.MovingTransformations, LiuyaoMovingTransformationOutput{
			Position: LiuyaoPosition(value.Position), FromBranch: LiuyaoZhiName(value.FromBranch),
			ToBranch: LiuyaoZhiName(value.ToBranch), ReturnRelation: value.ReturnRelation,
			AdvanceRetreat: value.AdvanceRetreat, ChangedVoid: value.ChangedVoid,
			ChangedBreak: value.ChangedBreak, Basis: append([]string(nil), value.Basis...),
		})
	}
	for _, value := range chart.SanHeCandidates {
		positions := make([]LiuyaoPosition, 0, len(value.Positions))
		for _, position := range value.Positions {
			positions = append(positions, LiuyaoPosition(position))
		}
		branches := make([]LiuyaoZhiName, 0, len(value.Branches))
		for _, branch := range value.Branches {
			branches = append(branches, LiuyaoZhiName(branch))
		}
		missing := make([]LiuyaoZhiName, 0, len(value.Missing))
		for _, branch := range value.Missing {
			missing = append(missing, LiuyaoZhiName(branch))
		}
		output.SanHeCandidates = append(output.SanHeCandidates, LiuyaoSanHeOutput{
			Positions: positions, Branches: branches, Element: ElementName(value.Element),
			Complete: value.Complete, MovingCount: value.MovingCount,
			Qualification: value.Qualification, LineComplete: value.LineComplete,
			DayPresent: value.DayPresent, MonthPresent: value.MonthPresent,
			Missing: missing, StaticCount: value.StaticCount, VoidCount: value.VoidCount,
			BreakCount: value.BreakCount, Activated: value.Activated,
			Targets: append([]string(nil), value.Targets...), ConclusionScope: value.ConclusionScope,
		})
	}
	if chart.ForceChain != nil {
		force := LiuyaoForceChainOutput{
			YongShen:    LiuyaoYongShenName(chart.ForceChain.YongShen),
			YongElement: ElementName(chart.ForceChain.YongElement),
			Position:    LiuyaoOptionalPosition(chart.ForceChain.Position),
			IsHidden:    chart.ForceChain.IsHidden,
		}
		for _, entry := range chart.ForceChain.Entries {
			positions := make([]LiuyaoPosition, 0, len(entry.Positions))
			for _, position := range entry.Positions {
				positions = append(positions, LiuyaoPosition(position))
			}
			branches := make([]LiuyaoZhiName, 0, len(entry.Branches))
			for _, branch := range entry.Branches {
				branches = append(branches, LiuyaoZhiName(branch))
			}
			force.Entries = append(force.Entries, LiuyaoForceEntryOutput{
				Role: LiuyaoForceRole(entry.Role), Element: ElementName(entry.Element),
				Positions: positions, Branches: branches,
				Relations: append([]string(nil), entry.Relations...),
				States:    append([]string(nil), entry.States...), Note: entry.Note,
			})
		}
		output.ForceChain = &force
	}
	for _, value := range chart.YongShenCandidates {
		output.YongShenCandidates = append(output.YongShenCandidates, LiuyaoCandidateOutput{
			Position: LiuyaoOptionalPosition(value.Position), Branch: LiuyaoZhiName(value.Branch),
			IsHidden: value.IsHidden, Selected: value.Selected,
			WangShuai: LiuyaoWangShuai(value.WangShuai), Reason: value.Reason,
			Basis: append([]string(nil), value.Basis...),
		})
	}
	for _, value := range chart.Conflicts {
		output.Conflicts = append(output.Conflicts, LiuyaoConflictOutput{
			ID: LiuyaoConflictID(value.ID), Reason: value.Reason,
			Facts: append([]string(nil), value.Facts...),
		})
	}
	return output
}

func liuyaoCastingWire(casting liuyao.Casting) LiuyaoCastingOutput {
	output := LiuyaoCastingOutput{
		Mode:      LiuyaoCastMode(casting.Mode),
		Order:     LiuyaoOrderName(casting.Order),
		Yaos:      make(LiuyaoYaoValuesOutput, 0, len(casting.Yaos)),
		DongYao:   make(LiuyaoPositions, 0, len(casting.DongYao)),
		CastingID: LiuyaoCastingID(casting.CastingID),
	}
	if casting.CoinConvention != nil {
		output.CoinConvention = &LiuyaoCoinConventionOutput{
			Heads: LiuyaoHeadsValue(casting.CoinConvention["正"]),
			Tails: LiuyaoTailsValue(casting.CoinConvention["反"]),
		}
	}
	for _, round := range casting.Rounds {
		coins := [3]LiuyaoCoinSymbol{}
		for i, coin := range round.Coins {
			if i < len(coins) {
				coins[i] = LiuyaoCoinSymbol(coin)
			}
		}
		output.Rounds = append(output.Rounds, LiuyaoCoinRoundOutput{
			Position: LiuyaoPosition(round.Position),
			Coins:    coins,
			Value:    LiuyaoYaoValue(round.Value),
			Label:    LiuyaoYaoLabel(round.Label),
			Changing: round.Changing,
		})
	}
	for _, value := range casting.Yaos {
		output.Yaos = append(output.Yaos, LiuyaoYaoValue(value))
	}
	for _, position := range casting.DongYao {
		output.DongYao = append(output.DongYao, LiuyaoPosition(position))
	}
	return output
}
