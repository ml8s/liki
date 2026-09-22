package liuyao

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const castingSchemaVersion = "liuyao-cast-v1"

// CoinRound is one cast of three coins. Coins keeps the normalized symbols in
// the order supplied by the caller; it is an audit record, not an input to the
// Yin/Yang conversion (that only depends on the sum).
type CoinRound struct {
	Position int      `json:"position"`
	Coins    []string `json:"coins"`
	Value    int      `json:"value"`
	Label    string   `json:"label"`
	Changing bool     `json:"changing"`
}

// Casting is the complete, auditable input to one divination.
type Casting struct {
	SchemaVersion  string         `json:"schema_version"`
	Mode           string         `json:"mode"`
	Order          string         `json:"order"`
	CoinConvention map[string]int `json:"coin_convention,omitempty"`
	Rounds         []CoinRound    `json:"rounds,omitempty"`
	Yaos           [6]int         `json:"yaos"`
	DongYao        []int          `json:"dong_yao"`
	CastingID      string         `json:"casting_id"`
}

const (
	coinHeads = "正"
	coinTails = "反"
)

func yaoLabel(v YaoType) string {
	switch v {
	case LaoYin:
		return "老阴"
	case ShaoYang:
		return "少阳"
	case ShaoYin:
		return "少阴"
	case LaoYang:
		return "老阳"
	default:
		return ""
	}
}

func coinValue(symbol string) (int, error) {
	switch symbol {
	case coinHeads:
		return 3, nil
	case coinTails:
		return 2, nil
	default:
		return 0, fmt.Errorf("invalid coin %q: must be 正 or 反", symbol)
	}
}

// NewCoinsCasting normalizes six three-coin rounds, bottom line first.
func NewCoinsCasting(rounds [][]string) (Casting, error) {
	if len(rounds) != 6 {
		return Casting{}, fmt.Errorf("coins requires exactly 6 rounds, got %d", len(rounds))
	}
	receipt := Casting{
		SchemaVersion:  castingSchemaVersion,
		Mode:           "coins",
		Order:          "bottom_up",
		CoinConvention: map[string]int{coinHeads: 3, coinTails: 2},
	}
	values := [6]YaoType{}
	for i, coins := range rounds {
		if len(coins) != 3 {
			return Casting{}, fmt.Errorf("round %d requires exactly 3 coins, got %d", i+1, len(coins))
		}
		sum := 0
		normalized := make([]string, 3)
		for j, coin := range coins {
			value, err := coinValue(coin)
			if err != nil {
				return Casting{}, fmt.Errorf("round %d coin %d: %w", i+1, j+1, err)
			}
			sum += value
			normalized[j] = coin
		}
		value := YaoType(sum)
		values[i] = value
		receipt.Rounds = append(receipt.Rounds, CoinRound{
			Position: i + 1,
			Coins:    normalized,
			Value:    int(value),
			Label:    yaoLabel(value),
			Changing: value.IsChanging(),
		})
	}
	receipt.Yaos = yaosToInts(values)
	receipt.DongYao = dongYao(values)
	receipt.CastingID = castingID(receipt)
	return receipt, nil
}

// NewValuesCasting records user-supplied 6/7/8/9 line values.
func NewValuesCasting(values [6]int) (Casting, error) {
	yts, err := validateValues(values)
	if err != nil {
		return Casting{}, err
	}
	receipt := Casting{
		SchemaVersion: castingSchemaVersion,
		Mode:          "yaos",
		Order:         "bottom_up",
		Yaos:          values,
		DongYao:       dongYao(yts),
	}
	receipt.CastingID = castingID(receipt)
	return receipt, nil
}

