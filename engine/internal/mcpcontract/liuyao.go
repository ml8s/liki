package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/liuyao"
)

// LiuyaoCoinRoundInput is one raw three-coin outcome. The engine owns totaling,
// labels, changing flags, and the normalized receipt.
type LiuyaoCoinRoundInput [3]LiuyaoCoinSymbol

type LiuyaoCoinRoundsInput []LiuyaoCoinRoundInput

type LiuyaoYaoValuesInput []LiuyaoYaoValue

// LiuyaoCastingRequestInput is the domain-owned source fact for one casting.
// It is the request form of LiuyaoCastingOutput.
type LiuyaoCastingRequestInput struct {
	Mode   LiuyaoQiguaMode       `json:"mode" jsonschema:"起卦模式：auto 安全随机，coins 传六组铜钱，yaos 传六个爻值。"`
	Rounds LiuyaoCoinRoundsInput `json:"rounds,omitempty" jsonschema:"coins 模式：六组三枚铜钱，初爻到上爻。"`
	Yaos   LiuyaoYaoValuesInput  `json:"yaos,omitempty" jsonschema:"yaos 模式：六个爻值，初爻到上爻。"`
}

type LiuyaoQiguaInput struct {
	Origin         LiuyaoCastingRequestInput `json:"origin" jsonschema:"起卦创建方式；只用于创建一次卦象事件。"`
	CastingContext LiuyaoCastingContextInput `json:"casting_context" jsonschema:"起卦地标准时间、经度与用户确认的占问焦点。"`
}

type LiuyaoAutoCastingRequest struct {
	Mode LiuyaoQiguaMode `json:"mode" jsonschema:"起卦模式；auto 使用安全随机。"`
}

type LiuyaoCoinsCastingRequest struct {
	Mode   LiuyaoQiguaMode       `json:"mode" jsonschema:"起卦模式；coins 要求六组铜钱记录。"`
	Rounds LiuyaoCoinRoundsInput `json:"rounds" jsonschema:"六组三枚铜钱，初爻到上爻。"`
}

type LiuyaoYaosCastingRequest struct {
	Mode LiuyaoQiguaMode      `json:"mode" jsonschema:"起卦模式；yaos 要求六个爻值。"`
	Yaos LiuyaoYaoValuesInput `json:"yaos" jsonschema:"六个爻值，初爻到上爻。"`
}

type LiuyaoQiguaAutoForm struct {
	Origin         LiuyaoAutoCastingRequest  `json:"origin" jsonschema:"起卦创建方式。"`
	CastingContext LiuyaoCastingContextInput `json:"casting_context" jsonschema:"起卦地标准时间、经度与用户确认的占问焦点。"`
}

type LiuyaoQiguaCoinsForm struct {
	Origin         LiuyaoCoinsCastingRequest `json:"origin" jsonschema:"起卦创建方式。"`
	CastingContext LiuyaoCastingContextInput `json:"casting_context" jsonschema:"起卦地标准时间、经度与用户确认的占问焦点。"`
}

type LiuyaoQiguaYaosForm struct {
	Origin         LiuyaoYaosCastingRequest  `json:"origin" jsonschema:"起卦创建方式。"`
	CastingContext LiuyaoCastingContextInput `json:"casting_context" jsonschema:"起卦地标准时间、经度与用户确认的占问焦点。"`
}

func liuyaoQiguaFormsSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	modes := []string{"auto", "coins", "yaos"}
	forms := make([]json.RawMessage, 0, len(modes))
	for _, mode := range modes {
		var raw json.RawMessage
		var err error
		nested, err := liuyaoOriginBranchSchema(options, mode)
		if err != nil {
			return nil, err
		}
		switch mode {
		case "auto":
			_, raw, err = inputSchemaFor[LiuyaoQiguaAutoForm](options)
		case "coins":
			_, raw, err = inputSchemaFor[LiuyaoQiguaCoinsForm](options)
		case "yaos":
			_, raw, err = inputSchemaFor[LiuyaoQiguaYaosForm](options)
		}
		if err != nil {
			return nil, err
		}
		raw, err = withSchemaProperty(raw, "origin", nested)
		if err != nil {
			return nil, err
		}
		castingContext, err := liuyaoCastingContextSchema(options)
		if err != nil {
			return nil, err
		}
		raw, err = withSchemaProperty(raw, "casting_context", castingContext)
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

type LiuyaoHeadsScore int
type LiuyaoTailsScore int

type LiuyaoCoinConvention struct {
	Heads LiuyaoHeadsScore `json:"正" jsonschema:"正面记 3。"`
	Tails LiuyaoTailsScore `json:"反" jsonschema:"反面记 2。"`
}

type LiuyaoCoinRoundsOutput []LiuyaoCoinRoundOutput

type LiuyaoCoinRoundOutput struct {
	Position LiuyaoPosition      `json:"position" jsonschema:"爻位，初爻到上爻 1-6。"`
	Coins    [3]LiuyaoCoinSymbol `json:"coins" jsonschema:"三枚铜钱记录，正/反按调用顺序保留。"`
	Value    LiuyaoYaoValue      `json:"yao_value" jsonschema:"三币合计爻值 6-9。"`
	Label    LiuyaoYaoLabel      `json:"yao_label" jsonschema:"老阴/少阳/少阴/老阳。"`
	Changing bool                `json:"is_dong_yao" jsonschema:"是否动爻。"`
}

type LiuyaoYaoValuesOutput []LiuyaoYaoValue
type LiuyaoPositions []LiuyaoPosition

type LiuyaoCastingOutput struct {
	Mode           LiuyaoCastMode         `json:"mode" jsonschema:"实际使用的起卦模式。"`
	Order          LiuyaoOrderName        `json:"order" jsonschema:"爻序；当前固定为初爻到上爻。"`
	CoinConvention *LiuyaoCoinConvention  `json:"coin_convention,omitempty" jsonschema:"coins 模式的铜钱记分约定。"`
	Rounds         LiuyaoCoinRoundsOutput `json:"rounds,omitempty" jsonschema:"coins 模式的六组铜钱记录。"`
	Yaos           LiuyaoYaoValuesOutput  `json:"yaos" jsonschema:"六个爻值，初爻到上爻。"`
	DongYao        LiuyaoPositions        `json:"dong_yao" jsonschema:"动爻位置列表，1-6。"`
}

type LiuyaoQiguaOutput struct {
	Event LiuyaoCastingEvent `json:"event" jsonschema:"一次起卦创建的不可变卦象事件。"`
}

type LiuyaoCoinsCastingReceipt struct {
	Mode           LiuyaoCastMode         `json:"mode" jsonschema:"起卦方式。"`
	Order          LiuyaoOrderName        `json:"order" jsonschema:"爻序。"`
	CoinConvention *LiuyaoCoinConvention  `json:"coin_convention,omitempty" jsonschema:"铜钱记分约定。"`
	Rounds         LiuyaoCoinRoundsOutput `json:"rounds" jsonschema:"六组铜钱记录。"`
	Yaos           LiuyaoYaoValuesOutput  `json:"yaos" jsonschema:"六个爻值。"`
	DongYao        LiuyaoPositions        `json:"dong_yao" jsonschema:"动爻位置列表。"`
}

type LiuyaoYaosCastingReceipt struct {
	Mode    LiuyaoCastMode        `json:"mode" jsonschema:"起卦方式。"`
	Order   LiuyaoOrderName       `json:"order" jsonschema:"爻序。"`
	Yaos    LiuyaoYaoValuesOutput `json:"yaos" jsonschema:"六个爻值。"`
	DongYao LiuyaoPositions       `json:"dong_yao" jsonschema:"动爻位置列表。"`
}

type LiuyaoMatterFocusInput struct {
	Kind        string `json:"kind" jsonschema:"焦点类型；matter 表示按问事类别推导。"`
	Matter      string `json:"matter" jsonschema:"问事类别；relationship 需同时提供 perspective。"`
	Perspective string `json:"perspective,omitempty" jsonschema:"仅 relationship 使用：问卦人性别，male 男占女，female 女占男。"`
}

type LiuyaoYongShenFocusInput struct {
	Kind     string             `json:"kind" jsonschema:"焦点类型；yong_shen 表示显式用神。"`
	YongShen LiuyaoYongShenName `json:"yong_shen" jsonschema:"高级显式用神。"`
}

type LiuyaoFocusInput struct {
	Kind        string `json:"kind"`
	Matter      string `json:"matter,omitempty"`
	Perspective string `json:"perspective,omitempty"`
	YongShen    string `json:"yong_shen,omitempty"`
}

type LiuyaoCastingContextInput struct {
	StandardTime StandardTime     `json:"standard_time" jsonschema:"起卦地标准时间（无 DST），RFC3339 带 offset。"`
	Longitude    Longitude        `json:"longitude" jsonschema:"起卦地经度（-180..180）；真太阳时校正用于时辰。"`
	Focus        LiuyaoFocusInput `json:"focus" jsonschema:"占问焦点；起卦时确定性解析为用神。"`
}

// LiuyaoCastingContext is the immutable event context. It carries only the
// resolved yong_shen so chart/ask never re-interpret the user-facing focus.
type LiuyaoCastingContext struct {
	StandardTime StandardTime       `json:"standard_time" jsonschema:"起卦地标准时间（无 DST），RFC3339 带 offset。"`
	Longitude    Longitude          `json:"longitude" jsonschema:"起卦地经度（-180..180）；真太阳时校正用于时辰。"`
	YongShen     LiuyaoYongShenName `json:"yong_shen" jsonschema:"起卦时解析出的用神；chart/ask 唯一取用依据。"`
}

type LiuyaoCastingEvent struct {
	Casting        LiuyaoCastingOutput  `json:"casting" jsonschema:"一次起卦的完整源事实。"`
	CastingContext LiuyaoCastingContext `json:"casting_context" jsonschema:"起卦地标准时间、经度与解析出的用神。"`
}

// LiuyaoCastingEventInput accepts both the canonical minimal event and legacy
// events that still carry question/origin/focus. Legacy fields are normalized
// away before chart/ask logic runs.
type LiuyaoCastingEventInput struct {
	Question       string                          `json:"question,omitempty" jsonschema:"旧事件兼容字段；新事件不携带。"`
	Origin         *LiuyaoCastingRequestInput      `json:"origin,omitempty" jsonschema:"旧事件兼容字段；新事件不携带。"`
	Casting        LiuyaoCastingOutput             `json:"casting" jsonschema:"一次起卦的完整源事实。"`
	CastingContext LiuyaoCastingContextInputCompat `json:"casting_context" jsonschema:"起卦地标准时间、经度与用神或旧版焦点。"`
}

type LiuyaoCastingContextInputCompat struct {
	StandardTime StandardTime      `json:"standard_time" jsonschema:"起卦地标准时间（无 DST），RFC3339 带 offset。"`
	Longitude    Longitude         `json:"longitude" jsonschema:"起卦地经度（-180..180）；真太阳时校正用于时辰。"`
	Focus        *LiuyaoFocusInput `json:"focus,omitempty" jsonschema:"旧事件兼容字段；新事件不携带。"`
	YongShen     string            `json:"yong_shen,omitempty" jsonschema:"起卦时解析出的用神；新事件必填。"`
}

func liuyaoMatterFocusSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[LiuyaoMatterFocusInput](options)
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaConst(raw, "kind", "matter")
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaEnum(
		raw, "matter",
		"self", "other", "parents", "children", "siblings",
		"wealth", "career", "study", "health", "legal", "home", "travel", "lost_item",
	)
	if err != nil {
		return nil, err
	}
	// perspective is only valid on the relationship branch.
	return withoutSchemaProperty(raw, "perspective")
}

func liuyaoRelationshipFocusSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[LiuyaoMatterFocusInput](options)
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaConst(raw, "kind", "matter")
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaConst(raw, "matter", "relationship")
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaEnum(raw, "perspective", "male", "female")
	if err != nil {
		return nil, err
	}
	return withSchemaRequiredFields(raw, "kind", "matter", "perspective")
}

func liuyaoYongShenFocusSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[LiuyaoYongShenFocusInput](options)
	if err != nil {
		return nil, err
	}
	return withSchemaConst(raw, "kind", "yong_shen")
}

func liuyaoOriginBranchSchema(options *jsonschema.ForOptions, mode string) (json.RawMessage, error) {
	var raw json.RawMessage
	var err error
	switch mode {
	case "auto":
		_, raw, err = inputSchemaFor[LiuyaoAutoCastingRequest](options)
	case "coins":
		_, raw, err = inputSchemaFor[LiuyaoCoinsCastingRequest](options)
	case "yaos":
		_, raw, err = inputSchemaFor[LiuyaoYaosCastingRequest](options)
	default:
		return nil, fmt.Errorf("unsupported liuyao origin mode %q", mode)
	}
	if err != nil {
		return nil, err
	}
	return withSchemaConst(raw, "mode", mode)
}

func liuyaoOriginSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	modes := []string{"auto", "coins", "yaos"}
	branches := make([]json.RawMessage, 0, len(modes))
	for _, mode := range modes {
		branch, err := liuyaoOriginBranchSchema(options, mode)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}
	return json.Marshal(map[string]any{"oneOf": branches})
}

func patchLiuyaoEvent(event map[string]any, options *jsonschema.ForOptions) (map[string]any, error) {
	raw, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	casting, err := liuyaoCastingReceiptSchema(options)
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaProperty(raw, "casting", casting)
	if err != nil {
		return nil, err
	}
	context, err := liuyaoEventCastingContextSchema(options)
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaProperty(raw, "casting_context", context)
	if err != nil {
		return nil, err
	}
	var patched map[string]any
	if err := json.Unmarshal(raw, &patched); err != nil {
		return nil, err
	}
	return patched, nil
}

func patchLiuyaoEventInput(event map[string]any, options *jsonschema.ForOptions) (map[string]any, error) {
	raw, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	origin, err := liuyaoOriginSchema(options)
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaProperty(raw, "origin", origin)
	if err != nil {
		return nil, err
	}
	casting, err := liuyaoCastingReceiptSchema(options)
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaProperty(raw, "casting", casting)
	if err != nil {
		return nil, err
	}
	context, err := liuyaoEventInputCastingContextSchema(options)
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaProperty(raw, "casting_context", context)
	if err != nil {
		return nil, err
	}
	var patched map[string]any
	if err := json.Unmarshal(raw, &patched); err != nil {
		return nil, err
	}
	return patched, nil
}

func withLiuyaoEventSchema(document map[string]any, options *jsonschema.ForOptions) (map[string]any, error) {
	properties, ok := document["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no properties")
	}
	event, ok := properties["event"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no event properties")
	}
	patched, err := patchLiuyaoEvent(event, options)
	if err != nil {
		return nil, err
	}
	properties["event"] = patched
	return document, nil
}

func withLiuyaoEventInputSchema(document map[string]any, options *jsonschema.ForOptions) (map[string]any, error) {
	properties, ok := document["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no properties")
	}
	event, ok := properties["event"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no event properties")
	}
	patched, err := patchLiuyaoEventInput(event, options)
	if err != nil {
		return nil, err
	}
	properties["event"] = patched
	return document, nil
}

func liuyaoFocusSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	matter, err := liuyaoMatterFocusSchema(options)
	if err != nil {
		return nil, err
	}
	relationship, err := liuyaoRelationshipFocusSchema(options)
	if err != nil {
		return nil, err
	}
	yongShen, err := liuyaoYongShenFocusSchema(options)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"oneOf": []json.RawMessage{matter, relationship, yongShen},
	})
}

func liuyaoCastingContextSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[LiuyaoCastingContextInput](options)
	if err != nil {
		return nil, err
	}
	focus, err := liuyaoFocusSchema(options)
	if err != nil {
		return nil, err
	}
	return withSchemaProperty(raw, "focus", focus)
}

func liuyaoEventCastingContextSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[LiuyaoCastingContext](options)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func liuyaoEventInputCastingContextSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[LiuyaoCastingContextInputCompat](options)
	if err != nil {
		return nil, err
	}
	focus, err := liuyaoFocusSchema(options)
	if err != nil {
		return nil, err
	}
	raw, err = withSchemaProperty(raw, "focus", focus)
	if err != nil {
		return nil, err
	}
	return withSchemaEitherRequired(raw, "yong_shen", "focus")
}

func liuyaoCastingReceiptSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	_, coins, err := inputSchemaFor[LiuyaoCoinsCastingReceipt](options)
	if err != nil {
		return nil, err
	}
	coins, err = withSchemaConst(coins, "mode", "coins")
	if err != nil {
		return nil, err
	}
	coins, err = withSchemaConst(coins, "order", "bottom_up")
	if err != nil {
		return nil, err
	}
	_, yaos, err := inputSchemaFor[LiuyaoYaosCastingReceipt](options)
	if err != nil {
		return nil, err
	}
	yaos, err = withSchemaConst(yaos, "mode", "yaos")
	if err != nil {
		return nil, err
	}
	yaos, err = withSchemaConst(yaos, "order", "bottom_up")
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"oneOf": []json.RawMessage{coins, yaos}})
}

func liuyaoChartSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[LiuyaoChartInput](options)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	document, err = withLiuyaoEventInputSchema(document, options)
	if err != nil {
		return nil, err
	}
	return json.Marshal(document)
}

func liuyaoStrictArraySchema[T any](options *jsonschema.ForOptions) (json.RawMessage, error) {
	raw, _, err := outputSchemaFor[T](options)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	strictArrays(document)
	if _, ok := document["oneOf"]; !ok {
		return json.Marshal(document)
	}
	branches, ok := document["oneOf"].([]any)
	if !ok {
		return nil, fmt.Errorf("schema oneOf is not an array")
	}
	for _, branch := range branches {
		strictArrays(branch)
	}
	return json.Marshal(document)
}

func liuyaoQiguaOutputSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	raw, _, err := outputSchemaFor[LiuyaoQiguaOutput](options)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	document, err = withLiuyaoEventSchema(document, options)
	if err != nil {
		return nil, err
	}
	strictArrays(document)
	return json.Marshal(document)
}

func liuyaoChartOutputSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	raw, err := liuyaoStrictArraySchema[LiuyaoChartResultOutput](options)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	document, err = withLiuyaoEventSchema(document, options)
	if err != nil {
		return nil, err
	}
	return json.Marshal(document)
}

func strictArrays(value any) {
	switch node := value.(type) {
	case map[string]any:
		if types, ok := node["type"].([]any); ok && len(types) == 2 {
			hasNull, hasArray := false, false
			for _, kind := range types {
				if kind == "null" {
					hasNull = true
				}
				if kind == "array" {
					hasArray = true
				}
			}
			if hasNull && hasArray {
				node["type"] = "array"
			}
		}
		for _, child := range node {
			strictArrays(child)
		}
	case []any:
		for _, child := range node {
			strictArrays(child)
		}
	}
}

type LiuyaoTimingOutput struct {
	ID            LiuyaoTimingID         `json:"id" jsonschema:"应期候选 ID。"`
	Mechanism     string                 `json:"mechanism" jsonschema:"应期机制。"`
	Position      LiuyaoOptionalPosition `json:"position,omitempty" jsonschema:"相关爻位；0 表示卦级候选。"`
	TriggerBranch ZhiName                `json:"trigger_branch,omitempty" jsonschema:"触发地支。"`
	Window        string                 `json:"window,omitempty" jsonschema:"应期窗口。"`
	Basis         []string               `json:"basis,omitempty" jsonschema:"确定性依据。"`
}

