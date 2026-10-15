package mcpcontract

import (
	"context"
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
	SchemaVersion  LiuyaoSchemaVersion         `json:"schema_version" jsonschema:"casting receipt schema 版本。"`
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

func liuyaoCastingWire(casting liuyao.Casting) LiuyaoCastingOutput {
	output := LiuyaoCastingOutput{
		SchemaVersion: LiuyaoSchemaVersion(casting.SchemaVersion),
		Mode:          LiuyaoCastMode(casting.Mode),
		Order:         LiuyaoOrderName(casting.Order),
		Yaos:          make(LiuyaoYaoValuesOutput, 0, len(casting.Yaos)),
		DongYao:       make(LiuyaoPositions, 0, len(casting.DongYao)),
		CastingID:     LiuyaoCastingID(casting.CastingID),
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
