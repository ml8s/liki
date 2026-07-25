package ganzhi

import (
	"encoding/json"
	"fmt"
)

// ShiShen classifies the ten-god relationship (十神).
type ShiShen int

const (
	ShiShenBiJian    ShiShen = iota // 比肩
	ShiShenJieCai                  // 劫财
	ShiShenShiShen                 // 食神
	ShiShenShangGuan               // 伤官
	ShiShenPianCai                 // 偏财
	ShiShenZhengCai                // 正财
	ShiShenQiSha                   // 七杀
	ShiShenZhengGuan               // 正官
	ShiShenPianYin                 // 偏印
	ShiShenZhengYin                // 正印
)

var shiShenNamesZH = [10]string{
	"比肩", "劫财", "食神", "伤官", "偏财",
	"正财", "七杀", "正官", "偏印", "正印",
}

func (tg ShiShen) String() string {
	if tg >= 0 && int(tg) < len(shiShenNamesZH) {
		return shiShenNamesZH[tg]
	}
	return ""
}

func (tg ShiShen) MarshalJSON() ([]byte, error) {
	return json.Marshal(tg.String())
}

func (tg *ShiShen) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		parsed, err := ParseShiShen(name)
		if err != nil {
			return fmt.Errorf("unknown shishen: %q", name)
		}
		*tg = parsed
		return nil
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		if i < 0 || i > int(ShiShenZhengYin) {
			return fmt.Errorf("shishen value %d out of range [0,9]", i)
		}
		*tg = ShiShen(i)
		return nil
	}
	return fmt.Errorf("cannot unmarshal %q as ShiShen", string(data))
}

// ParseShiShen converts a Chinese ten-god name to a ShiShen value.
func ParseShiShen(s string) (ShiShen, error) {
	for i, name := range shiShenNamesZH {
		if name == s {
			return ShiShen(i), nil
		}
	}
	return -1, fmt.Errorf("unknown shi_shen: %q", s)
}

// ShiShenName returns the Chinese name for a ten god type.
func ShiShenName(tg ShiShen) string { return tg.String() }

// ShiShenFromGan returns the ShiShen for another stem relative to the day master.
func ShiShenFromGan(riYuan, other Gan) ShiShen {
	dmElem := GanWuxing(riYuan)
	otherElem := GanWuxing(other)
	dmYY := GanYinYang(riYuan)
	otherYY := GanYinYang(other)
	return ShiShenType(dmElem, dmYY, otherElem, otherYY)
}

// ShiShenType classifies the ten-god relationship between day master and another stem.
func ShiShenType(dmElem Wuxing, dmYY YinYang, otherElem Wuxing, otherYY YinYang) ShiShen {
	switch {
	case dmElem == otherElem:
		if dmYY == otherYY {
			return ShiShenBiJian
		}
		return ShiShenJieCai
	case Sheng(dmElem, otherElem):
		if dmYY == otherYY {
			return ShiShenShiShen
		}
		return ShiShenShangGuan
	case Sheng(otherElem, dmElem):
		if dmYY == otherYY {
			return ShiShenPianYin
		}
		return ShiShenZhengYin
	case Ke(dmElem, otherElem):
		if dmYY == otherYY {
			return ShiShenPianCai
		}
		return ShiShenZhengCai
	default:
		if dmYY == otherYY {
			return ShiShenQiSha
		}
		return ShiShenZhengGuan
	}
}