type LiuyaoDongRelationOutput struct {
	Position  LiuyaoPosition           `json:"position" jsonschema:"动爻位置 1-6。"`
	Relations []LiuyaoDongRelationType `json:"relations" jsonschema:"动爻与用神/原神/忌神的关系集合。"`
}

type LiuyaoHiddenLineOutput struct {
	Position       LiuyaoPosition   `json:"position" jsonschema:"伏神爻位。"`
	LiuQin         LiuyaoLiuQinName `json:"liu_qin" jsonschema:"伏神六亲。"`
	Gan            GanName          `json:"gan" jsonschema:"伏神天干。"`
	Zhi            ZhiName          `json:"zhi" jsonschema:"伏神地支。"`
	Wuxing         ElementName      `json:"wuxing" jsonschema:"伏神五行。"`
	FlyingPosition LiuyaoPosition   `json:"flying_position" jsonschema:"飞神爻位。"`
	FlyingLiuQin   LiuyaoLiuQinName `json:"flying_liu_qin" jsonschema:"飞神六亲。"`
	FlyingGan      GanName          `json:"flying_gan" jsonschema:"飞神天干。"`
	FlyingZhi      ZhiName          `json:"flying_zhi" jsonschema:"飞神地支。"`
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
	FromBranch     ZhiName        `json:"from_branch" jsonschema:"动爻本变前地支。"`
	ToBranch       ZhiName        `json:"to_branch" jsonschema:"变爻地支。"`
	ReturnRelation string         `json:"return_relation,omitempty" jsonschema:"变爻返回本爻的关系。"`
	AdvanceRetreat string         `json:"advance_retreat,omitempty" jsonschema:"进神或退神。"`
	ChangedVoid    bool           `json:"changed_void,omitempty" jsonschema:"变爻是否旬空。"`
	ChangedBreak   bool           `json:"changed_break,omitempty" jsonschema:"变爻是否月破。"`
	Basis          []string       `json:"basis,omitempty" jsonschema:"机械变化依据。"`
}

type LiuyaoSanHeOutput struct {
	Positions       []LiuyaoPosition `json:"positions" jsonschema:"参与三合候选的爻位。"`
	Branches        []ZhiName        `json:"branches" jsonschema:"参与三合候选的地支。"`
	Element         ElementName      `json:"wuxing" jsonschema:"三合局五行。"`
	Complete        bool             `json:"complete" jsonschema:"卦爻与日月补支后三支是否齐。"`
	MovingCount     int              `json:"moving_count" jsonschema:"候选中发动爻数。"`
	Qualification   string           `json:"qualification" jsonschema:"成局资格限定。"`
	LineComplete    bool             `json:"line_complete" jsonschema:"卦爻本身是否三支齐。"`
	DayPresent      bool             `json:"day_present" jsonschema:"日建是否补支。"`
	MonthPresent    bool             `json:"month_present" jsonschema:"月建是否补支。"`
	Missing         []ZhiName        `json:"missing,omitempty" jsonschema:"缺失的地支。"`
	StaticCount     int              `json:"static_count" jsonschema:"候选中安静爻数。"`
	VoidCount       int              `json:"void_count" jsonschema:"候选中旬空爻数。"`
	BreakCount      int              `json:"break_count" jsonschema:"候选中月破爻数。"`
	Activated       bool             `json:"activated" jsonschema:"是否满足当前确定性成局候选条件。"`
	Targets         []string         `json:"targets,omitempty" jsonschema:"候选指向的目标。"`
	ConclusionScope string           `json:"conclusion_scope" jsonschema:"结论边界；仅结构候选，不直接断结果。"`
}

type LiuyaoForceEntryOutput struct {
	Role      LiuyaoForceRole  `json:"role" jsonschema:"力量链角色。"`
	Element   ElementName      `json:"wuxing" jsonschema:"角色五行。"`
	Positions []LiuyaoPosition `json:"positions,omitempty" jsonschema:"同五行爻位。"`
	Branches  []ZhiName        `json:"branches,omitempty" jsonschema:"同五行地支。"`
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
	Branch    ZhiName                `json:"branch" jsonschema:"候选地支。"`
	IsHidden  bool                   `json:"is_hidden" jsonschema:"是否伏神候选。"`
	Selected  bool                   `json:"selected" jsonschema:"是否当前确定性选用。"`
	WangShuai WangShuaiName          `json:"wang_shuai,omitempty" jsonschema:"候选旺衰。"`
	Reason    string                 `json:"reason" jsonschema:"选择或列出原因。"`
	Basis     []string               `json:"basis,omitempty" jsonschema:"选择依据。"`
}

type LiuyaoConflictOutput struct {
	ID     LiuyaoConflictID `json:"id" jsonschema:"并列或矛盾信号 ID。"`
	Reason string           `json:"reason" jsonschema:"为什么必须解释而不是静默择一。"`
	Facts  []string         `json:"facts,omitempty" jsonschema:"构成矛盾的事实。"`
}

type LiuyaoChartResultOutput struct {
	Event LiuyaoCastingEvent `json:"event" jsonschema:"本次排盘原样回传的不可变卦象事件。"`
	Chart LiuyaoChartView    `json:"chart" jsonschema:"六爻盘面统一视图；analysis 一并置于 chart 内。"`
}

// liuyaoMatterYongShen is the single source of truth for matter -> yong_shen.
// rationale:
//   - self/health/travel: 自占以世爻为用；疾病以世为身，官鬼为病神；出行吉凶看世。
//   - other: 无尊卑称呼、未深交、泛指对象取应爻。
//   - relationship: 男占女以妻财为主，女占男以官鬼为主；由 perspective 决定。
//   - parents/home: 父母为墙城宅舍、文书。
//   - children/siblings: 六亲分类。
//   - wealth/lost_item: 财爻为钱物、所失物之主。
//   - career/legal: 官鬼为功名官府、问官。
//   - study: 父爻为学业文书，官鬼只作名次辅助。
//
// Legacy matters remain readable for historical events only.
var liuyaoMatterYongShen = map[string]LiuyaoYongShenName{
	"self":        "世爻",
	"other":       "应爻",
	"parents":     "父母",
	"children":    "子孙",
	"siblings":    "兄弟",
	"wealth":      "妻财",
	"career":      "官鬼",
	"study":       "父母",
	"health":      "世爻",
	"legal":       "官鬼",
	"home":        "父母",
	"travel":      "世爻",
	"lost_item":   "妻财",
	"general":     "世爻",
	"family":      "世爻",
	"competition": "兄弟",
}

