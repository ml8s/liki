package qiming

// BuildNames 从过滤后的 combos 组合出 given name（不含姓）。
// 只处理双名（有 second 的 combo）：first × second 笛卡尔积。
// 单名（无 second）跳过——单名由 LLM 直接选字，不走 build。
func BuildNames(combos []Combo) []string {
	seen := make(map[string]bool)
	var names []string
	for _, c := range combos {
		if len(c.Second) == 0 {
			continue // 单名跳过
		}
		for _, f := range c.First {
			for _, s := range c.Second {
				name := f + s
				if seen[name] {
					continue
				}
				names = append(names, name)
				seen[name] = true
			}
		}
	}
	return names
}