// Validate ensures a decoded casting can be bound to one and only one set of
// six lines. It recomputes moving lines and never trusts the caller's array.
func (c Casting) Validate() error {
	if c.SchemaVersion != castingSchemaVersion {
		return fmt.Errorf("unsupported casting schema_version %q", c.SchemaVersion)
	}
	if c.CastingID == "" {
		return fmt.Errorf("casting_id is required")
	}
	if c.Mode != "coins" && c.Mode != "yaos" {
		return fmt.Errorf("unsupported casting mode %q", c.Mode)
	}
	if c.Order != "bottom_up" {
		return fmt.Errorf("casting order must be bottom_up")
	}
	yts, err := validateValues(c.Yaos)
	if err != nil {
		return err
	}
	expected := dongYao(yts)
	if len(c.DongYao) != len(expected) {
		return fmt.Errorf("casting dong_yao mismatch: got %v, want %v", c.DongYao, expected)
	}
	for i, pos := range expected {
		if c.DongYao[i] != pos {
			return fmt.Errorf("casting dong_yao mismatch: got %v, want %v", c.DongYao, expected)
		}
	}
	switch c.Mode {
	case "coins":
		if len(c.Rounds) != 6 {
			return fmt.Errorf("coins casting requires exactly 6 rounds, got %d", len(c.Rounds))
		}
		if len(c.CoinConvention) != 2 ||
			c.CoinConvention[coinHeads] != 3 || c.CoinConvention[coinTails] != 2 {
			return fmt.Errorf("casting coin convention mismatch")
		}
		for i, round := range c.Rounds {
			if round.Position != i+1 {
				return fmt.Errorf("casting round position mismatch: got %d, want %d", round.Position, i+1)
			}
			if len(round.Coins) != 3 {
				return fmt.Errorf("casting round %d requires exactly 3 coins, got %d", i+1, len(round.Coins))
			}
			sum := 0
			for _, coin := range round.Coins {
				value, err := coinValue(coin)
				if err != nil {
					return fmt.Errorf("casting round %d coin: %w", i+1, err)
				}
				sum += value
			}
			if round.Value != sum || int(yts[i]) != sum {
				return fmt.Errorf("casting round %d conflicts with yaos", i+1)
			}
			if round.Label != yaoLabel(yts[i]) || round.Changing != yts[i].IsChanging() {
				return fmt.Errorf("casting round %d audit mismatch", i+1)
			}
		}
	case "yaos":
		if len(c.Rounds) != 0 || len(c.CoinConvention) != 0 {
			return fmt.Errorf("yaos casting must not contain coin rounds")
		}
	}
	if c.CastingID != castingID(c) {
		return fmt.Errorf("casting_id mismatch")
	}
	return nil
}

func validateValues(values [6]int) ([6]YaoType, error) {
	var out [6]YaoType
	for i, value := range values {
		if value < 6 || value > 9 {
			return out, fmt.Errorf("yao[%d] = %d, must be 6-9", i, value)
		}
		out[i] = YaoType(value)
	}
	return out, nil
}

func randomCoinBit() (bool, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return false, err
	}
	return n.Int64() == 1, nil
}

// SecureCoins generates three fair coin symbols.
func SecureCoins() ([]string, error) {
	coins := make([]string, 3)
	for i := range coins {
		head, err := randomCoinBit()
		if err != nil {
			return nil, fmt.Errorf("secure coin generation: %w", err)
		}
		if head {
			coins[i] = coinHeads
		} else {
			coins[i] = coinTails
		}
	}
	return coins, nil
}

// SecureQigua casts six rounds using crypto/rand.
func SecureQigua() (Casting, error) {
	rounds := make([][]string, 6)
	for i := range rounds {
		coins, err := SecureCoins()
		if err != nil {
			return Casting{}, err
		}
		rounds[i] = coins
	}
	return NewCoinsCasting(rounds)
}

// castingID is the stable domain identity of one normalized cast. It excludes
// the derived casting_id itself and engineering telemetry.
func castingID(c Casting) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\x1f%s\x1f%s", c.SchemaVersion, c.Mode, c.Order)
	if len(c.Rounds) != 0 {
		for _, round := range c.Rounds {
			fmt.Fprintf(&b, "\x1f%d|%s", round.Position, strings.Join(round.Coins, ","))
		}
	} else {
		for _, value := range c.Yaos {
			fmt.Fprintf(&b, "\x1f%d", value)
		}
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