func liuyaoResolveFocus(focus LiuyaoFocusInput) (LiuyaoYongShenName, error) {
	switch focus.Kind {
	case "matter":
		if focus.YongShen != "" {
			return "", fmt.Errorf("matter focus does not accept yong_shen")
		}
		if focus.Matter == "relationship" {
			if focus.Perspective != "male" && focus.Perspective != "female" {
				return "", fmt.Errorf("relationship focus requires perspective=male/female")
			}
			if focus.Perspective == "male" {
				return "妻财", nil
			} else {
				return "官鬼", nil
			}
		}
		if focus.Perspective != "" {
			return "", fmt.Errorf("perspective is only valid for relationship")
		}
		name, ok := liuyaoMatterYongShen[focus.Matter]
		if !ok {
			return "", fmt.Errorf("unsupported matter focus %q", focus.Matter)
		}
		return name, nil
	case "yong_shen":
		if focus.Matter != "" || focus.Perspective != "" {
			return "", fmt.Errorf("yong_shen focus does not accept matter or perspective")
		}
		if focus.YongShen == "" {
			return "", fmt.Errorf("yong_shen focus requires yong_shen")
		}
		if _, err := liuyao.ParseYongShen(focus.YongShen); err != nil {
			return "", fmt.Errorf("invalid yong_shen focus: %w", err)
		}
		return LiuyaoYongShenName(focus.YongShen), nil
	default:
		return "", fmt.Errorf("focus.kind must be matter/yong_shen")
	}
}

// liuyaoEventYongShen reads the resolved yong_shen from the event. The focus
// fallback exists only so pre-upgrade events remain consumable; new events
// written by liuyao_qigua always carry yong_shen and never carry focus.
func liuyaoEventYongShen(ctx LiuyaoCastingContextInputCompat) (LiuyaoYongShenName, error) {
	if ctx.YongShen != "" {
		if ctx.Focus != nil {
			return "", fmt.Errorf("casting_context cannot carry both yong_shen and focus")
		}
		if _, err := liuyao.ParseYongShen(ctx.YongShen); err != nil {
			return "", fmt.Errorf("invalid yong_shen: %w", err)
		}
		return LiuyaoYongShenName(ctx.YongShen), nil
	}
	if ctx.Focus == nil {
		return "", fmt.Errorf("casting_context.yong_shen is required")
	}
	return liuyaoResolveFocus(*ctx.Focus)
}

type LiuyaoAnalysisOutput struct {
	TimingCandidates      []LiuyaoTimingOutput               `json:"timing_candidates,omitempty" jsonschema:"应期候选。"`
	DongYaoRelations      []LiuyaoDongRelationOutput         `json:"dong_yao_relations,omitempty" jsonschema:"动爻与用神/原神/忌神关系。"`
	HiddenLines           []LiuyaoHiddenLineOutput           `json:"hidden_lines,omitempty" jsonschema:"本宫全部伏神。"`
	BranchRelationFacts   []LiuyaoBranchRelationOutput       `json:"branch_relation_facts,omitempty" jsonschema:"爻间地支关系。"`
	TombSchool            LiuyaoTombOutput                   `json:"tomb_school" jsonschema:"土爻墓库流派。"`
	DayClashFacts         []LiuyaoDayClashOutput             `json:"day_clash_facts,omitempty" jsonschema:"日冲事实。"`
	MovingTransformations []LiuyaoMovingTransformationOutput `json:"moving_transformations,omitempty" jsonschema:"动爻机械变化。"`
	SanHeCandidates       []LiuyaoSanHeOutput                `json:"san_he_candidates,omitempty" jsonschema:"三合结构候选。"`
	ForceChain            *LiuyaoForceChainOutput            `json:"yong_shen_force_chain,omitempty" jsonschema:"用神生克力量链。"`
	YongShenCandidates    []LiuyaoCandidateOutput            `json:"yong_shen_candidates,omitempty" jsonschema:"用神候选。"`
	Conflicts             []LiuyaoConflictOutput             `json:"conflicts,omitempty" jsonschema:"并列或矛盾信号。"`
}

type LiuyaoChartLineFlags struct {
	Moving      bool   `json:"moving" jsonschema:"本爻是否发动。"`
	YuePo       bool   `json:"yue_po" jsonschema:"是否月破。"`
	XunKong     bool   `json:"xun_kong" jsonschema:"是否旬空。"`
	MuKu        bool   `json:"mu_ku" jsonschema:"是否入墓库。"`
	MuKuBranch  string `json:"mu_ku_branch,omitempty" jsonschema:"墓库地支。"`
	MuKuElement string `json:"mu_ku_element,omitempty" jsonschema:"墓库五行。"`
	DongSheng   bool   `json:"dong_sheng" jsonschema:"是否受动爻生。"`
	DongKe      bool   `json:"dong_ke" jsonschema:"是否受动爻克。"`
}

type LiuyaoChartLineView struct {
	Position        LiuyaoOptionalPosition `json:"position" jsonschema:"爻位；0 表示未定位。"`
	YaoValue        LiuyaoOptionalYaoValue `json:"yao_value" jsonschema:"爻值；0 表示无变爻。"`
	GanZhi          string                 `json:"gan_zhi" jsonschema:"该爻干支。"`
	Wuxing          string                 `json:"wuxing" jsonschema:"该爻五行。"`
	LiuQin          string                 `json:"liu_qin" jsonschema:"六亲。"`
	ShiYing         LiuyaoShiYing          `json:"shi_ying,omitempty" jsonschema:"世爻或应爻标记。"`
	LiuShou         string                 `json:"liu_shou" jsonschema:"六神。"`
	RiChenRelations map[string][]string    `json:"ri_chen_relations,omitempty" jsonschema:"与日辰的关系。"`
	WangShuai       string                 `json:"wang_shuai,omitempty" jsonschema:"旺衰状态。"`
	ChangShengYue   string                 `json:"chang_sheng_yue,omitempty" jsonschema:"月令长生状态。"`
	Flags           LiuyaoChartLineFlags   `json:"flags" jsonschema:"爻位状态标记。"`
}

