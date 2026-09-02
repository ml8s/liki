package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/qiming"
)

func qimingPickHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Wuxing1 string `json:"wuxing1"`
		Wuxing2 string `json:"wuxing2"`
		Count   int    `json:"count"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("qiming.pick: %w", err)
	}
	if params.Count == 0 {
		params.Count = 2
	}
	result, err := qiming.PickChars(params.Wuxing1, params.Wuxing2, params.Count)
	if err != nil {
		return nil, fmt.Errorf("qiming.pick: %w", err)
	}
	return wrapResult("qiming_pick", result)
}

func qimingComposeHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var params struct {
		First    []string `json:"first"`
		Second   []string `json:"second"`
		MaxNames int      `json:"max_names"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("qiming.compose: %w", err)
	}
	result, err := qiming.ComposeNames(qiming.ComposeRequest{
		First:    params.First,
		Second:   params.Second,
		MaxNames: params.MaxNames,
	})
	if err != nil {
		return nil, fmt.Errorf("qiming.compose: %w", err)
	}
	return wrapResult("qiming_compose", result)
}

func qimingCharHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Char string `json:"char"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("qiming.char: %w", err)
	}
	if len([]rune(params.Char)) != 1 {
		return nil, fmt.Errorf("qiming.char: must be a single character")
	}
	result := qiming.LookupChar(params.Char)
	if result == nil {
		return nil, fmt.Errorf("qiming.char: character %q not found in database", params.Char)
	}
	return wrapResult("qiming_char", result)
}

func qimingCheckHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var params struct {
		GivenNames []string `json:"given_names"`
		YongShen   string   `json:"yongshen"`
		XiShen     []string `json:"xishen"`
		JiShen     []string `json:"jishen"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("qiming.check: %w", err)
	}
	results, err := qiming.EvaluateNames(params.GivenNames, params.YongShen, params.XiShen, params.JiShen)
	if err != nil {
		return nil, fmt.Errorf("qiming.check: %w", err)
	}
	return wrapResult("qiming_check", results)
}

const qimingWuxingEnum = `{"type":"string","enum":["木","火","土","金","水"]}`

const qimingCharacterSchema = `{
	"type":"object",
	"additionalProperties":false,
	"properties":{
		"char":{"type":"string","minLength":1},
		"wuxing":{"type":"string","enum":["金","木","水","火","土"]},
		"stroke":{"type":"integer","minimum":1,"description":"现代规范汉字笔画"},
		"kangxi_stroke":{"type":"integer","minimum":1,"description":"Kangxi 形体对应的笔画"},
		"radical":{"type":"string","minLength":1},
		"pinyin":{"type":"string","minLength":1},
		"tone":{"type":"integer","minimum":1,"maximum":5},
		"kangxi_form":{"type":"string","minLength":1,"description":"笔画对应的 Kangxi 形体"}
	},
	"required":["char","wuxing","stroke","kangxi_stroke","pinyin","tone","kangxi_form"]
}`

