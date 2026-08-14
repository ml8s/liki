package ganzhi

import (
	"encoding/json"
	"fmt"
)

// -- Symbol → Text -----------------------------------------------------------

// GanName returns the Chinese character for a heavenly stem (e.g. "甲").
func GanName(g Gan) string { return ganNames[g] }

// ZhiName returns the Chinese character for an earthly branch (e.g. "子").
func ZhiName(z Zhi) string { return zhiNames[z] }

func (g Gan) String() string { return GanName(g) }
func (z Zhi) String() string { return ZhiName(z) }

func (w Wuxing) String() string {
	switch w {
	case WxMu:
		return "木"
	case WxHuo:
		return "火"
	case WxTu:
		return "土"
	case WxJin:
		return "金"
	case WxShui:
		return "水"
	}
	return "未知"
}

func (g Gender) String() string { return string(g) }

func (y YinYang) String() string {
	if bool(y) {
		return "阳"
	}
	return "阴"
}

// -- Text → Symbol -----------------------------------------------------------

// ParseGan converts a Chinese stem name (e.g. "甲") to a Gan value.
func ParseGan(s string) (Gan, error) {
	for i := 1; i <= 10; i++ {
		if ganNames[i] == s {
			return Gan(i), nil
		}
	}
	return 0, fmt.Errorf("unknown gan: %q", s)
}

// ParseZhi converts a Chinese branch name (e.g. "子") to a Zhi value.
func ParseZhi(s string) (Zhi, error) {
	for i := 1; i <= 12; i++ {
		if zhiNames[i] == s {
			return Zhi(i), nil
		}
	}
	return 0, fmt.Errorf("unknown zhi: %q", s)
}

// ParseWuxing converts a Chinese or English element name to a Wuxing value.
func ParseWuxing(s string) (Wuxing, error) {
	switch s {
	case "木", "wood":
		return WxMu, nil
	case "火", "fire":
		return WxHuo, nil
	case "土", "earth":
		return WxTu, nil
	case "金", "metal":
		return WxJin, nil
	case "水", "water":
		return WxShui, nil
	}
	return 0, fmt.Errorf("unknown wuxing: %q", s)
}

// ParseGender converts a string to a Gender value.
func ParseGender(s string) (Gender, error) {
	switch s {
	case "male", "男":
		return Male, nil
	case "female", "女":
		return Female, nil
	}
	return "", fmt.Errorf("unknown gender: %q", s)
}

// ParseYinYang converts a Chinese yin/yang name to a YinYang value.
func ParseYinYang(s string) (YinYang, error) {
	switch s {
	case "阳":
		return Yang, nil
	case "阴":
		return Yin, nil
	}
	return false, fmt.Errorf("unknown yin yang: %q", s)
}

// -- WangShuai (五行旺衰) ------------------------------------------------------

// WangShuai classifies the five-phase prosperity state.
type WangShuai int

const (
	WSWang  WangShuai = iota // 旺
	WSXiang                  // 相
	WSXiu                    // 休
	WSQiu                    // 囚
	WSSi                     // 死
)

var wangShuaiNames = [5]string{"旺", "相", "休", "囚", "死"}

func (ws WangShuai) String() string {
	if ws >= 0 && int(ws) < len(wangShuaiNames) {
		return wangShuaiNames[ws]
	}
	return ""
}

func (ws WangShuai) MarshalJSON() ([]byte, error) { return json.Marshal(ws.String()) }

func (ws *WangShuai) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("wang_shuai: expected string, got %s", string(data))
	}
	for i, n := range wangShuaiNames {
		if n == s {
			*ws = WangShuai(i)
			return nil
		}
	}
	return fmt.Errorf("wang_shuai: unknown %q", s)
}

// ParseWangShuai converts a Chinese prosperity name to a WangShuai value.
func ParseWangShuai(s string) (WangShuai, error) {
	for i, name := range wangShuaiNames {
		if name == s {
			return WangShuai(i), nil
		}
	}
	return 0, fmt.Errorf("unknown wangshuai: %q", s)
}

