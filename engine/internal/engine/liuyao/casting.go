package liuyao

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// CoinRound is one cast of three coins. Coins keeps the normalized symbols in
// the order supplied by the caller; only their sum drives Yin/Yang.
type CoinRound struct {
	Position int        `json:"position"`
	Coins    []CoinFace `json:"coins"`
	Value    YaoType    `json:"yao_value"`
	Label    string     `json:"yao_label"`
	Changing bool       `json:"is_dong_yao"`
}

// Casting is the complete source-fact input to one divination.
type Casting struct {
	Mode           string          `json:"mode"`
	Order          string          `json:"order"`
	CoinConvention *CoinConvention `json:"coin_convention,omitempty"`
	Rounds         []CoinRound     `json:"rounds,omitempty"`
	Yaos           [6]int          `json:"yaos"`
	DongYao        []int           `json:"dong_yao"`
}

// CoinFace is a physical coin outcome; CoinScore is its fixed counting value.
type CoinFace string
type CoinScore int

const (
	CoinFaceHeads  CoinFace  = "正"
	CoinFaceTails  CoinFace  = "反"
	CoinHeadsScore CoinScore = 3
	CoinTailsScore CoinScore = 2
)

// CoinConvention records the counting convention as an auditable casting fact.
type CoinConvention struct {
	Heads CoinScore `json:"正"`
	Tails CoinScore `json:"反"`
}

func defaultCoinConvention() *CoinConvention {
	return &CoinConvention{Heads: CoinHeadsScore, Tails: CoinTailsScore}
}

func coinValue(symbol CoinFace) (CoinScore, error) {
	switch symbol {
	case CoinFaceHeads:
		return CoinHeadsScore, nil
	case CoinFaceTails:
		return CoinTailsScore, nil
	default:
		return 0, fmt.Errorf("invalid coin %q: must be 正 or 反", string(symbol))
	}
}

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

// NewCoinsCasting normalizes six three-coin rounds, bottom line first.
func NewCoinsCasting(rounds [][]CoinFace) (Casting, error) {
	if len(rounds) != 6 {
		return Casting{}, fmt.Errorf("coins requires exactly 6 rounds, got %d", len(rounds))
	}
	receipt := Casting{
		Mode:           "coins",
		Order:          "bottom_up",
		CoinConvention: defaultCoinConvention(),
	}
	values := [6]YaoType{}
	for i, coins := range rounds {
		if len(coins) != 3 {
			return Casting{}, fmt.Errorf("round %d requires exactly 3 coins, got %d", i+1, len(coins))
		}
		var sum CoinScore
		normalized := make([]CoinFace, 3)
		for j, coin := range coins {
			value, err := coinValue(coin)
			if err != nil {
				return Casting{}, fmt.Errorf("round %d coin %d: %w", i+1, j+1, err)
			}
			sum += value
			normalized[j] = coin
		}
		value := YaoType(int(sum))
		values[i] = value
		receipt.Rounds = append(receipt.Rounds, CoinRound{
			Position: i + 1,
			Coins:    normalized,
			Value:    value,
			Label:    yaoLabel(value),
			Changing: value.IsChanging(),
		})
	}
	receipt.Yaos = yaosToInts(values)
	receipt.DongYao = dongYao(values)
	return receipt, nil
}

// NewValuesCasting records user-supplied 6/7/8/9 line values.
func NewValuesCasting(values [6]int) (Casting, error) {
	yts, err := validateValues(values)
	if err != nil {
		return Casting{}, err
	}
	receipt := Casting{
		Mode:    "yaos",
		Order:   "bottom_up",
		Yaos:    values,
		DongYao: dongYao(yts),
	}
	return receipt, nil
}

// Validate ensures a decoded casting describes exactly one set of six lines.
// It recomputes moving lines and rejects internally inconsistent source facts.
func (c Casting) Validate() error {
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
		if c.CoinConvention == nil || *c.CoinConvention != (CoinConvention{
			Heads: CoinHeadsScore,
			Tails: CoinTailsScore,
		}) {
			return fmt.Errorf("casting coin convention mismatch")
		}
		for i, round := range c.Rounds {
			if round.Position != i+1 {
				return fmt.Errorf("casting round position mismatch: got %d, want %d", round.Position, i+1)
			}
			if len(round.Coins) != 3 {
				return fmt.Errorf("casting round %d requires exactly 3 coins, got %d", i+1, len(round.Coins))
			}
			var sum CoinScore
			for _, coin := range round.Coins {
				value, err := coinValue(coin)
				if err != nil {
					return fmt.Errorf("casting round %d coin: %w", i+1, err)
				}
				sum += value
			}
			if round.Value != YaoType(sum) || yts[i] != YaoType(sum) {
				return fmt.Errorf("casting round %d conflicts with yaos", i+1)
			}
			if round.Label != yaoLabel(yts[i]) || round.Changing != yts[i].IsChanging() {
				return fmt.Errorf("casting round %d fact mismatch", i+1)
			}
		}
	case "yaos":
		if len(c.Rounds) != 0 || c.CoinConvention != nil {
			return fmt.Errorf("yaos casting must not contain coin rounds")
		}
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

// secureCoins generates three fair coin symbols.
func secureCoins() ([]CoinFace, error) {
	coins := make([]CoinFace, 3)
	for i := range coins {
		head, err := randomCoinBit()
		if err != nil {
			return nil, fmt.Errorf("secure coin generation: %w", err)
		}
		if head {
			coins[i] = CoinFaceHeads
		} else {
			coins[i] = CoinFaceTails
		}
	}
	return coins, nil
}

// SecureQigua casts six rounds using crypto/rand.
func SecureQigua() (Casting, error) {
	rounds := make([][]CoinFace, 6)
	for i := range rounds {
		coins, err := secureCoins()
		if err != nil {
			return Casting{}, err
		}
		rounds[i] = coins
	}
	return NewCoinsCasting(rounds)
}
