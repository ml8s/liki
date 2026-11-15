package mcpcontract

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"

	"liki-engine/internal/engine/liuyao"
)

func TestLiuyaoCoinScoresUseConstSchema(t *testing.T) {
	definition, err := testEngine().Definition("liuyao_qigua")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties struct {
			Event struct {
				Properties struct {
					Casting struct {
						OneOf []struct {
							Properties struct {
								CoinConvention struct {
									Properties struct {
										Heads map[string]any `json:"正"`
										Tails map[string]any `json:"反"`
									} `json:"properties"`
								} `json:"coin_convention"`
							} `json:"properties"`
						} `json:"oneOf"`
					} `json:"casting"`
				} `json:"properties"`
			} `json:"event"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(definition.OutputSchema, &schema); err != nil {
		t.Fatal(err)
	}
	var coins struct {
		Properties struct {
			CoinConvention struct {
				Properties struct {
					Heads map[string]any `json:"正"`
					Tails map[string]any `json:"反"`
				} `json:"properties"`
			} `json:"coin_convention"`
		} `json:"properties"`
	}
	for _, branch := range schema.Properties.Event.Properties.Casting.OneOf {
		if branch.Properties.CoinConvention.Properties.Heads != nil {
			if err := json.Unmarshal(mustJSON(t, branch), &coins); err != nil {
				t.Fatal(err)
			}
		}
	}
	heads := coins.Properties.CoinConvention.Properties.Heads
	tails := coins.Properties.CoinConvention.Properties.Tails
	for name, property := range map[string]map[string]any{
		"heads": heads,
		"tails": tails,
	} {
		if property["const"] == nil {
			t.Fatalf("%s score schema has no const: %v", name, property)
		}
		if _, hasEnum := property["enum"]; hasEnum {
			t.Fatalf("%s score schema must not use singleton enum: %v", name, property)
		}
	}
	if heads["const"] != float64(liuyao.CoinHeadsScore) || tails["const"] != float64(liuyao.CoinTailsScore) {
		t.Fatalf("coin score consts = heads %v tails %v", heads["const"], tails["const"])
	}
}

func TestLiuyaoQiguaCoinsInputUsesRawThreeCoinArrays(t *testing.T) {
	definition, err := testEngine().Definition("liuyao_qigua")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		OneOf []struct {
			Properties struct {
				Origin struct {
					Properties struct {
						Mode   map[string]any `json:"mode"`
						Rounds map[string]any `json:"rounds"`
					} `json:"properties"`
				} `json:"origin"`
			} `json:"properties"`
		} `json:"oneOf"`
	}
	if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
		t.Fatal(err)
	}

	var roundsSchema map[string]any
	for _, form := range schema.OneOf {
		if form.Properties.Origin.Properties.Mode["const"] == "coins" {
			roundsSchema = form.Properties.Origin.Properties.Rounds
			break
		}
	}
	if roundsSchema == nil {
		t.Fatal("coins form is missing rounds schema")
	}
	if roundsSchema["minItems"] != float64(6) || roundsSchema["maxItems"] != float64(6) {
		t.Fatalf("rounds bounds = %v, want exactly 6", roundsSchema)
	}
	items, ok := roundsSchema["items"].(map[string]any)
	if !ok {
		t.Fatalf("rounds items is not an object schema: %#v", roundsSchema["items"])
	}
	if items["type"] != "array" {
		t.Fatalf("round items type = %#v, want array", items["type"])
	}
	if items["minItems"] != float64(3) || items["maxItems"] != float64(3) {
		t.Fatalf("round item bounds = %#v, want exactly 3", items)
	}
	coinItems, ok := items["items"].(map[string]any)
	if !ok {
		t.Fatalf("coin items is not an object schema: %#v", items["items"])
	}
	enum, ok := coinItems["enum"].([]any)
	if !ok || len(enum) != 2 || enum[0] != "正" || enum[1] != "反" {
		t.Fatalf("coin item enum = %#v, want 正/反", coinItems["enum"])
	}
}

func TestLiuyaoFocusSchemaUsesClosedDiscriminatedBranches(t *testing.T) {
	definition, err := testEngine().Definition("liuyao_chart")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
		t.Fatal(err)
	}
	event := schema["properties"].(map[string]any)["event"].(map[string]any)
	context := event["properties"].(map[string]any)["casting_context"].(map[string]any)
	focusAny := context["properties"].(map[string]any)["focus"].(map[string]any)
	var focus struct {
		AdditionalProperties *bool `json:"additionalProperties"`
		OneOf                []struct {
			AdditionalProperties *bool    `json:"additionalProperties"`
			Required             []string `json:"required"`
			Properties           map[string]struct {
				Const string `json:"const"`
			} `json:"properties"`
		} `json:"oneOf"`
	}
	if err := json.Unmarshal(mustJSON(t, focusAny), &focus); err != nil {
		t.Fatal(err)
	}
	if focus.AdditionalProperties != nil && *focus.AdditionalProperties {
		t.Fatal("focus composition root must let closed branches own properties")
	}
	if len(focus.OneOf) != 3 {
		t.Fatalf("focus forms = %d, want 3", len(focus.OneOf))
	}
	for _, form := range focus.OneOf {
		if form.AdditionalProperties == nil || *form.AdditionalProperties {
			t.Fatalf("focus branch is not closed: %#v", form)
		}
	}
	matter, relationship, yongShen := focus.OneOf[0], focus.OneOf[1], focus.OneOf[2]
	if matter.Properties["kind"].Const != "matter" ||
		relationship.Properties["matter"].Const != "relationship" ||
		yongShen.Properties["kind"].Const != "yong_shen" {
		t.Fatalf("focus discriminators = %#v", focus.OneOf)
	}
	if !containsString(relationship.Required, "perspective") {
		t.Fatalf("relationship focus does not require perspective = %#v", relationship)
	}
}

func assertLiuyaoCanonicalEventSchema(t *testing.T, raw json.RawMessage) {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	event, ok := schema["properties"].(map[string]any)["event"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no event properties: %#v", schema)
	}
	eventProperties := event["properties"].(map[string]any)
	if _, exists := eventProperties["origin"]; exists {
		t.Fatalf("canonical event must not publish origin: %#v", eventProperties)
	}
	if _, exists := eventProperties["question"]; exists {
		t.Fatalf("canonical event must not publish question: %#v", eventProperties)
	}
	branches, ok := eventProperties["casting"].(map[string]any)["oneOf"].([]any)
	if !ok || len(branches) != 2 {
		t.Fatalf("casting branches = %#v, want 2", eventProperties["casting"])
	}
	context := eventProperties["casting_context"].(map[string]any)
	contextProperties := context["properties"].(map[string]any)
	if _, exists := contextProperties["focus"]; exists {
		t.Fatalf("canonical casting_context must not publish focus: %#v", contextProperties)
	}
	yongShen, ok := contextProperties["yong_shen"].(map[string]any)
	if !ok || (yongShen["type"] != "string" && yongShen["$ref"] == nil) {
		t.Fatalf("canonical casting_context.yong_shen = %#v", contextProperties["yong_shen"])
	}
	required, _ := context["required"].([]any)
	if !containsAnyString(required, "yong_shen") {
		t.Fatalf("canonical casting_context must require yong_shen: %#v", context)
	}
}

func assertLiuyaoLegacyInputEventSchema(t *testing.T, raw json.RawMessage) {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	event, ok := schema["properties"].(map[string]any)["event"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no event properties: %#v", schema)
	}
	eventProperties := event["properties"].(map[string]any)
	for field, want := range map[string]int{"origin": 3, "casting": 2} {
		branches, ok := eventProperties[field].(map[string]any)["oneOf"].([]any)
		if !ok || len(branches) != want {
			t.Fatalf("%s branches = %#v, want %d", field, eventProperties[field], want)
		}
	}
	context := eventProperties["casting_context"].(map[string]any)
	focus, ok := context["properties"].(map[string]any)["focus"].(map[string]any)
	if ok {
		branches, ok := focus["oneOf"].([]any)
		if !ok || len(branches) != 3 {
			t.Fatalf("legacy focus branches = %#v, want 3", focus)
		}
	}
}

func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestLiuyaoEventSchemaSeparatesCanonicalOutputFromLegacyInput(t *testing.T) {
	registry := testEngine()
	qigua, err := registry.Definition("liuyao_qigua")
	if err != nil {
		t.Fatal(err)
	}
	assertLiuyaoCanonicalEventSchema(t, qigua.OutputSchema)

	chart, err := registry.Definition("liuyao_chart")
	if err != nil {
		t.Fatal(err)
	}
	assertLiuyaoLegacyInputEventSchema(t, chart.InputSchema)
	assertLiuyaoCanonicalEventSchema(t, chart.OutputSchema)
}

func TestLiuyaoEventSchemaRejectsOriginReceiptMismatch(t *testing.T) {
	registry := testEngine()
	casting := LiuyaoCastingOutput{
		Mode: "yaos", Order: "bottom_up",
		Yaos: LiuyaoYaoValuesOutput{7, 8, 8, 7, 8, 9}, DongYao: LiuyaoPositions{6},
	}
	castingContext := LiuyaoCastingContextInputCompat{
		StandardTime: standardInputFromInstant(t, "2026-07-31T02:00:00Z", "Asia/Shanghai"),
		Longitude:    116.4,
		Focus:        &LiuyaoFocusInput{Kind: "matter", Matter: "career"},
	}
	origin := LiuyaoCastingRequestInput{Mode: "yaos", Yaos: LiuyaoYaoValuesInput{7, 8, 8, 7, 8, 9}}
	valid := LiuyaoCastingEventInput{
		Origin:         &origin,
		Casting:        casting,
		CastingContext: castingContext,
	}
	if _, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, LiuyaoChartInput{Event: &valid})); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}

	missingRounds := valid
	missingOrigin := LiuyaoCastingRequestInput{Mode: "coins"}
	missingRounds.Origin = &missingOrigin
	if _, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, LiuyaoChartInput{Event: &missingRounds})); err == nil {
		t.Fatal("coins origin without rounds passed the chart input schema")
	}

	if _, err := registry.Call(context.Background(), "liuyao_qigua", mustJSON(t, LiuyaoQiguaInput{
		Origin: LiuyaoCastingRequestInput{
			Mode:   "auto",
			Rounds: LiuyaoCoinRoundsInput{{"正", "正", "反"}},
		},
		CastingContext: testCastingContext(t),
	})); err == nil {
		t.Fatal("auto origin with rounds passed the qigua input schema")
	}
}

func TestLiuyaoQiguaReceiptFeedsChartForAutoAndCoins(t *testing.T) {
	registry := testEngine()
	callChart := func(t *testing.T, event LiuyaoCastingEvent) {
		t.Helper()
		input := liuyaoCastingEventInput(event)
		raw, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, LiuyaoChartInput{Event: input}))
		if err != nil {
			t.Fatalf("liuyao_chart: %v", err)
		}
		if err := registry.ValidateOutput("liuyao_chart", raw); err != nil {
			t.Fatalf("liuyao_chart output contract: %v", err)
		}
	}

	autoRaw, err := registry.Call(context.Background(), "liuyao_qigua", mustJSON(t, LiuyaoQiguaInput{
		Origin:         LiuyaoCastingRequestInput{Mode: "auto"},
		CastingContext: testCastingContext(t),
	}))
	if err != nil {
		t.Fatalf("liuyao_qigua auto: %v", err)
	}
	if err := registry.ValidateOutput("liuyao_qigua", autoRaw); err != nil {
		t.Fatalf("liuyao_qigua auto output contract: %v", err)
	}
	var auto LiuyaoQiguaOutput
	if err := json.Unmarshal(autoRaw, &auto); err != nil {
		t.Fatal(err)
	}
	if auto.Event.Casting.Mode != "coins" || auto.Event.Casting.CoinConvention == nil {
		t.Fatalf("auto casting is not a coins receipt: %#v", auto.Event.Casting)
	}
	if *auto.Event.Casting.CoinConvention != (LiuyaoCoinConvention{
		Heads: LiuyaoHeadsScore(liuyao.CoinHeadsScore),
		Tails: LiuyaoTailsScore(liuyao.CoinTailsScore),
	}) {
		t.Fatalf("unexpected coin convention: %#v", auto.Event.Casting.CoinConvention)
	}
	callChart(t, auto.Event)

	coinsRaw, err := registry.Call(context.Background(), "liuyao_qigua", mustJSON(t, LiuyaoQiguaInput{
		Origin: LiuyaoCastingRequestInput{
			Mode: "coins",
			Rounds: LiuyaoCoinRoundsInput{
				{"正", "正", "反"},
				{"正", "反", "反"},
				{"反", "反", "反"},
				{"正", "正", "正"},
				{"正", "反", "反"},
				{"正", "正", "反"},
			},
		},
		CastingContext: testCastingContext(t),
	}))
	if err != nil {
		t.Fatalf("liuyao_qigua coins: %v", err)
	}
	if err := registry.ValidateOutput("liuyao_qigua", coinsRaw); err != nil {
		t.Fatalf("liuyao_qigua coins output contract: %v", err)
	}
	var coins LiuyaoQiguaOutput
	if err := json.Unmarshal(coinsRaw, &coins); err != nil {
		t.Fatal(err)
	}
	if len(coins.Event.Casting.Rounds) != 6 || !equalYaoValues(coins.Event.Casting.Yaos, [6]int{8, 7, 6, 9, 7, 8}) {
		t.Fatalf("unexpected coins receipt: %#v", coins.Event.Casting)
	}
	callChart(t, coins.Event)
}

func TestLiuyaoChartResolvesFocusDeterministically(t *testing.T) {
	registry := testEngine()
	casting := LiuyaoCastingOutput{
		Mode:    "yaos",
		Order:   "bottom_up",
		Yaos:    LiuyaoYaoValuesOutput{7, 8, 7, 8, 7, 8},
		DongYao: LiuyaoPositions{},
	}
	tests := []struct {
		name   string
		focus  LiuyaoFocusInput
		want   LiuyaoYongShenName
		errSub string
	}{
		{name: "career", focus: LiuyaoFocusInput{Kind: "matter", Matter: "career"}, want: "官鬼"},
		{name: "male relationship", focus: LiuyaoFocusInput{Kind: "matter", Matter: "relationship", Perspective: "male"}, want: "妻财"},
		{name: "female relationship", focus: LiuyaoFocusInput{Kind: "matter", Matter: "relationship", Perspective: "female"}, want: "官鬼"},
		{name: "explicit", focus: LiuyaoFocusInput{Kind: "yong_shen", YongShen: "子孙"}, want: "子孙"},
		{name: "missing perspective", focus: LiuyaoFocusInput{Kind: "matter", Matter: "relationship"}, errSub: "requires perspective"},
		{name: "unexpected perspective", focus: LiuyaoFocusInput{Kind: "matter", Matter: "career", Perspective: "male"}, errSub: "only valid for relationship"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			origin := LiuyaoCastingRequestInput{Mode: "yaos", Yaos: LiuyaoYaoValuesInput{7, 8, 7, 8, 7, 8}}
			input := LiuyaoChartInput{Event: &LiuyaoCastingEventInput{
				Origin:  &origin,
				Casting: casting,
				CastingContext: LiuyaoCastingContextInputCompat{
					StandardTime: standardInputFromInstant(t, "2026-07-31T02:00:00Z", "Asia/Shanghai"),
					Longitude:    116.4,
					Focus:        &test.focus,
				},
			}}
			raw, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, input))
			if test.errSub != "" {
				if err == nil || !strings.Contains(err.Error(), test.errSub) {
					t.Fatalf("focus error = %v, want contains %q", err, test.errSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("liuyao_chart: %v", err)
			}
			var result LiuyaoChartResultOutput
			if err := json.Unmarshal(raw, &result); err != nil {
				t.Fatal(err)
			}
			if result.Chart.YongShen.Name != string(test.want) {
				t.Fatalf("resolved yongshen = %q, want %q", result.Chart.YongShen.Name, test.want)
			}
			if !equalYaoValues(result.Event.Casting.Yaos, [6]int{7, 8, 7, 8, 7, 8}) ||
				result.Event.Casting.Mode != casting.Mode ||
				result.Chart.GuaName == "" || len(result.Chart.Lines) != 6 {
				t.Fatalf("chart result did not preserve case facts: %#v", result)
			}
		})
	}
}

func TestLiuyaoChartUsesEventYongShenWithoutFocus(t *testing.T) {
	registry := testEngine()
	casting := LiuyaoCastingOutput{
		Mode:    "yaos",
		Order:   "bottom_up",
		Yaos:    LiuyaoYaoValuesOutput{7, 8, 7, 8, 7, 8},
		DongYao: LiuyaoPositions{},
	}
	event := LiuyaoCastingEventInput{
		Casting: casting,
		CastingContext: LiuyaoCastingContextInputCompat{
			StandardTime: standardInputFromInstant(t, "2026-07-31T02:00:00Z", "Asia/Shanghai"),
			Longitude:    116.4,
			YongShen:     "子孙",
		},
	}
	raw, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, LiuyaoChartInput{Event: &event}))
	if err != nil {
		t.Fatalf("liuyao_chart: %v", err)
	}
	var result LiuyaoChartResultOutput
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if result.Chart.YongShen.Name != "子孙" || result.Event.CastingContext.YongShen != "子孙" {
		t.Fatalf("event yong_shen was not used: %#v", result)
	}

	conflict := event
	conflict.CastingContext.Focus = &LiuyaoFocusInput{Kind: "matter", Matter: "career"}
	if _, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, LiuyaoChartInput{Event: &conflict})); err == nil ||
		!strings.Contains(err.Error(), "both yong_shen and focus") {
		t.Fatalf("conflicting yong_shen/focus error = %v", err)
	}
}

func TestLiuyaoQiguaRejectsEmptyModeWithoutLegacyDefault(t *testing.T) {
	_, err := LiuyaoQigua(context.Background(), LiuyaoQiguaInput{
		Origin: LiuyaoCastingRequestInput{Mode: ""},
		CastingContext: LiuyaoCastingContextInput{
			StandardTime: "2026-07-31T10:00:00+08:00",
			Longitude:    116.4,
			Focus: LiuyaoFocusInput{
				Kind:   "matter",
				Matter: "career",
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("empty origin mode error = %v, want unsupported", err)
	}
}

func TestSupportsFastInputSchemaRejectsNumericEnums(t *testing.T) {
	schema := &jsonschema.Schema{Type: "integer", Enum: []any{3}}
	if supportsFastInputSchema(schema) {
		t.Fatal("numeric enum must use the canonical JSON Schema validator")
	}
}

func equalYaoValues(values LiuyaoYaoValuesOutput, want [6]int) bool {
	if len(values) != len(want) {
		return false
	}
	for i, value := range values {
		if int(value) != want[i] {
			return false
		}
	}
	return true
}

func testCastingContext(t *testing.T) LiuyaoCastingContextInput {
	t.Helper()
	return LiuyaoCastingContextInput{
		StandardTime: standardInputFromInstant(t, "2026-07-31T02:00:00Z", "Asia/Shanghai"),
		Longitude:    116.4,
		Focus: LiuyaoFocusInput{
			Kind:   "matter",
			Matter: "career",
		},
	}
}

func liuyaoCastingEventInput(event LiuyaoCastingEvent) *LiuyaoCastingEventInput {
	return &LiuyaoCastingEventInput{
		Casting: event.Casting,
		CastingContext: LiuyaoCastingContextInputCompat{
			StandardTime: event.CastingContext.StandardTime,
			Longitude:    event.CastingContext.Longitude,
			YongShen:     string(event.CastingContext.YongShen),
		},
	}
}
