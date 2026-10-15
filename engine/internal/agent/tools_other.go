package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"liki-engine/internal/engine/qimen"
)

// ── qimen ──

func qimenChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime        string   `json:"solar_time"`
		Scope            string   `json:"scope"`
		School           string   `json:"school"`
		DingjuMethod     string   `json:"dingju_method"`
		QuarterRule      string   `json:"quarter_rule"`
		BaseDingjuMethod string   `json:"base_dingju_method"`
		DunSource        string   `json:"dun_source"`
		HourBoundary     string   `json:"hour_boundary"`
		YongShen         []string `json:"yong_shen"`
		BirthDate        string   `json:"birth_date"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	method, err := qimen.ResolveMethodWithOptions(
		p.Scope, p.School, p.DingjuMethod, p.QuarterRule,
		p.BaseDingjuMethod, p.DunSource, p.HourBoundary,
	)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	if method.School == qimen.SchoolJinhanYuJing {
		if len(p.YongShen) > 0 || p.BirthDate != "" {
			return nil, fmt.Errorf("qimen.chart: jinhan_yujing does not support yong_shen or birth_date")
		}
		return wrapResult("qimen", qimen.ComputeJinhanChart(st))
	}
	var birthDate qimen.BirthDate
	if p.BirthDate != "" {
		birthDate, err = qimen.ParseBirthDate(p.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("qimen.chart: %w", err)
		}
	}
	if len(p.YongShen) > 0 || birthDate.Has {
		syms := make([]qimen.YongShenSymbol, 0, len(p.YongShen))
		for _, name := range p.YongShen {
			sym, err := qimen.ParseYongShen(name)
			if err != nil {
				return nil, fmt.Errorf("qimen.chart: %w", err)
			}
			syms = append(syms, sym)
		}
		result, err := qimen.ComputeChartWithYongShenAndMethod(st, syms, birthDate, method)
		if err != nil {
			return nil, fmt.Errorf("qimen.chart: %w", err)
		}
		return wrapResult("qimen", result)
	}
	result, err := qimen.ComputeChartWithMethod(st, method)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	return wrapResult("qimen", result)
}

// ── time / infra ──

//go:embed qimen_schema.json
var qimenSchemaJSON []byte

var qimenSchemas = mustLoadQimenSchemas()

type qimenSchemaDocument struct {
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
}

func mustLoadQimenSchemas() qimenSchemaDocument {
	var doc qimenSchemaDocument
	if err := json.Unmarshal(qimenSchemaJSON, &doc); err != nil {
		panic("invalid qimen schema document: " + err.Error())
	}
	if len(doc.Params) == 0 || len(doc.Result) == 0 {
		panic("qimen schema document lacks params or result")
	}
	return doc
}

var otherMethods = []RPCMethod{
	{
		Name: "qimen.chart", Description: "奇门排盘：时/刻/日/月/年家 × 转盘/飞盘/鸣法；时家与日家支持拆补/置闰；刻家包含十分钟三元与十二分钟十分局；金函玉镜为 day 专用盘。yong_shen 可显式定位用神，事象路由由上层 skill 完成；ying_qi 返回日期窗口。solar_time 必须来自 tianwen.time 的真太阳时。",
		Params:  qimenSchemas.Params,
		Handler: qimenChartHandler,
		Result:  envelopeSchema(string(qimenSchemas.Result)),
	},
}
