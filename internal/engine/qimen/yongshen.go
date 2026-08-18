package qimen

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// QianShiType 占事类型（奇门问事）。
type QianShiType string

const (
	QianShiye    QianShiType = "事业"  // 事业/升迁/公职
	QianQiucai   QianShiType = "求财"  // 求财/生意
	QianHunyin   QianShiType = "婚姻"  // 婚姻/感情
	QianJiankang QianShiType = "健康"  // 健康/疾病
	QianSusong   QianShiType = "诉讼"  // 诉讼/官非
	QianXueye    QianShiType = "学业"  // 学业/考试
	QianChuxing  QianShiType = "出行"  // 出行/迁移
	QianYinshi   QianShiType = "隐藏"  // 隐藏/避世
	QianZonghe   QianShiType = "综合"  // 综合/当前运势
)

// ParseQianShi 解析占事类型。
func ParseQianShi(s string) (QianShiType, error) {
	switch s {
	case "事业", "升迁", "公职", "工作", "官运":
		return QianShiye, nil
	case "求财", "生意", "财运", "投资":
		return QianQiucai, nil
	case "婚姻", "感情", "恋爱", "婚恋":
		return QianHunyin, nil
	case "健康", "疾病", "身体", "病":
		return QianJiankang, nil
	case "诉讼", "官非", "纠纷", "打官司":
		return QianSusong, nil
	case "学业", "考试", "文书", "学习":
		return QianXueye, nil
	case "出行", "迁移", "外出", "旅行":
		return QianChuxing, nil
	case "隐藏", "避世", "脱身":
		return QianYinshi, nil
	case "综合", "运势", "当前":
		return QianZonghe, nil
	}
	return "", fmt.Errorf("未知占事类型 %q，可选：事业/求财/婚姻/健康/诉讼/学业/出行/隐藏/综合", s)
}

// YongShenBody 事象用神（奇门以门/星/神为事象符号）。
// 命理依据：《烟波钓叟歌》。奇门用神多维：求测人以日干/年命干，事象以门/星/神。
type YongShenBody struct {
	Door   *DoorIndex   `json:"door,omitempty"`    // 事象门（开门/生门/景门/死门/惊门/杜门）
	Star   *StarIndex   `json:"star,omitempty"`    // 事象星（天心/天辅/天芮）
	Spirit *SpiritIndex `json:"spirit,omitempty"`  // 事象神（六合等）
	Stem   *ganzhi.Gan  `json:"stem,omitempty"`    // 事象干（求财戊、婚姻庚/乙）
}

// YongShenResult 奇门用神领域对象（聚合求测人 + 事象用神 + 落宫状态）。
type YongShenResult struct {
	Name          string        `json:"name"`                  // 占事类型
	RiGanPalace   GongIndex     `json:"ri_gan_gong"`           // 日干落宫（求测人"我"）
	ShiGanPalace  GongIndex     `json:"shi_gan_gong"`          // 时干落宫（所问之事）
	NianGanPalace *GongIndex    `json:"nian_gan_gong,omitempty"` // 年命干落宫（本命根基，需 birth_year）
	Body          YongShenBody  `json:"body"`                  // 事象用神（门/星/神/干）
	BodyPalace    GongIndex     `json:"body_palace"`           // 事象用神落宫（门落宫为主）
	RiShiShengKe  string        `json:"ri_shi_sheng_ke"`       // 日干宫-时干宫生克（我 vs 事）
	KongWang      bool          `json:"kong_wang"`             // 用神落宫是否空亡
	MaXing        bool          `json:"ma_xing"`               // 用神落宫是否马星
}

// 婚姻用神：男看庚（男方），女看乙（女方），配六合神。
var hunyinYang = ganzhi.GanGeng // 男：庚
var hunyinYin = ganzhi.GanYi    // 女：乙

