package mcpcontract

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"

	"liki-engine/internal/engine/liuyao"
)

func TestLiuyaoCoinScoresUseConstSchema(t *testing.T) {
	definition, err := testEngine().Definition("liuyao_chart")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties struct {
			Casting struct {
				Properties struct {
					CoinConvention struct {
						Properties struct {
							Heads map[string]any `json:"正"`
							Tails map[string]any `json:"反"`
						} `json:"properties"`
					} `json:"coin_convention"`
				} `json:"properties"`
			} `json:"casting"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
		t.Fatal(err)
	}
	heads := schema.Properties.Casting.Properties.CoinConvention.Properties.Heads
	tails := schema.Properties.Casting.Properties.CoinConvention.Properties.Tails
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
				Casting struct {
					Properties struct {
						Mode   map[string]any `json:"mode"`
						Rounds map[string]any `json:"rounds"`
					} `json:"properties"`
				} `json:"casting"`
			} `json:"properties"`
		} `json:"oneOf"`
	}
	if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
		t.Fatal(err)
	}

	var roundsSchema map[string]any
	for _, form := range schema.OneOf {
		if form.Properties.Casting.Properties.Mode["const"] == "coins" {
			roundsSchema = form.Properties.Casting.Properties.Rounds
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

func TestLiuyaoQiguaReceiptFeedsChartForAutoAndCoins(t *testing.T) {
	registry := testEngine()
	chartInput := func(t *testing.T, casting LiuyaoCastingOutput) LiuyaoChartInput {
		t.Helper()
		return LiuyaoChartInput{
			StandardTime: standardInputFromInstant(t, "2026-07-31T02:00:00Z", "Asia/Shanghai"),
			Longitude:    116.4,
			Casting:      &casting,
		}
	}
	callChart := func(t *testing.T, casting LiuyaoCastingOutput) {
		t.Helper()
		raw, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, chartInput(t, casting)))
		if err != nil {
			t.Fatalf("liuyao_chart: %v", err)
		}
		if err := registry.ValidateOutput("liuyao_chart", raw); err != nil {
			t.Fatalf("liuyao_chart output contract: %v", err)
		}
	}

	autoRaw, err := registry.Call(context.Background(), "liuyao_qigua", mustJSON(t, LiuyaoQiguaInput{
		Casting: LiuyaoCastingRequestInput{Mode: "auto"},
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
	if auto.Casting.Mode != "coins" || auto.Casting.CoinConvention == nil {
		t.Fatalf("auto casting is not a coins receipt: %#v", auto.Casting)
	}
	if *auto.Casting.CoinConvention != (LiuyaoCoinConvention{
		Heads: LiuyaoHeadsScore(liuyao.CoinHeadsScore),
		Tails: LiuyaoTailsScore(liuyao.CoinTailsScore),
	}) {
		t.Fatalf("unexpected coin convention: %#v", auto.Casting.CoinConvention)
	}
	callChart(t, auto.Casting)

	coinsRaw, err := registry.Call(context.Background(), "liuyao_qigua", mustJSON(t, LiuyaoQiguaInput{
		Casting: LiuyaoCastingRequestInput{
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
	if len(coins.Casting.Rounds) != 6 || !equalYaoValues(coins.Casting.Yaos, [6]int{8, 7, 6, 9, 7, 8}) {
		t.Fatalf("unexpected coins receipt: %#v", coins.Casting)
	}
	callChart(t, coins.Casting)
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
