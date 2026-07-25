package qiming

// Evaluation is the output of evaluating a single name.
type Evaluation struct {
	Name        string           `json:"name,omitempty"`
	Surname     string           `json:"surname"`
	GivenName   string           `json:"given_name"`
	Characters  []Character `json:"characters"`
	WuGe        WuGe             `json:"wuge"`
	SanCai      SanCai           `json:"sancai"`
	Phonetic    Phonetic     `json:"phonetic"`
	WuxingMatch bool             `json:"wuxing_match"`
	Wuxing      *struct {
		Yong bool `json:"yong"`
		Xi   bool `json:"xi,omitempty"`
		Ji   bool `json:"ji,omitempty"`
	} `json:"wuxing,omitempty"`
}


