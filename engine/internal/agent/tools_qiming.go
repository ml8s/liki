package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/qiming"
)

func qimingPickHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname string `json:"surname"`
		Wuxing1 string `json:"wuxing1"`
		Wuxing2 string `json:"wuxing2"`
		Count   int    `json:"count"`
		Wuge    bool   `json:"wuge"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qiming.pick: %w", err)
	}
	if p.Count != 1 && p.Count != 2 {
		p.Count = 2
	}
	combos, err := qiming.PickChars(p.Surname, p.Wuxing1, p.Wuxing2, p.Count, p.Wuge)
	if err != nil {
		return nil, fmt.Errorf("qiming.pick: %w", err)
	}
	return wrapResult("naming_pick", map[string]any{"combos": combos})
}

func qimingBuildHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Combos []qiming.Combo `json:"combos"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qiming.build: %w", err)
	}

	names := qiming.BuildNames(p.Combos)
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

func qimingCheckHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname  string   `json:"surname"`
		Names    []string `json:"names"`
		YongShen string   `json:"yongshen"`
		XiShen   []string `json:"xishen"`
		JiShen   []string `json:"jishen"`
		Wuge     *bool    `json:"wuge"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qiming.check: %w", err)
	}
	wuge := true
	if p.Wuge != nil {
		wuge = *p.Wuge
	}

	results, err := qiming.EvaluateNames(p.Surname, p.Names, p.YongShen, p.XiShen, p.JiShen, wuge)
	if err != nil {
		return nil, fmt.Errorf("qiming.check: %w", err)
	}
	return wrapResult("naming_check", results)
}

var qimingMethods = []RPCMethod{
	{
		Name: "qiming.pick", Description: "起名取字（合并五格计算）。wuge=true 时算吉笔画并按其取字（surname 必填）；wuge=false 时按笔画取字（surname 不需要）。count 恒生效：1=单名（每组合仅 first 第1字池）/2=双名（每组合 first 第1字池 + second 第2字池）。wuxing2 默认同 wuxing1（用+喜时传不同五行）。输出 combos 数组，每组合 {id, first, second}，first/second 为纯字数组。LLM 筛选后传给 qiming.build。",
		Params:  mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏（wuge=true 算五格需要）"},"wuxing1":{"type":"string","enum":["木","火","土","金","水"],"description":"第1字五行"},"wuxing2":{"type":"string","enum":["木","火","土","金","水"],"description":"第2字五行（双名用+喜；默认同 wuxing1）"},"count":{"type":"integer","enum":[1,2],"description":"1=单名/2=双名，默认2"},"wuge":{"type":"boolean","description":"true=考虑三才五格按吉笔画取字（默认）/false=不考虑按笔画取字"}},"required":["wuxing1"]}`),
		Handler: qimingPickHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"combos":{"type":"array","description":"取字组合","items":{"type":"object","properties":{"id":{"type":"integer","description":"组合key"},"first":{"type":"array","items":{"type":"string"},"description":"第1字池（纯字）"},"second":{"type":"array","items":{"type":"string"},"description":"第2字池（双名）"}},"required":["id","first"]}}},"required":["combos"]}`),
	},
	{
		Name: "qiming.build", Description: "起名组名（仅双名）。combos 传入 qiming.pick 输出、经 LLM 筛选后的组合（每组 first/second 是挑好的字）。对每组 first×second 笛卡尔积组合出 given name（不含姓）。单名不走 build（LLM 直接从 pick 选字）。返回 given name 数组，LLM 再筛后传 qiming.check。",
		Params:  mustSchema(`{"type":"object","properties":{"combos":{"type":"array","description":"pick 输出的组合（LLM 筛选后）","items":{"type":"object","properties":{"id":{"type":"integer"},"first":{"type":"array","items":{"type":"string"},"description":"第1字挑好的字"},"second":{"type":"array","items":{"type":"string"},"description":"第2字挑好的字（双名）"}},"required":["id","first"]}}},"required":["combos"]}`),
		Handler: qimingBuildHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"names":{"type":"array","items":{"type":"string"},"description":"given name 数组（不含姓）"}},"required":["names"]}`),
	},
	{
		Name: "qiming.check", Description: "起名评估。names 传候选名（仅名字部分，不含姓），surname 传姓氏，yongshen 传用神五行。wuge=true 时返回五格三才全量判定（默认），false 时跳过五格三才、仅返回五行匹配/音韵/字义。LLM 综合挑最终推荐名。",
		Params:  mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏"},"names":{"type":"array","items":{"type":"string"},"description":"候选名字列表（仅名字部分，不含姓）"},"yongshen":{"type":"string","enum":["木","火","土","金","水"],"description":"用神五行（可选）"},"xishen":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]},"description":"喜神五行（可选）"},"jishen":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]},"description":"忌神五行（可选）"},"wuge":{"type":"boolean","description":"true=返回五格三才（默认）/false=跳过五格三才"}},"required":["surname","names"]}`),
		Handler: qimingCheckHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"surname":{"type":"string"},"given_name":{"type":"string"},"characters":{"type":"array"},"wuge":{"type":"object","description":"五格配置（wuge=false 时不返回）","properties":{"tiange":{"type":"object","description":"天格"},"renge":{"type":"object","description":"人格"},"dige":{"type":"object","description":"地格"},"waige":{"type":"object","description":"外格"},"zongge":{"type":"object","description":"总格"}}},"sancai":{"type":"object","description":"三才配置（wuge=false 时不返回）","properties":{"configuration":{"type":"string","description":"三才组合（如木火土）"},"ji_xiong":{"type":"string","description":"三才吉凶","enum":["大吉","吉","半吉","凶"]},"description":{"type":"string","description":"三才解释"}},"required":["configuration","ji_xiong"]},"phonetic":{"type":"object","properties":{"tones":{"type":"string"}},"required":["tones"]},"wuxing_match":{"type":"boolean"},"wuxing":{"type":"object","properties":{"yong":{"type":"boolean"},"xi":{"type":"boolean"},"ji":{"type":"boolean"}},"required":["yong"]}},"required":["name","surname","given_name","phonetic","wuxing_match"]}}`),
	},
	{
		Name: "qiming.char", Description: "查字。查询单个汉字的五行、笔画、部首、拼音等信息。",
		Params:  mustSchema(`{"type":"object","properties":{"char":{"type":"string","minLength":1,"maxLength":1,"description":"要查询的单个汉字"}},"required":["char"]}`),
		Handler: qimingCharHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"char":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土"]},"stroke":{"type":"integer"},"radical":{"type":"string"},"pinyin":{"type":"string"},"tone":{"type":"integer"},"traditional":{"type":"string","description":"繁体字"}},"required":["char","wuxing","stroke","pinyin"]}`),
	},
}
