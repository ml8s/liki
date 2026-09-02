package qiming

// CandidatePool is one ordered pool of naming characters.
type CandidatePool struct {
	Slot  string   `json:"slot"`
	Chars []string `json:"chars"`
}

// PickResult is the character pool output for a naming request.
type PickResult struct {
	Wuxing1 string          `json:"wuxing1"`
	Wuxing2 string          `json:"wuxing2,omitempty"`
	Pools   []CandidatePool `json:"pools"`
}

// ComposeRequest contains filtered characters from qiming.pick.
type ComposeRequest struct {
	First    []string
	Second   []string
	MaxNames int
}

// ComposeResult is the generated given-name output.
type ComposeResult struct {
	TotalPossible int      `json:"total_possible"`
	Names         []string `json:"names"`
}

// EvaluationError identifies why a candidate name is invalid.
type EvaluationError struct {
	Code string `json:"code"`
	Char string `json:"char,omitempty"`
}

// WuxingHit reports matches against naming five-element constraints.
type WuxingHit struct {
	Yong *bool `json:"yong,omitempty"`
	Xi   *bool `json:"xi,omitempty"`
	Ji   *bool `json:"ji,omitempty"`
}

// Evaluation is the result of evaluating one candidate name.
type Evaluation struct {
	GivenName  string            `json:"given_name"`
	Valid      bool              `json:"valid"`
	Errors     []EvaluationError `json:"errors,omitempty"`
	Characters []Character       `json:"characters,omitempty"`
	Phonetic   *Phonetic         `json:"phonetic,omitempty"`
	Wuxing     *WuxingHit        `json:"wuxing,omitempty"`
}
