package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/qiming"
)

func qimingPickHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Wuxing string        `json:"wuxing"`
		Pairs  []qiming.StrokePair  `json:"pairs,omitempty"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qiming.pick: %w", err)
	}

	chars, err := qiming.GetChars(p.Wuxing)
	if err != nil {
		return nil, fmt.Errorf("qiming.pick: %w", err)
	}

	// If pairs are provided, filter chars to only include strokes matching the pairs
	if len(p.Pairs) > 0 {
		validStrokes := make(map[int]bool)
		for _, pair := range p.Pairs {
			validStrokes[pair.S1] = true
			validStrokes[pair.S2] = true
		}
		filtered := make(map[int][]qiming.CharLite)
		for stroke, charList := range chars {
			if validStrokes[stroke] {
				filtered[stroke] = charList
			}
		}
		chars = filtered
	}

	return wrapResult("naming_pick", map[string]any{
		"chars": chars,
	})
}

func qimingBuildHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname string                    `json:"surname"`
		Chars1  map[int][]qiming.CharLite `json:"chars1"`
		Chars2  map[int][]qiming.CharLite `json:"chars2"`
		Pairs   []qiming.StrokePair       `json:"pairs"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qiming.build: %w", err)
	}

	names := qiming.ComposeNames(p.Surname, p.Chars1, p.Chars2, p.Pairs)
	return wrapResult("naming_build", map[string]any{"names": names})
}


func qimingCharHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Char string `json:"char"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qiming.char: %w", err)
	}
	if len([]rune(p.Char)) != 1 {
		return nil, fmt.Errorf("qiming.char: must be a single character")
	}
	result := qiming.LookupChar(p.Char)
	if result == nil {
		return nil, fmt.Errorf("qiming.char: character %q not found in database", p.Char)
	}
	return wrapResult("qiming_char", result)
}
func qimingWugeHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname string `json:"surname"`
		Count   int    `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qiming.wuge: %w", err)
	}
	if p.Count != 1 && p.Count != 2 {
		p.Count = 2
	}

	stroke, err := qiming.SurnameStroke(p.Surname)
	if err != nil {
		return nil, fmt.Errorf("qiming.wuge: %w", err)
	}

	pairs := qiming.ListViableStrokes(stroke, p.Count)
	pairs = qiming.FilterSancai(stroke, pairs)
	return wrapResult("naming_wuge", map[string]any{
		"surname_stroke": stroke,
		"pairs":          pairs,
	})
}

func qimingCheckHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname  string   `json:"surname"`
		Names    []string `json:"names"`
		YongShen string   `json:"yongshen"`
		XiShen   []string `json:"xishen"`
		JiShen   []string `json:"jishen"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qiming.check: %w", err)
	}

	results, err := qiming.EvaluateNames(p.Surname, p.Names, p.YongShen, p.XiShen, p.JiShen)
	if err != nil {
		return nil, fmt.Errorf("qiming.check: %w", err)
	}
	return wrapResult("naming_check", results)
}

var qimingMethods = []RPCMethod{
	{
		Name: "qiming.pick", Description: "按五行取可用起名字。返回按笔画分组的字库（供 qiming.build 组名）。可选传 pairs 按笔画过滤（pairs 来自 qiming.wuge）。用+喜需调两次（分别取用神/喜神五行）。输出对象直接作为 qiming.build 的 chars1/chars2 传入。",
		Params: mustSchema(`{"type":"object","properties":{"wuxing":{"type":"string","enum":["木","火","土","金","水"],"description":"单个五行"},"pairs":{"type":"array","items":{"type":"object","properties":{"s1":{"type":"integer"},"s2":{"type":"integer"}},"required":["s1","s2"]},"description":"可选笔画过滤：只返回这些笔画的字。pairs 从 qiming.wuge 返回。不传则返回该五行全部字"}},"required":["wuxing"]}`),
		Handler: qimingPickHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"chars":{"type":"object","description":"字库，按笔画分组"}},"required":["chars"]}`),
	},
	{
		Name: "qiming.build", Description: "起名组名。chars1/chars2 传入 qiming.pick 返回的字库对象（可先语义过滤），pairs 传入 qiming.wuge 返回的笔画对（可选，不传则 chars1×chars2 全组合）。双名必须传 chars2，单名不传。返回候选名列表，LLM 需再语义过滤后用 qiming.check 评估。",
		Params: mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏","examples":["李"]},"chars1":{"type":"object","description":"qiming.pick 返回的按笔画分组的字库对象。格式：{ stroke_number: [{char: '某', tone: 1}, ...] }。不要传数组或扁平列表。","examples":[{"8":[{"char":"林","tone":2}],"10":[{"char":"桂","tone":4}],"14":[{"char":"熊","tone":2}]}]},"chars2":{"type":"object","description":"字2字库，单名不传。格式与 chars1 相同（从 pick 返回的 chars）。","examples":[{"12":[{"char":"舜","tone":4}],"16":[{"char":"澄","tone":2}],"20":[{"char":"曦","tone":1}]}]},"pairs":{"type":"array","items":{"type":"object","properties":{"s1":{"type":"integer"},"s2":{"type":"integer"}},"required":["s1","s2"]},"description":"可选笔画约束对，从 wuge 透传。不传则全量笛卡尔积","examples":[[{"s1":1,"s2":5},{"s1":1,"s2":7},{"s1":1,"s2":10}]]}},"required":["surname","chars1"]}`),
		Handler: qimingBuildHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"names":{"type":"array","items":{"type":"string"},"description":"候选名列表"}},"required":["names"]}`),
	},
	{
		Name: "qiming.wuge", Description: "查五格可行笔画对。surname 传姓氏（勿传笔画数），count 传字数（1=单名/2=双名）。返回人/地/外/总四格全吉的笔画组合 pairs，供 qiming.pick 按笔画取字、qiming.build 组名。",
		Params: mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏"},"count":{"type":"integer","enum":[1,2],"description":"字数：1=单名，2=双名，默认2"}},"required":["surname"]}`),
		Handler: qimingWugeHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"surname_stroke":{"type":"integer","description":"姓氏康熙笔画"},"pairs":{"type":"array","items":{"type":"object","properties":{"s1":{"type":"integer"},"s2":{"type":"integer"}},"required":["s1","s2"]},"description":"可行笔画对"}},"required":["surname_stroke","pairs"]}`),
	},
	{
		Name: "qiming.check", Description: "起名评估。names 传候选名（仅名字部分，不含姓），surname 传姓氏，yongshen 传用神五行。返回五格三才五行音韵全量判定，LLM 综合挑最终推荐名。",
		Params: mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏"},"names":{"type":"array","items":{"type":"string"},"description":"候选名字列表（仅名字部分，不含姓）"},"yongshen":{"type":"string","enum":["木","火","土","金","水"],"description":"用神五行（可选）"},"xishen":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]},"description":"喜神五行（可选）"},"jishen":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]},"description":"忌神五行（可选）"}},"required":["surname","names"]}`),
		Handler: qimingCheckHandler,
		Result: envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"surname":{"type":"string"},"given_name":{"type":"string"},"characters":{"type":"array"},"wuge":{"type":"object","description":"五格配置","properties":{"tiange":{"type":"object","description":"天格"},"renge":{"type":"object","description":"人格"},"dige":{"type":"object","description":"地格"},"waige":{"type":"object","description":"外格"},"zongge":{"type":"object","description":"总格"}}},"sancai":{"type":"object","description":"三才五行","properties":{"tian_cai":{"type":"string","description":"天才五行","enum":["金","木","水","火","土"]},"ren_cai":{"type":"string","description":"人才五行","enum":["金","木","水","火","土"]},"di_cai":{"type":"string","description":"地才五行","enum":["金","木","水","火","土"]}}},"phonetic":{"type":"object","properties":{"tones":{"type":"string"}},"required":["tones"]},"wuxing_match":{"type":"boolean"},"wuxing":{"type":"object","properties":{"yong":{"type":"boolean"},"xi":{"type":"boolean"},"ji":{"type":"boolean"}},"required":["yong"]}},"required":["name","surname","given_name","wuge","sancai","phonetic","wuxing_match"]}}`),
	},
	{
		Name: "qiming.char", Description: "查字。查询单个汉字的五行、笔画、部首、拼音等信息。",
		Params: mustSchema(`{"type":"object","properties":{"char":{"type":"string","minLength":1,"maxLength":1,"description":"要查询的单个汉字"}},"required":["char"]}`),
		Handler: qimingCharHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"char":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土"]},"stroke":{"type":"integer"},"radical":{"type":"string"},"pinyin":{"type":"string"},"tone":{"type":"integer"}},"required":["char","wuxing","stroke","pinyin"]}`),
	},
}