var qimingMethods = []RPCMethod{
	{
		Name:        "qiming.char",
		Description: "查字。查询单个汉字的五行、现代笔画、Kangxi 笔画、部首、拼音等信息。",
		Params: mustSchema(`{
			"type":"object",
			"additionalProperties":false,
			"properties":{"char":{"type":"string","minLength":1,"maxLength":1,"description":"要查询的单个汉字"}},
			"required":["char"]
		}`),
		Handler: qimingCharHandler,
		Result:  envelopeSchema(qimingCharacterSchema),
	},
	{
		Name:        "qiming.pick",
		Description: "起名取字。按五行返回候选字池；count=1 返回第一字池，count=2 返回第一、第二字池。",
		Params: mustSchema(`{
			"type":"object",
			"additionalProperties":false,
			"properties":{
				"wuxing1":` + qimingWuxingEnum + `,
				"wuxing2":{"type":"string","enum":["木","火","土","金","水"],"description":"第2字五行（count=2 时使用；默认同 wuxing1，响应在相同时省略）"},
				"count":{"type":"integer","enum":[1,2],"description":"1=单名/2=双名，默认2"}
			},
			"required":["wuxing1"]
		}`),
		Handler: qimingPickHandler,
		Result: envelopeSchema(`{
			"type":"object",
			"additionalProperties":false,
			"properties":{
				"wuxing1":` + qimingWuxingEnum + `,
				"wuxing2":` + qimingWuxingEnum + `,
				"pools":{
					"type":"array","minItems":1,"maxItems":2,
					"items":{
						"type":"object",
						"additionalProperties":false,
						"properties":{
							"slot":{"type":"string","enum":["first","second"]},
							"chars":{"type":"array","minItems":1,"uniqueItems":true,"items":{"type":"string","minLength":1,"maxLength":1}}
						},
						"required":["slot","chars"]
					}
				}
			},
			"required":["wuxing1","pools"]
		}`),
	},
	{
		Name:        "qiming.compose",
		Description: "起名组名。first 只传单字数组；传 second 生成双字名，不传 second 生成单字名；字符事实由 qiming.check 返回。",
		Params: mustSchema(`{
			"type":"object",
			"additionalProperties":false,
			"properties":{
				"first":{"type":"array","minItems":1,"maxItems":256,"uniqueItems":true,"items":{"type":"string","minLength":1,"maxLength":1}},
				"second":{"type":"array","minItems":1,"maxItems":256,"uniqueItems":true,"items":{"type":"string","minLength":1,"maxLength":1}},
				"max_names":{"type":"integer","minimum":1,"maximum":1000,"description":"返回名数量上限，默认100"}
			},
			"required":["first"]
		}`),
		Handler: qimingComposeHandler,
		Result: envelopeSchema(`{
			"type":"object",
			"additionalProperties":false,
			"properties":{
				"total_possible":{"type":"integer","minimum":1},
				"names":{
					"type":"array","maxItems":1000,
					"items":{"type":"string","minLength":1,"maxLength":2}
				}
			},
			"required":["total_possible","names"]
		}`),
	},
	{
		Name:        "qiming.check",
		Description: "起名评估。given_names 传候选名（不含姓），服务端校验字库并返回五行匹配与音韵信息。",
		Params: mustSchema(`{
			"type":"object",
			"additionalProperties":false,
			"properties":{
				"given_names":{"type":"array","minItems":1,"maxItems":50,"uniqueItems":true,"items":{"type":"string","minLength":1,"maxLength":2},"description":"候选名字列表（不含姓）"},
				"yongshen":{"type":"string","enum":["木","火","土","金","水"]},
				"xishen":{"type":"array","maxItems":5,"uniqueItems":true,"items":{"type":"string","enum":["木","火","土","金","水"]}},
				"jishen":{"type":"array","maxItems":5,"uniqueItems":true,"items":{"type":"string","enum":["木","火","土","金","水"]}}
			},
			"required":["given_names"]
		}`),
		Handler: qimingCheckHandler,
		Result: envelopeSchema(`{
			"type":"array",
			"items":{
				"type":"object",
				"additionalProperties":false,
				"properties":{
					"given_name":{"type":"string","minLength":1,"maxLength":2},
					"valid":{"type":"boolean","description":"名字通过字库与负面字校验"},
					"errors":{
						"type":"array","maxItems":2,
						"items":{
							"type":"object",
							"additionalProperties":false,
							"properties":{
								"code":{"type":"string","enum":["invalid_name_length","character_not_found","negative_character_forbidden"]},
								"char":{"type":"string","minLength":1}
							},
							"required":["code"]
						}
					},
					"characters":{"type":"array","minItems":1,"maxItems":2,"items":` + qimingCharacterSchema + `},
					"phonetic":{
						"type":"object",
						"additionalProperties":false,
						"properties":{"tones":{"type":"string","minLength":1}},
						"required":["tones"]
					},
					"wuxing":{
						"type":"object",
						"additionalProperties":false,
						"properties":{
							"yong":{"type":"boolean","description":"任一字匹配用神五行；仅提供 yongshen 时返回"},
							"xi":{"type":"boolean","description":"任一字匹配喜神五行；仅提供 xishen 时返回"},
							"ji":{"type":"boolean","description":"任一字匹配忌神五行；仅提供 jishen 时返回"}
						}
					}
				},
				"required":["given_name","valid"]
			}
			}`),
	},
}
