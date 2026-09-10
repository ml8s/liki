package ziwei

// PalaceFact is a deterministic atomic fact for Ziwei factor queries. Python
// operators match these facts exactly; they do not reinterpret palace stars.
type PalaceFact struct {
	Palace string `json:"palace"`
	Kind   string `json:"kind"`
	Target string `json:"target"`
	Star   string `json:"star,omitempty"`
	Value  string `json:"value,omitempty"`
}

var (
	mainStars = map[starIndex]bool{
		ZiWei: true, TianJi: true, TaiYang: true, WuQu: true,
		TianTong: true, LianZhen: true, TianFu: true, TaiYin: true,
		TanLang: true, JuMen: true, TianXiang: true, TianLiang: true,
		QiSha: true, PoJun: true,
	}
	auspiciousStars = map[starIndex]bool{
		LuCun: true, TianKui: true, TianYue: true,
		ZuoFu: true, YouBi: true, WenChang: true, WenQu: true,
	}
	literaryStars      = map[starIndex]bool{WenChang: true, WenQu: true}
	palaceMaleficStars = map[starIndex]bool{
		QingYang: true, TuoLuo: true, HuoXing: true,
		LingXing: true, DiKong: true, DiJie: true,
	}
	brightnessGroups = map[string][]string{
		"庙旺": {"庙", "旺", "得"},
		"落陷": {"陷", "平"},
	}
)

func starGroups(star starIndex) []string {
	groups := []string{}
	if mainStars[star] {
		groups = append(groups, "紫微主星")
	}
	if auspiciousStars[star] {
		groups = append(groups, "紫微六吉星")
	}
	if literaryStars[star] {
		groups = append(groups, "紫微文星")
	}
	if palaceMaleficStars[star] {
		groups = append(groups, "煞星")
	}
	return groups
}

func computePalaceFacts(chart Chart) []PalaceFact {
	facts := make([]PalaceFact, 0, 128)
	add := func(palace, kind, target, star, value string) {
		facts = append(facts, PalaceFact{
			Palace: palace, Kind: kind, Target: target,
			Star: star, Value: value,
		})
	}

	for _, palace := range chart.GongWei {
		majorCount := 0
		for _, item := range palace.Stars {
			if !mainStars[item.Star] {
				continue
			}
			majorCount++
		}

		for _, item := range palace.Stars {
			add(palace.Name, "star", item.Name, item.Name, "")
			for _, group := range starGroups(item.Star) {
				add(palace.Name, "star", group, item.Name, "")
			}
			if item.SiHua != "" {
				add(palace.Name, "si_hua", string(item.SiHua), item.Name, "")
			}
			if item.Brightness == "" {
				continue
			}
			add(palace.Name, "brightness", item.Brightness, item.Name, item.Brightness)
			if mainStars[item.Star] {
				add(palace.Name, "brightness", "紫微主星", item.Name, item.Brightness)
			}
			for target, values := range brightnessGroups {
				for _, value := range values {
					if item.Brightness != value {
						continue
					}
					add(palace.Name, "brightness", target, item.Name, item.Brightness)
				}
			}
		}

		if majorCount == 0 {
			add(palace.Name, "special", "无主星", "", "true")
		}
		if majorCount == 1 {
			for _, item := range palace.Stars {
				if mainStars[item.Star] {
					add(palace.Name, "special", "唯一主星", item.Name, "true")
				}
			}
		}
	}
	return facts
}