type LiuyaoChartPatternView struct {
	ID          string                 `json:"id" jsonschema:"格局 ID。"`
	PatternKind string                 `json:"pattern_kind" jsonschema:"格局类型。"`
	SubType     string                 `json:"sub_type" jsonschema:"格局子类型。"`
	Position    LiuyaoOptionalPosition `json:"position" jsonschema:"关联爻位。"`
	IsTrue      bool                   `json:"is_true" jsonschema:"结构是否成立。"`
}

type LiuyaoGuaCiView struct {
	Name  string   `json:"name" jsonschema:"卦名。"`
	Text  string   `json:"text" jsonschema:"卦辞原文。"`
	YaoCi []string `json:"yao_ci" jsonschema:"各爻爻辞。"`
}

type LiuyaoFuShenView struct {
	Position LiuyaoPosition `json:"position" jsonschema:"伏神所在爻位。"`
	LiuQin   string         `json:"liu_qin" jsonschema:"伏神六亲。"`
	Zhi      string         `json:"zhi" jsonschema:"伏神地支。"`
}

type LiuyaoYongShenView struct {
	Name             string             `json:"name" jsonschema:"用神类别。"`
	Position         int                `json:"position" jsonschema:"用神爻位；0 表示本卦不现。"`
	IsHidden         bool               `json:"is_hidden,omitempty" jsonschema:"本卦不现时是否取伏神。"`
	WangShuai        string             `json:"wang_shuai" jsonschema:"用神旺衰。"`
	YuePo            bool               `json:"yue_po,omitempty" jsonschema:"用神是否月破。"`
	XunKong          bool               `json:"xun_kong,omitempty" jsonschema:"用神是否旬空。"`
	MuKu             bool               `json:"mu_ku,omitempty" jsonschema:"用神是否入墓。"`
	LiuShou          string             `json:"liu_shou,omitempty" jsonschema:"用神临的六兽；伏神不继承飞神六兽。"`
	ChangSheng       string             `json:"chang_sheng,omitempty" jsonschema:"用神五行在月支的十二长生。"`
	FuShen           *LiuyaoFuShenView  `json:"fu_shen,omitempty" jsonschema:"用神不现时选用的伏神。"`
	FuShenCandidates []LiuyaoFuShenView `json:"fu_shen_candidates,omitempty" jsonschema:"可用伏神候选。"`
}

type LiuyaoChartView struct {
	GuaName      string                   `json:"gua_name" jsonschema:"本卦名称。"`
	BianGua      *string                  `json:"bian_gua" jsonschema:"变卦名称；无动爻时为 null。"`
	Palace       string                   `json:"palace" jsonschema:"卦宫。"`
	PalaceWuxing string                   `json:"palace_wuxing" jsonschema:"卦宫五行。"`
	Lines        []LiuyaoChartLineView    `json:"lines" jsonschema:"本卦六爻。"`
	BianLines    []LiuyaoChartLineView    `json:"bian_lines" jsonschema:"发生变化的爻。"`
	RiGan        string                   `json:"ri_gan" jsonschema:"日辰天干。"`
	RiZhi        string                   `json:"ri_zhi" jsonschema:"日辰地支。"`
	YueGan       string                   `json:"yue_gan" jsonschema:"月建天干。"`
	YueZhi       string                   `json:"yue_zhi" jsonschema:"月建地支。"`
	XunKong      []string                 `json:"xun_kong" jsonschema:"旬空地支。"`
	Patterns     []LiuyaoChartPatternView `json:"patterns" jsonschema:"盘面格局。"`
	YongShen     LiuyaoYongShenView       `json:"yong_shen" jsonschema:"最终用神状态。"`
	GuaCi        *LiuyaoGuaCiView         `json:"gua_ci,omitempty" jsonschema:"卦爻辞。"`
	Analysis     LiuyaoAnalysisOutput     `json:"analysis" jsonschema:"六爻判断层确定性分析事实。"`
}

func LiuyaoQigua(_ context.Context, input LiuyaoQiguaInput) (LiuyaoQiguaOutput, error) {
	yongShen, err := liuyaoResolveFocus(input.CastingContext.Focus)
	if err != nil {
		return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: invalid focus: %w", err)
	}
	if _, _, err := parseStandardTime(input.CastingContext.StandardTime); err != nil {
		return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: %w", err)
	}
	if float64(input.CastingContext.Longitude) < -180 || float64(input.CastingContext.Longitude) > 180 {
		return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: longitude must be within [-180,180], got %v", float64(input.CastingContext.Longitude))
	}
	var casting liuyao.Casting
	request := input.Origin
	switch request.Mode {
	case "auto":
		if request.Rounds != nil || request.Yaos != nil {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: auto mode does not accept rounds or yaos")
		}
		casting, err = liuyao.SecureQigua()
	case "coins":
		if request.Yaos != nil {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: coins mode accepts rounds only")
		}
		if len(request.Rounds) != 6 {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: casting.coins requires exactly 6 rounds, got %d", len(request.Rounds))
		}
		rounds := make([][]liuyao.CoinFace, 0, 6)
		for _, round := range request.Rounds {
			coins := make([]liuyao.CoinFace, 0, 3)
			for _, coin := range round {
				coins = append(coins, liuyao.CoinFace(coin))
			}
			rounds = append(rounds, coins)
		}
		casting, err = liuyao.NewCoinsCasting(rounds)
	case "yaos":
		if request.Rounds != nil {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: yaos mode accepts yaos only")
		}
		if len(request.Yaos) != 6 {
			return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: casting.yaos requires exactly 6 values, got %d", len(request.Yaos))
		}
		var values [6]int
		for i, value := range request.Yaos {
			values[i] = int(value)
		}
		casting, err = liuyao.NewValuesCasting(values)
	default:
		return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: casting.mode unsupported %q", string(request.Mode))
	}
	if err != nil {
		return LiuyaoQiguaOutput{}, fmt.Errorf("liuyao_qigua: %w", err)
	}
	return LiuyaoQiguaOutput{Event: LiuyaoCastingEvent{
		Casting: liuyaoCastingWire(casting),
		CastingContext: LiuyaoCastingContext{
			StandardTime: input.CastingContext.StandardTime,
			Longitude:    input.CastingContext.Longitude,
			YongShen:     yongShen,
		},
	}}, nil
}

// LiuyaoChartInput consumes the immutable casting event created by liuyao_qigua.
// The input type tolerates legacy events (question/origin/focus), while new
// events are normalized to the minimal canonical shape before use.
type LiuyaoChartInput struct {
	Event *LiuyaoCastingEventInput `json:"event" jsonschema:"liuyao_qigua 返回的不可变卦象事件。"`
}

