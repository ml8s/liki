package qiming

// SurnameStrokes 姓氏康熙笔画信息，用于五格/三才计算。
type SurnameStrokes struct {
	Total    int  // 姓氏全部笔画之和（复姓天格/总格用）
	Last     int  // 姓氏最后一字笔画（人格用）
	Compound bool // 是否复姓（天格不加 1）
}

// Evaluation is the output of evaluating a single name.
type Evaluation struct {
	Name        string           `json:"name,omitempty"`
	Surname     string           `json:"surname"`
	GivenName   string           `json:"given_name"`
	Characters  []Character `json:"characters"`
	WuGe        *WuGe            `json:"wuge,omitempty"`
	SanCai      *SanCai          `json:"sancai,omitempty"`
	Phonetic    Phonetic     `json:"phonetic"`
	WuxingMatch bool             `json:"wuxing_match"`
	Wuxing      *struct {
		Yong bool `json:"yong"`
		Xi   bool `json:"xi,omitempty"`
		Ji   bool `json:"ji,omitempty"`
	} `json:"wuxing,omitempty"`
}


