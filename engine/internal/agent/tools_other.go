package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"liki-engine/internal/engine/liuyao"
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

// ── liuyao ──

// schemaLiuyaoCasting is the complete receipt contract for chart input. It is
// retained only until liuyao.chart migrates to the typed MCP registry.
const schemaLiuyaoCasting = `{"type":"object","additionalProperties":false,"properties":{"schema_version":{"type":"string","const":"liuyao-cast-v1"},"mode":{"type":"string","enum":["coins","yaos"]},"order":{"type":"string","const":"bottom_up"},"coin_convention":{"type":"object","additionalProperties":false,"properties":{"正":{"type":"integer","const":3},"反":{"type":"integer","const":2}},"required":["正","反"]},"rounds":{"type":"array","minItems":6,"maxItems":6,"items":{"type":"object","additionalProperties":false,"properties":{"position":{"type":"integer","minimum":1,"maximum":6},"coins":{"type":"array","minItems":3,"maxItems":3,"items":{"type":"string","enum":["正","反"]}},"value":{"type":"integer","minimum":6,"maximum":9},"label":{"type":"string","enum":["老阴","少阳","少阴","老阳"]},"changing":{"type":"boolean"}},"required":["position","coins","value","label","changing"]}},"yaos":{"type":"array","minItems":6,"maxItems":6,"items":{"type":"integer","minimum":6,"maximum":9}},"dong_yao":{"type":"array","items":{"type":"integer","minimum":1,"maximum":6}},"casting_id":{"type":"string","pattern":"^[a-f0-9]{64}$"}},"required":["schema_version","mode","order","yaos","dong_yao","casting_id"]}`

func liuyaoChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime string          `json:"solar_time"`
		YongShen  string          `json:"yong_shen"`
		Casting   *liuyao.Casting `json:"casting"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.chart: %w", err)
	}
	if p.Casting == nil {
		return nil, fmt.Errorf("liuyao.chart: casting is required")
	}
	if err := p.Casting.Validate(); err != nil {
		return nil, fmt.Errorf("liuyao.chart: invalid casting: %w", err)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("liuyao.chart: %w", err)
	}
	if p.YongShen == "" {
		p.YongShen = "世爻"
	}
	ys, err := liuyao.ParseYongShen(p.YongShen)
	if err != nil {
		return nil, fmt.Errorf("liuyao.chart: %w", err)
	}
	result := liuyao.ComputeChart(st, ys, p.Casting.Yaos)
	result.Casting = p.Casting
	return wrapResult("liuyao", result)
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

	{
		Name: "liuyao.chart", Description: "六爻装卦。只接受 liuyao.qigua 返回的完整 casting 收据；装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。lines.liu_qin: 0=父母 1=兄弟 2=官鬼 3=妻财 4=子孙；lines.liu_shou: 0=青龙 1=朱雀 2=勾陈 3=螣蛇 4=白虎 5=玄武；wang_shuai: 0=旺 1=相 2=休 3=囚 4=死",
		Params:  mustSchema(`{"type":"object","additionalProperties":false,"properties":{"solar_time":` + schemaSolarTime + `,"yong_shen":{"type":"string","enum":["父母","兄弟","官鬼","妻财","子孙","世爻","应爻"],"description":"用神；默认世爻"},"casting":` + schemaLiuyaoCasting + `},"required":["solar_time","casting"]}`),
		Handler: liuyaoChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"ben_gua":{"type":"string","enum":["乾","姤","遁","否","观","剥","晋","大有","兑","困","萃","咸","蹇","谦","小过","归妹","离","旅","鼎","未济","蒙","涣","讼","同人","震","豫","解","恒","升","井","大过","随","巽","小畜","家人","益","无妄","噬嗑","颐","蛊","坎","节","屯","既济","革","丰","明夷","师","艮","贲","大畜","损","睽","履","中孚","渐","坤","复","临","泰","大壮","夬","需","比"]},"lines":{"type":"array","description":"每爻：六亲/六神/世应 + 确定性状态（yue_po 月破/dong_self 发动/dong_sheng 动爻生/dong_ke 动爻克）","items":{"type":"object","properties":{"position":{"type":"integer"},"type":{"type":"integer"},"gan":{"type":"string"},"zhi":{"type":"string"},"wuxing":{"type":"string"},"liu_qin":{"type":"string","enum":["父母","兄弟","官鬼","妻财","子孙"]},"shi_ying":{"type":"string","description":"世/应"},"liu_shou":{"type":"string","enum":["青龙","朱雀","勾陈","螣蛇","白虎","玄武"]},"yue_po":{"type":"boolean","description":"月破"},"dong_self":{"type":"boolean","description":"本爻发动"},"dong_sheng":{"type":"boolean","description":"有动爻生此爻"},"dong_ke":{"type":"boolean","description":"有动爻克此爻"},"xun_kong":{"type":"boolean","description":"该爻地支值日柱旬空"},"mu_ku_branch":{"type":"string"},"mu_ku_element":{"type":"string"},"mu_ku_types":{"type":"array","items":{"type":"string","enum":["day","moving","transformed"]}},"chang_sheng_yue":{"type":"string"}},"required":["position","type","gan","zhi","wuxing","liu_qin","shi_ying","liu_shou"]}},"yong_shen":{"type":"object","description":"用神结果","properties":{"name":{"type":"string","description":"用神六亲"},"position":{"type":"integer","description":"爻位1-6（0=本卦不现）"},"is_hidden":{"type":"boolean","description":"本卦不现，取本宫伏神"},"wang_shuai":{"type":"string","description":"用神旺衰"},"yue_po":{"type":"boolean","description":"用神月破"},"xun_kong":{"type":"boolean","description":"用神旬空"},"mu_ku":{"type":"boolean","description":"用神入墓"},"liu_shou":{"type":"string","description":"用神临的六神","enum":["青龙","朱雀","勾陈","螣蛇","白虎","玄武"]},"fu_shen":{"type":"object","description":"飞伏（用神不现时）","properties":{"position":{"type":"integer","description":"爻位"},"liu_qin":{"type":"string","description":"伏神六亲"},"zhi":{"type":"string","description":"伏神地支"}},"required":["position","liu_qin","zhi"]},"fu_shen_candidates":{"type":"array","description":"本宫全部同六亲伏神候选","items":{"type":"object","properties":{"position":{"type":"integer"},"liu_qin":{"type":"string"},"zhi":{"type":"string"}},"required":["position","liu_qin","zhi"]}},"chang_sheng":{"type":"string"}},"required":["name","position"]},"wang_shuai":{"type":"array"},"yue_jian_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"yue_jian_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"bian_gua":{"type":"string","description":"变卦名"},"bian_yao":{"type":"array","description":"变爻"},"gong":{"type":"string","description":"八宫"},"gua_ci":{"type":"object","description":"卦辞爻辞"},"gong_wuxing":{"type":"string","description":"宫五行"},"ri_chen_gan":{"type":"string","description":"日辰天干"},"ri_chen_zhi":{"type":"string","description":"日辰地支"},"ri_chen_relations":{"type":"array","description":"日辰与爻关系集合；冲合可与生扶克并存","items":{"type":"object","additionalProperties":false,"properties":{"relations":{"type":"array","items":{"type":"string","enum":["冲","合","生","扶","克","平"]}}},"required":["relations"]}},"xun_kong":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"description":"日柱旬空地支（甲子旬空戌亥…）"},"dong_yao_relations":{"type":"array","description":"动爻与用神的关系集合；直接与间接作用可并列，静卦为空数组","items":{"type":"object","properties":{"position":{"type":"integer","description":"动爻位置"},"relations":{"type":"array","items":{"type":"string","enum":["生用","克用","比和","冲用","生原神","克原神","生忌神","克忌神"]}}},"required":["position","relations"]}},"patterns":{"type":"array","description":"特殊格局","items":{"type":"object","properties":{"type":{"type":"string","description":"格局类型"},"sub_type":{"type":"string","description":"子类型"},"position":{"type":"integer"},"is_true":{"type":"boolean","description":"结构是否有实质效力；空亡/月破中 true=真空/真破，false=假空/假破"}},"required":["type"]}},"casting":{"type":"object","description":"完整起卦收据"},"timing_candidates":{"type":"array","description":"应期触发机制候选；非确定日期","items":{"type":"object","properties":{"id":{"type":"string"},"mechanism":{"type":"string"},"position":{"type":"integer"},"trigger_branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"window":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}},"confidence":{"type":"string"}},"required":["id","mechanism","confidence"]}},"day_clash_facts":{"type":"array","description":"日冲细分事实","items":{"type":"object","properties":{"position":{"type":"integer"},"kind":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}},"condition":{"type":"string"}},"required":["position","kind"]}},"moving_transformations":{"type":"array","description":"动爻到变爻的机械变化","items":{"type":"object","properties":{"position":{"type":"integer"},"from_branch":{"type":"string"},"to_branch":{"type":"string"},"return_relation":{"type":"string"},"advance_retreat":{"type":"string"},"changed_void":{"type":"boolean"},"changed_break":{"type":"boolean"},"basis":{"type":"array","items":{"type":"string"}}},"required":["position","from_branch","to_branch"]}},"force_chain":{"type":"object","properties":{"yong_shen":{"type":"string"},"yong_element":{"type":"string"},"position":{"type":"integer"},"is_hidden":{"type":"boolean"},"entries":{"type":"array","items":{"type":"object","properties":{"role":{"type":"string"},"element":{"type":"string"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"relations":{"type":"array","items":{"type":"string"}},"states":{"type":"array","items":{"type":"string"}},"note":{"type":"string"}},"required":["role","element"]}}},"required":["yong_shen","yong_element","position","entries"]},"yong_shen_candidates":{"type":"array","description":"用神候选与当前确定性选择","items":{"type":"object","properties":{"position":{"type":"integer"},"branch":{"type":"string"},"is_hidden":{"type":"boolean"},"selected":{"type":"boolean"},"wang_shuai":{"type":"string"},"reason":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}}},"required":["position","branch","is_hidden","selected","reason"]}},"san_he_candidates":{"type":"array","properties":{"moving_count":{"type":"integer"},"static_count":{"type":"integer"},"void_count":{"type":"integer"},"break_count":{"type":"integer"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"element":{"type":"string"},"complete":{"type":"boolean"},"qualification":{"type":"string"},"line_complete":{"type":"boolean"},"day_present":{"type":"boolean"},"month_present":{"type":"boolean"},"missing":{"type":"array","items":{"type":"string"}},"activated":{"type":"boolean"},"targets":{"type":"array","items":{"type":"string"}},"conclusion_scope":{"type":"string"}},"required":["positions","branches","element","complete","moving_count","qualification","line_complete","activated","conclusion_scope"],"description":"三合候选；complete=true 只是三支齐，成局仍须按发动与空破月日复核","items":{"type":"object","properties":{"moving_count":{"type":"integer"},"static_count":{"type":"integer"},"void_count":{"type":"integer"},"break_count":{"type":"integer"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"element":{"type":"string"},"complete":{"type":"boolean"},"qualification":{"type":"string"},"line_complete":{"type":"boolean"},"day_present":{"type":"boolean"},"month_present":{"type":"boolean"},"missing":{"type":"array","items":{"type":"string"}},"activated":{"type":"boolean"},"targets":{"type":"array","items":{"type":"string"}},"conclusion_scope":{"type":"string"}},"required":["positions","branches","element","complete","moving_count","qualification","line_complete","activated","conclusion_scope"]}},"tomb_school":{"type":"object","additionalProperties":false,"properties":{"earth_branch":{"type":"string","enum":["辰"]},"school":{"type":"string","enum":["engine_default_chen"]}},"required":["earth_branch","school"]},"hidden_lines":{"type":"array","description":"本宫全部伏神","items":{"type":"object","properties":{"position":{"type":"integer"},"liu_qin":{"type":"string"},"gan":{"type":"string"},"zhi":{"type":"string"},"wuxing":{"type":"string"},"flying_position":{"type":"integer"},"flying_liu_qin":{"type":"string"},"flying_gan":{"type":"string"},"flying_zhi":{"type":"string"},"flying_wuxing":{"type":"string"},"fei_fu_relation":{"type":"string"},"xun_kong":{"type":"boolean"},"yue_po":{"type":"boolean"},"mu_ku":{"type":"boolean"},"exit_condition":{"type":"string"}},"required":["position","liu_qin","gan","zhi","wuxing","flying_position","flying_liu_qin","fei_fu_relation"]}},"branch_relation_facts":{"type":"array","description":"爻间地支关系","items":{"type":"object","properties":{"left_position":{"type":"integer"},"right_position":{"type":"integer"},"relations":{"type":"array","items":{"type":"string"}},"targets":{"type":"array","items":{"type":"string"}},"basis":{"type":"array","items":{"type":"string"}}},"required":["left_position","right_position","relations"]}},"conflicts":{"type":"array","description":"engine 已判定的并列/矛盾信号；解释必须覆盖","items":{"type":"object","properties":{"id":{"type":"string","enum":["moving-support-opposition","yuanshen-support-opposition","strong-but-void","strong-but-break"]},"reason":{"type":"string"},"facts":{"type":"array","items":{"type":"string"}}},"required":["id","reason"]}}},"required":["name","ben_gua","lines","yong_shen","tomb_school","casting"]}`),
	},
}