func validateLiuyaoEventOrigin(origin *LiuyaoCastingRequestInput, casting LiuyaoCastingOutput) error {
	if origin == nil {
		return nil
	}
	switch origin.Mode {
	case "auto":
		if origin.Rounds != nil || origin.Yaos != nil {
			return fmt.Errorf("auto origin does not accept rounds or yaos")
		}
		if casting.Mode != "coins" {
			return fmt.Errorf("auto origin requires a coins casting receipt")
		}
	case "coins":
		if origin.Yaos != nil || len(origin.Rounds) != 6 {
			return fmt.Errorf("coins origin requires exactly 6 rounds and no yaos")
		}
		if casting.Mode != "coins" {
			return fmt.Errorf("coins origin requires a coins casting receipt")
		}
	case "yaos":
		if origin.Rounds != nil || len(origin.Yaos) != 6 {
			return fmt.Errorf("yaos origin requires exactly 6 yaos and no rounds")
		}
		if casting.Mode != "yaos" {
			return fmt.Errorf("yaos origin requires a yaos casting receipt")
		}
	default:
		return fmt.Errorf("unsupported origin mode %q", string(origin.Mode))
	}
	return nil
}

func LiuyaoChart(_ context.Context, input LiuyaoChartInput) (LiuyaoChartResultOutput, error) {
	if input.Event == nil {
		return LiuyaoChartResultOutput{}, fmt.Errorf("liuyao.chart: event is required")
	}
	if err := validateLiuyaoEventOrigin(input.Event.Origin, input.Event.Casting); err != nil {
		return LiuyaoChartResultOutput{}, fmt.Errorf("liuyao.chart: invalid event origin: %w", err)
	}
	casting, err := liuyaoCastingDomain(input.Event.Casting)
	if err != nil {
		return LiuyaoChartResultOutput{}, fmt.Errorf("liuyao.chart: invalid casting: %w", err)
	}
	yongShen, err := liuyaoEventYongShen(input.Event.CastingContext)
	if err != nil {
		return LiuyaoChartResultOutput{}, fmt.Errorf("liuyao.chart: %w", err)
	}
	standard, _, err := parseStandardTime(input.Event.CastingContext.StandardTime)
	if err != nil {
		return LiuyaoChartResultOutput{}, fmt.Errorf("liuyao.chart: %w", err)
	}
	if float64(input.Event.CastingContext.Longitude) < -180 || float64(input.Event.CastingContext.Longitude) > 180 {
		return LiuyaoChartResultOutput{}, fmt.Errorf("liuyao.chart: longitude must be within [-180,180], got %v", float64(input.Event.CastingContext.Longitude))
	}
	ys, err := liuyao.ParseYongShen(string(yongShen))
	if err != nil {
		return LiuyaoChartResultOutput{}, fmt.Errorf("liuyao.chart: %w", err)
	}
	chart := liuyao.ComputeChartAt(standard, float64(input.Event.CastingContext.Longitude), ys, casting.Yaos)
	view := liuyaoChartViewWire(chart)
	view.Analysis = liuyaoAnalysisWire(chart)
	return LiuyaoChartResultOutput{
		Event: LiuyaoCastingEvent{
			Casting: input.Event.Casting,
			CastingContext: LiuyaoCastingContext{
				StandardTime: input.Event.CastingContext.StandardTime,
				Longitude:    input.Event.CastingContext.Longitude,
				YongShen:     yongShen,
			},
		},
		Chart: view,
	}, nil
}

