package liuyao

import "liki-engine/internal/engine/ganzhi"

// YongShen categorizes what the user is asking about.
type YongShen int

const (
	YongFumu    YongShen = iota // 父母: 长辈、文书、房屋
	YongXiongDi                 // 兄弟: 朋友、同事、竞争
	YongGuanGui                 // 官鬼: 工作、官运、疾病
	YongQiCai                   // 妻财: 财运、妻子、物品
	YongZiSun                   // 子孙: 子女、健康、宠物
	YongShiYao                  // 世爻: 自身、求问人
	YongYingYao                 // 应爻: 对方、外部环境
)

var yongShenNames = [7]string{"父母", "兄弟", "官鬼", "妻财", "子孙", "世爻", "应爻"}

func (y YongShen) String() string { return yongShenNames[y] }

// findYongShen finds the 用神 in the visible ben-gua only. The bian-gua is a
// transformation layer; it must never become the primary 用神 location. When
// the target is absent here the caller must use findFuShen.
func (p *Chart) findYongShen(typ YongShen) int {
	switch typ {
	case YongShiYao:
		return p.findShiYao()
	case YongYingYao:
		return p.findYingYao()
	}
	target := yongShenToLiuQin(typ)
	// 多现取旺: 取月建旺衰最佳者 (Wang=0 > Xiang=1 > Xiu=2 > Qiu=3 > Si=4).
	bestPos := 0
	bestWs := 999
	for _, l := range p.Lines {
		if l.LiuQin == target {
			ws := 999
			if l.Position > 0 && l.Position <= 6 {
				ws = int(p.WangShuai[l.Position-1])
			}
			if ws < bestWs {
				bestWs = ws
				bestPos = l.Position
			}
		}
	}
	return bestPos
}

func (p *Chart) findShiYao() int {
	for _, l := range p.Lines {
		if l.ShiYing == "世" {
			return l.Position
		}
	}
	return 0
}

func (p *Chart) findYingYao() int {
	for _, l := range p.Lines {
		if l.ShiYing == "应" {
			return l.Position
		}
	}
	return 0
}

// FuShen holds the 飞伏 information when 用神 is not present.
type FuShen struct {
	Position int    `json:"position"` // 爻位 1-6
	LiuQin   LiuQin `json:"liu_qin"`  // 伏神六亲
	Zhi      string `json:"zhi"`      // 伏神地支
}

// yongShenToLiuQin maps YongShen → LiuQin for the first 5 types.
func yongShenToLiuQin(typ YongShen) LiuQin {
	switch typ {
	case YongFumu:
		return QinFumu
	case YongXiongDi:
		return QinXiongDi
	case YongGuanGui:
		return QinGuanGui
	case YongQiCai:
		return QinQiCai
	case YongZiSun:
		return QinZiSun
	default:
		return -1
	}
}

// findFuShen finds the 伏神 when 用神 is not present.
func (p *Chart) findFuShen(typ YongShen) *FuShen {
	meta := guaTable[p.BenGua]
	palaceBase := meta.PalaceIdx * 8 // 本宫卦 = palace * 8
	baseMeta := guaTable[palaceBase]
	naZhi := naZhiTable[baseMeta.PalaceIdx]
	elem := palaceWuxing[baseMeta.PalaceIdx]

	target := yongShenToLiuQin(typ)

	// For each line of the palace base hexagram, check if it has the target六亲.
	for i := 0; i < 6; i++ {
		zhiWx := ganzhi.ZhiWuxing(naZhi[i])
		qin := computeLiuQin(zhiWx, elem)
		if qin == target {
			return &FuShen{
				Position: i + 1,
				LiuQin:   qin,
				Zhi:      ganzhi.ZhiName(naZhi[i]),
			}
		}
	}
	return nil
}

// fuShenZhi converts the already-validated public fu-shen branch name back to
// the internal branch value used by deterministic state derivation.
func fuShenZhi(fuShen *FuShen) ganzhi.Zhi {
	zhi, err := ganzhi.ParseZhi(fuShen.Zhi)
	if err != nil {
		return 0
	}
	return zhi
}