// qianshiBody 占事类型 → 事象用神（门/星/神/干）。
func qianshiBody(q QianShiType, gender string) YongShenBody {
	switch q {
	case QianShiye:
		star := StarTianXin
		door := DoorKai
		return YongShenBody{Door: &door, Star: &star}
	case QianQiucai:
		door := DoorSheng
		stem := ganzhi.GanWu // 戊为财
		return YongShenBody{Door: &door, Stem: &stem}
	case QianHunyin:
		spirit := SpiritLiuHe // 六合主婚姻
		if gender == "female" {
			return YongShenBody{Spirit: &spirit, Stem: &hunyinYin} // 女看乙
		}
		return YongShenBody{Spirit: &spirit, Stem: &hunyinYang} // 男看庚
	case QianJiankang:
		star := StarTianRui
		door := DoorSi
		return YongShenBody{Door: &door, Star: &star}
	case QianXueye:
		star := StarTianFu
		door := DoorJing
		return YongShenBody{Door: &door, Star: &star}
	case QianSusong:
		door := DoorJingMen
		return YongShenBody{Door: &door}
	case QianChuxing:
		door := DoorKai
		return YongShenBody{Door: &door}
	case QianYinshi:
		door := DoorDu
		return YongShenBody{Door: &door}
	default: // 综合/当前运势
		door := DoorKai
		return YongShenBody{Door: &door}
	}
}

// ComputeYongShen 聚合奇门用神（求测人 + 事象用神）。
// q: 占事类型；gender: 性别（male/female，婚姻用神分男女）。
func ComputeYongShen(chart Chart, q QianShiType, gender string) YongShenResult {
	return computeYongShen(chart, q, gender, 0, false)
}

// ComputeYongShenWithBirth 聚合奇门用神，并计算年命干落宫（需出生年份）。
func ComputeYongShenWithBirth(chart Chart, q QianShiType, gender string, birthYear int) YongShenResult {
	return computeYongShen(chart, q, gender, birthYear, true)
}

func computeYongShen(chart Chart, q QianShiType, gender string, birthYear int, hasBirth bool) YongShenResult {
	ys := YongShenResult{
		Name:          string(q),
		RiGanPalace:   chart.RiGanPalace,
		ShiGanPalace:  chart.ShiGanPalace,
		RiShiShengKe:  chart.RiShiShengKe,
		Body:          qianshiBody(q, gender),
	}

	// 事象用神落宫：门落宫为主，否则星/神落宫
	ys.BodyPalace = 0
	if ys.Body.Door != nil {
		ys.BodyPalace = findDoorPalaceIdx(chart.Pan, *ys.Body.Door)
	} else if ys.Body.Star != nil {
		ys.BodyPalace = findStarPalaceIdx(chart.Pan, *ys.Body.Star)
	} else if ys.Body.Spirit != nil {
		ys.BodyPalace = findSpiritPalaceIdx(chart.Pan, *ys.Body.Spirit)
	}

	// 用神落宫状态（空亡/马星）
	if ys.BodyPalace > 0 {
		for _, k := range chart.Pan.KongWang {
			if k == ys.BodyPalace {
				ys.KongWang = true
				break
			}
		}
		ys.MaXing = ys.BodyPalace == chart.Pan.MaXing
	}

	// 年命干落宫（需出生年份）
	if hasBirth && birthYear > 0 {
		// 出生年立春后的年柱 → 年干
		nian := tianwen.NianZhu(tianwen.GregorianTime(time.Date(birthYear, 2, 15, 0, 0, 0, 0, time.UTC)))
		if nian.Gan != 0 {
			palace := findGanPalaceIdx(chart.Pan, nian.Gan)
			if palace > 0 {
				ys.NianGanPalace = &palace
			}
		}
	}

	return ys
}

func findSpiritPalaceIdx(p pan, s SpiritIndex) GongIndex {
	for i, pg := range p.GongWei {
		if pg.Spirit == s {
			return GongIndex(i + 1)
		}
	}
	return 0
}