func liuyaoCastingDomain(output LiuyaoCastingOutput) (liuyao.Casting, error) {
	if len(output.Yaos) != 6 {
		return liuyao.Casting{}, fmt.Errorf("yaos requires exactly 6 values, got %d", len(output.Yaos))
	}
	casting := liuyao.Casting{
		Mode:    string(output.Mode),
		Order:   string(output.Order),
		Yaos:    [6]int{},
		DongYao: make([]int, 0, len(output.DongYao)),
	}
	for i, value := range output.Yaos {
		casting.Yaos[i] = int(value)
	}
	if output.CoinConvention != nil {
		casting.CoinConvention = &liuyao.CoinConvention{
			Heads: liuyao.CoinScore(output.CoinConvention.Heads),
			Tails: liuyao.CoinScore(output.CoinConvention.Tails),
		}
	}
	for _, round := range output.Rounds {
		coins := make([]liuyao.CoinFace, 0, 3)
		for _, coin := range round.Coins {
			coins = append(coins, liuyao.CoinFace(coin))
		}
		casting.Rounds = append(casting.Rounds, liuyao.CoinRound{
			Position: int(round.Position), Coins: coins, Value: liuyao.YaoType(round.Value),
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

func liuyaoLineView(line liuyao.Line, dayRelation liuyao.DayRelation, wangShuai ganzhi.WangShuai) LiuyaoChartLineView {
	flags := LiuyaoChartLineFlags{
		Moving: line.DongSelf, YuePo: line.YuePo, XunKong: line.XunKong,
		MuKu: line.MuKu, MuKuBranch: line.MuKuBranch, MuKuElement: line.MuKuElement,
		DongSheng: line.DongSheng, DongKe: line.DongKe,
	}
	view := LiuyaoChartLineView{
		Position: LiuyaoOptionalPosition(line.Position), YaoValue: LiuyaoOptionalYaoValue(line.Type),
		GanZhi: ganzhi.GanName(line.Gan) + ganzhi.ZhiName(line.Zhi),
		Wuxing: line.Wuxing.String(), LiuQin: line.LiuQin.String(),
		ShiYing: LiuyaoShiYing(line.ShiYing), LiuShou: line.LiuShou.String(),
		WangShuai: wangShuai.String(), ChangShengYue: line.ChangShengYue,
		Flags: flags,
	}
	if len(dayRelation.Relations) > 0 {
		view.RiChenRelations = map[string][]string{"relations": append([]string(nil), dayRelation.Relations...)}
	}
	return view
}

func liuyaoFuShenView(value *liuyao.FuShen) *LiuyaoFuShenView {
	if value == nil {
		return nil
	}
	return &LiuyaoFuShenView{
		Position: LiuyaoPosition(value.Position),
		LiuQin:   value.LiuQin.String(),
		Zhi:      value.Zhi,
	}
}

func liuyaoYongShenView(value liuyao.YongShenResult) LiuyaoYongShenView {
	view := LiuyaoYongShenView{
		Name:       string(value.Name),
		Position:   int(value.Position),
		IsHidden:   value.IsHidden,
		WangShuai:  value.WangShuai,
		YuePo:      value.YuePo,
		XunKong:    value.XunKong,
		MuKu:       value.MuKu,
		LiuShou:    value.LiuShou.String(),
		ChangSheng: value.ChangSheng,
		FuShen:     liuyaoFuShenView(value.FuShen),
	}
	for _, candidate := range value.FuShenCandidates {
		view.FuShenCandidates = append(view.FuShenCandidates, LiuyaoFuShenView{
			Position: LiuyaoPosition(candidate.Position),
			LiuQin:   candidate.LiuQin.String(),
			Zhi:      candidate.Zhi,
		})
	}
	return view
}

func liuyaoChartViewWire(chart liuyao.Chart) LiuyaoChartView {
	bianGua := LiuyaoHexagramName(liuyaoEncodedName(chart.BianGua))
	if int(chart.BianGua) == 0 {
		bianGua = ""
	}
	view := LiuyaoChartView{
		GuaName:      liuyaoEncodedName(chart.BenGua),
		Palace:       chart.Palace,
		PalaceWuxing: chart.PalaceWuxing.String(),
		Lines:        make([]LiuyaoChartLineView, 0, 6),
		BianLines:    make([]LiuyaoChartLineView, 0, 6),
		RiGan:        ganzhi.GanName(chart.RiGan),
		RiZhi:        ganzhi.ZhiName(chart.RiZhi),
		YueGan:       ganzhi.GanName(chart.YueGan),
		YueZhi:       ganzhi.ZhiName(chart.YueZhi),
		XunKong: []string{
			ganzhi.ZhiName(chart.XunKong[0]), ganzhi.ZhiName(chart.XunKong[1]),
		},
		Patterns: make([]LiuyaoChartPatternView, 0, len(chart.Patterns)),
		YongShen: liuyaoYongShenView(chart.YongShen),
	}
	if bianGua == "" {
		view.BianGua = nil
	} else {
		name := string(bianGua)
		view.BianGua = &name
	}
	for i, line := range chart.Lines {
		view.Lines = append(view.Lines, liuyaoLineView(line, chart.DayRelations[i], chart.WangShuai[i]))
	}
	for _, line := range chart.BianLines {
		if line.Position != 0 {
			view.BianLines = append(view.BianLines, liuyaoLineView(line, liuyao.DayRelation{}, ganzhi.WangShuai(-1)))
		}
	}
	for index, pattern := range chart.Patterns {
		view.Patterns = append(view.Patterns, LiuyaoChartPatternView{
			ID:          fmt.Sprintf("pattern-%d", index+1),
			PatternKind: string(pattern.Type), SubType: pattern.SubType,
			Position: LiuyaoOptionalPosition(pattern.Position), IsTrue: pattern.IsTrue,
		})
	}
	if chart.GuaCi.Name != "" {
		yaoCi := make([]string, 0, len(chart.GuaCi.YaoCi))
		yaoCi = append(yaoCi, chart.GuaCi.YaoCi[:]...)
		view.GuaCi = &LiuyaoGuaCiView{
			Name: liuyaoEncodedName(chart.GuaCi.Name), Text: chart.GuaCi.GuaCi, YaoCi: yaoCi,
		}
	}
	return view
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

func liuyaoAnalysisWire(chart liuyao.Chart) LiuyaoAnalysisOutput {
	output := LiuyaoAnalysisOutput{
		TombSchool: LiuyaoTombOutput{
			EarthBranch: LiuyaoTombBranch(chart.TombSchool.EarthBranch),
			School:      LiuyaoTombSchool(chart.TombSchool.School),
		},
	}
	for _, value := range chart.TimingCandidates {
		output.TimingCandidates = append(output.TimingCandidates, LiuyaoTimingOutput{
			ID:            LiuyaoTimingID(value.ID),
			Mechanism:     value.Mechanism,
			Position:      LiuyaoOptionalPosition(value.Position),
			TriggerBranch: ZhiName(value.TriggerBranch),
			Window:        value.Window,
			Basis:         append([]string(nil), value.Basis...),
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
			Gan: GanName(value.Gan), Zhi: ZhiName(value.Zhi),
			Wuxing: ElementName(value.Wuxing), FlyingPosition: LiuyaoPosition(value.FlyingPosition),
			FlyingLiuQin: LiuyaoLiuQinName(value.FlyingLiuQin), FlyingGan: GanName(value.FlyingGan),
			FlyingZhi: ZhiName(value.FlyingZhi), FlyingWuxing: ElementName(value.FlyingWuxing),
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
			Position: LiuyaoPosition(value.Position), FromBranch: ZhiName(value.FromBranch),
			ToBranch: ZhiName(value.ToBranch), ReturnRelation: value.ReturnRelation,
			AdvanceRetreat: value.AdvanceRetreat, ChangedVoid: value.ChangedVoid,
			ChangedBreak: value.ChangedBreak, Basis: append([]string(nil), value.Basis...),
		})
	}
	for _, value := range chart.SanHeCandidates {
		positions := make([]LiuyaoPosition, 0, len(value.Positions))
		for _, position := range value.Positions {
			positions = append(positions, LiuyaoPosition(position))
		}
		branches := make([]ZhiName, 0, len(value.Branches))
		for _, branch := range value.Branches {
			branches = append(branches, ZhiName(branch))
		}
		missing := make([]ZhiName, 0, len(value.Missing))
		for _, branch := range value.Missing {
			missing = append(missing, ZhiName(branch))
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
			branches := make([]ZhiName, 0, len(entry.Branches))
			for _, branch := range entry.Branches {
				branches = append(branches, ZhiName(branch))
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
			Position: LiuyaoOptionalPosition(value.Position), Branch: ZhiName(value.Branch),
			IsHidden: value.IsHidden, Selected: value.Selected,
			WangShuai: WangShuaiName(value.WangShuai), Reason: value.Reason,
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
		Mode:    LiuyaoCastMode(casting.Mode),
		Order:   LiuyaoOrderName(casting.Order),
		Yaos:    make(LiuyaoYaoValuesOutput, 0, len(casting.Yaos)),
		DongYao: make(LiuyaoPositions, 0, len(casting.DongYao)),
	}
	if casting.CoinConvention != nil {
		output.CoinConvention = &LiuyaoCoinConvention{
			Heads: LiuyaoHeadsScore(casting.CoinConvention.Heads),
			Tails: LiuyaoTailsScore(casting.CoinConvention.Tails),
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