// WangShuaiOf returns the five-phase prosperity state for a given element
// in a given solar month (branch 1-12).
//
// Rule: 当令者旺 / 令生者相 / 生令者休 / 克令者囚 / 令克者死
func WangShuaiOf(elem Wuxing, monthBranch Zhi) WangShuai {
	if monthBranch < 1 || monthBranch > 12 {
		return -1
	}
	mwx := ZhiWuxing(monthBranch)
	if elem == mwx {
		return WSWang
	}
	if Sheng(mwx, elem) {
		return WSXiang // 月令生该五行 → 令生者相
	}
	if Sheng(elem, mwx) {
		return WSXiu // 该五行生月令 → 生令者休
	}
	if Ke(elem, mwx) {
		return WSQiu // 该五行克月令 → 克令者囚
	}
	return WSSi // 月令克该五行 → 令克者死
}

// -- JSON --------------------------------------------------------------------

func (g Gan) MarshalJSON() ([]byte, error)    { return json.Marshal(GanName(g)) }
func (z Zhi) MarshalJSON() ([]byte, error)    { return json.Marshal(ZhiName(z)) }
func (w Wuxing) MarshalJSON() ([]byte, error) { return json.Marshal(w.String()) }

func (w *Wuxing) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("wuxing must be a string, got %s", string(data))
	}
	parsed, err := ParseWuxing(name)
	if err != nil {
		return fmt.Errorf("unknown wuxing: %q", name)
	}
	*w = parsed
	return nil
}

func (g *Gan) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("gan must be a string, got %s", string(data))
	}
	if name == "" {
		*g = 0
		return nil
	}
	parsed, err := ParseGan(name)
	if err != nil {
		return fmt.Errorf("unknown gan: %q", name)
	}
	*g = parsed
	return nil
}

func (z *Zhi) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("zhi must be a string, got %s", string(data))
	}
	if name == "" {
		*z = 0
		return nil
	}
	parsed, err := ParseZhi(name)
	if err != nil {
		return fmt.Errorf("unknown zhi: %q", name)
	}
	*z = parsed
	return nil
}

func (z Zhu) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Gan string `json:"gan"`
		Zhi string `json:"zhi"`
	}{Gan: GanName(z.Gan), Zhi: ZhiName(z.Zhi)})
}

// -- Labels ------------------------------------------------------------------

var zodiacNames = [13]string{
	"", "鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪",
}

func zhiLabel(s string, ok bool) string {
	if !ok {
		return "未知"
	}
	return s
}

func zodiac(z Zhi) (string, bool) {
	if int(z) >= 1 && int(z) <= 12 {
		return zodiacNames[z], true
	}
	return "", false
}

// ZodiacLabel returns the Chinese zodiac animal name, or "未知" if invalid.
func ZodiacLabel(z Zhi) string { s, ok := zodiac(z); return zhiLabel(s, ok) }

func zhiSeason(z Zhi) (string, bool) {
	switch int(z) {
	case 3, 4, 5:
		return "春", true
	case 6, 7, 8:
		return "夏", true
	case 9, 10, 11:
		return "秋", true
	case 12, 1, 2:
		return "冬", true
	}
	return "", false
}

// ZhiSeasonLabel returns the season label ("春"/"夏"/"秋"/"冬") for a zhi.
func ZhiSeasonLabel(z Zhi) string { s, ok := zhiSeason(z); return zhiLabel(s, ok) }

var lunarMonths = [13]string{
	"", "十一月", "十二月", "正月", "二月", "三月", "四月",
	"五月", "六月", "七月", "八月", "九月", "十月",
}

func zhiLunarMonth(z Zhi) (string, bool) {
	if int(z) >= 1 && int(z) <= 12 {
		return lunarMonths[z], true
	}
	return "", false
}

func zhiLunarMonthLabel(z Zhi) string { s, ok := zhiLunarMonth(z); return zhiLabel(s, ok) }

// ZhiLunarMonthLabel returns the lunar month label (e.g. "正月") for a zhi.
func ZhiLunarMonthLabel(z Zhi) string { return zhiLunarMonthLabel(z) }

func zhiHourRange(z Zhi) (string, bool) {
	if int(z) >= 1 && int(z) <= 12 {
		return HourRanges[z-1], true
	}
	return "", false
}

// ZhiHourRangeLabel returns the two-hour range for a zhi, or "未知" if invalid.
func ZhiHourRangeLabel(z Zhi) string { s, ok := zhiHourRange(z); return zhiLabel(s, ok) }
