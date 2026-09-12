package liuyao

import "liki-engine/internal/engine/ganzhi"

func feiFuRelation(flying, hidden ganzhi.Wuxing) string {
	switch {
	case flying == hidden:
		return "飞伏比和"
	case ganzhi.Sheng(flying, hidden):
		return "飞来生伏"
	case ganzhi.Ke(flying, hidden):
		return "飞来克伏"
	case ganzhi.Sheng(hidden, flying):
		return "伏来生飞"
	case ganzhi.Ke(hidden, flying):
		return "伏来克飞"
	default:
		return ""
	}
}

func hiddenExitCondition(relation string) string {
	switch relation {
	case "飞来克伏":
		return "待飞神受冲或合化，伏神旺相方可出伏"
	case "飞来生伏":
		return "飞神生扶，待值日或引拔而出"
	case "伏来克飞":
		return "伏神有力克飞，待冲开飞神"
	case "伏来生飞":
		return "伏气外泄，待生扶值日"
	default:
		return "待值日或冲引而出"
	}
}

func computeHiddenLines(p *Chart) []HiddenLine {
	palaceBase := guaTable[p.BenGua].PalaceIdx * 8
	if palaceBase < 0 || palaceBase >= len(guaTable) {
		return nil
	}
	baseLines := zhuangGua(guaIndex(palaceBase), p.RiGan, false, p.PalaceWuxing)
	visible := map[LiuQin]bool{}
	for _, line := range p.Lines {
		visible[line.LiuQin] = true
	}

	var facts []HiddenLine
	for i := range baseLines {
		base := baseLines[i]
		if visible[base.LiuQin] {
			continue
		}
		flying := p.Lines[i]
		relation := feiFuRelation(flying.Wuxing, base.Wuxing)
		tombTypes := fuTombSources(p, base.Wuxing, base.Zhi)
		facts = append(facts, HiddenLine{
			Position:       i + 1,
			LiuQin:         base.LiuQin.String(),
			Gan:            ganzhi.GanName(base.Gan),
			Zhi:            ganzhi.ZhiName(base.Zhi),
			Wuxing:         base.Wuxing.String(),
			FlyingPosition: flying.Position,
			FlyingLiuQin:   flying.LiuQin.String(),
			FlyingGan:      ganzhi.GanName(flying.Gan),
			FlyingZhi:      ganzhi.ZhiName(flying.Zhi),
			FlyingWuxing:   flying.Wuxing.String(),
			FeiFuRelation:  relation,
			XunKong:        base.Zhi == p.XunKong[0] || base.Zhi == p.XunKong[1],
			YuePo:          ganzhi.IsLiuChong(base.Zhi, p.YueZhi),
			MuKu:           len(tombTypes) > 0,
			ExitCondition:  hiddenExitCondition(relation),
		})
	}
	return facts
}

func fuTombSources(p *Chart, element ganzhi.Wuxing, branch ganzhi.Zhi) []string {
	_ = branch
	tomb := tombOf(element)
	if tomb == 0 {
		return nil
	}
	var types []string
	if p.RiZhi == tomb {
		types = append(types, "day")
	}
	for _, position := range p.DongYao {
		if position >= 1 && position <= 6 && p.Lines[position-1].Zhi == tomb {
			types = append(types, "moving")
			break
		}
	}
	return types
}
