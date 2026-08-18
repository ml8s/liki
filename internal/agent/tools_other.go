package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/agent/city"
	"liki-engine/internal/engine/bazhai"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/huangli"
	"liki-engine/internal/engine/liuyao"
	"liki-engine/internal/engine/qimen"
	"liki-engine/internal/engine/xuankong"
)

// ── qimen ──

func qimenChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime string   `json:"solar_time"`
		Kind      string   `json:"kind"`
		YongShen  []string `json:"yong_shen"`
		BirthYear int      `json:"birth_year"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	if p.Kind == "" {
		p.Kind = "shi"
	}
	kind, err := qimen.ParseChartKind(p.Kind)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	if len(p.YongShen) > 0 {
		// 带用神符号组合 → 聚合用神
		syms := make([]qimen.YongShenSymbol, 0, len(p.YongShen))
		for _, name := range p.YongShen {
			sym, err := qimen.ParseYongShen(name)
			if err != nil {
				return nil, fmt.Errorf("qimen.chart: %w", err)
			}
			syms = append(syms, sym)
		}
		result := qimen.ComputeChartWithYongShen(st, kind, syms, p.BirthYear)
		return wrapResult("qimen", result)
	}
	result := qimen.ComputeChart(st, kind)
	return wrapResult("qimen", result)
}



// ── bazhai ──

func bazhaiLayoutHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart     json.RawMessage `json:"chart"`
		DoorGua   string          `json:"door_gua"`
		MasterGua string          `json:"master_gua"`
		StoveGua  string          `json:"stove_gua"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("bazhai.layout: %w", err)
	}
	var chart bazhai.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("bazhai.layout: parse chart: %w", err)
	}
	result := bazhai.ComputeLayout(chart, p.DoorGua, p.MasterGua, p.StoveGua)
	return wrapResult("bazhai_layout", result)
}

func bazhaiChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime string         `json:"solar_time"`
		Gender    ganzhi.Gender `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("bazhai.chart: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("bazhai.chart: %w", err)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("bazhai.chart: %w", err)
	}
	result := bazhai.ComputeChart(st, p.Gender)
	return wrapResult("bazhai", result)
}

// ── xuankong ──

func xuankongChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime    string `json:"solar_time"`
		SitMountain  *int   `json:"zuo_shan"`
		FaceMountain *int   `json:"xiang_shan"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("xuankong.chart: %w", err)
	}
	if p.SitMountain == nil {
		return nil, fmt.Errorf("xuankong.chart: zuo_shan is required")
	}
	if p.FaceMountain == nil {
		return nil, fmt.Errorf("xuankong.chart: xiang_shan is required")
	}
	if *p.SitMountain < 0 || *p.SitMountain > 23 {
		return nil, fmt.Errorf("xuankong.chart: zuo_shan must be 0-23, got %d", *p.SitMountain)
	}
	if *p.FaceMountain < 0 || *p.FaceMountain > 23 {
		return nil, fmt.Errorf("xuankong.chart: xiang_shan must be 0-23, got %d", *p.FaceMountain)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("xuankong.chart: %w", err)
	}
	result := xuankong.ComputeChart(st, *p.SitMountain, *p.FaceMountain)
	return wrapResult("xuankong", result)
}

func xuankongLiunianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"` // 可选：宅盘
		Year  int             `json:"year"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("xuankong.liunian: %w", err)
	}
	if p.Year < 1864 || p.Year > 2200 {
		return nil, fmt.Errorf("xuankong.liunian: year %d out of range (1864-2200)", p.Year)
	}
	var chart *xuankong.Chart
	if len(p.Chart) > 0 && string(p.Chart) != "null" {
		var c xuankong.Chart
		if err := json.Unmarshal(p.Chart, &c); err != nil {
			return nil, fmt.Errorf("xuankong.liunian: parse chart: %w", err)
		}
		chart = &c
	}
	result := xuankong.ComputeLiuNian(p.Year, chart)
	return wrapResult("xuankong_liunian", result)
}

// ── liuyao ──

func liuyaoQiguaHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Seed *int64 `json:"seed,omitempty"` // 可选：固定随机种子（测试用）
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.qigua: %w", err)
	}
	var result liuyao.QiguaResult
	if p.Seed != nil {
		result = liuyao.QiguaWithSeed(*p.Seed)
	} else {
		result = liuyao.Qigua()
	}
	return wrapResult("liuyao_qigua", result)
}

func liuyaoChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime string `json:"solar_time"`
		YongShen  string `json:"yong_shen"`
		Yaos      [6]int `json:"yaos"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.chart: %w", err)
	}
	for i, v := range p.Yaos {
		if v < 6 || v > 9 {
			return nil, fmt.Errorf("liuyao.chart: yao[%d] = %d, must be 6-9", i, v)
		}
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
	result := liuyao.ComputeChart(st, ys, p.Yaos)
	return wrapResult("liuyao", result)
}

// ── huangli ──

func huangliDaysHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		StartDate string `json:"start_date"`
		Count     int    `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("huangli.days: %w", err)
	}
	start, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return nil, fmt.Errorf("huangli.days: parse date: %w", err)
	}
	if p.Count < 1 { p.Count = 3 }
	if p.Count > 30 { p.Count = 30 }
	days := make([]huangli.Day, 0, p.Count)
	for i := 0; i < p.Count; i++ {
		dateStr := start.AddDate(0, 0, i).Format("2006-01-02")
		entry, err := huangli.QueryDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("huangli.days: %w", err)
		}
		days = append(days, entry)
	}
	return wrapResult("huangli_days", days)
}

// ── time / infra ──

func cityCoordsHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	result, err := city.SearchCoords(ctx, raw)
	if err != nil {
		return nil, err
	}
	return wrapResult("city_coords", result)
}

func timeNowHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	now := time.Now()
	result := struct {
		UTC   string `json:"utc"`
		Local string `json:"local"`
		CST   string `json:"cst"`
	}{
		UTC:   now.UTC().Format(time.RFC3339),
		Local: now.Format(time.RFC3339),
		CST:   now.Format("2006-01-02T15:04:05+08:00"),
	}
	return wrapResult("time_now", result)
}

func schemaTimePointParams() json.RawMessage {
	return mustSchema(`{"type":"object","properties":{"time":{"type":"string","format":"date-time","description":"出生时间（RFC3339），如 1984-02-04T06:00:00+08:00"},"longitude":{"type":"number","description":"出生地经度，用于真太阳时校正。北京≈116.4"}},"required":["time","longitude"]}`)
}

var otherMethods = []RPCMethod{	{
		Name: "qimen.chart", Description: "奇门排盘。返回天盘、人盘、神盘、九星八门格局。kind 默认 shi（时家奇门），可选 ri/yue/nian。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"kind":{"type":"string","enum":["shi","ri","yue","nian"],"description":"奇门类型，默认 shi"},"yong_shen":{"type":"array","description":"用神符号组合（门/星/神/干），如 开门+天心，传入则聚合用神落宫","items":{"type":"string","description":"用神符号：门(休/生/伤/杜/景/死/惊/开)、星(天蓬/天芮/天冲/天辅/天禽/天心/天柱/天任/天英)、神(值符/螣蛇/太阴/六合/勾陈/朱雀/九地/九天)、干(甲/乙/丙/丁/戊/己/庚/辛/壬/癸)"}},"birth_year":{"type":"integer","description":"出生年份（年命干落宫）"}},"required":["solar_time"]}`),
		Handler: qimenChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"pan":{"type":"object","description":"奇门排盘结果","properties":{"jushu":{"type":"integer","description":"局数"},"yin_dun":{"type":"boolean","description":"阴遁/阳遁"},"ri_gan":{"type":"string","description":"日干","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"ri_zhi":{"type":"string","description":"日支","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"shi_gan":{"type":"string","description":"时干","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"shi_zhi":{"type":"string","description":"时支","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"zhi_fu_xing":{"type":"string","enum":["天蓬","天芮","天冲","天辅","天禽","天心","天柱","天任","天英"],"description":"值符星"},"zhi_shi_men":{"type":"string","enum":["休门","生门","伤门","杜门","景门","死门","惊门","开门"],"description":"值使门"},"gong_wei":{"type":"array","description":"九宫排盘结果","items":{"type":"object","properties":{"di_pan_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"],"description":"地盘干"},"tian_pan_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸",""],"description":"天盘干（中5宫虚空为空）"},"xing":{"type":"string","enum":["天蓬","天芮","天冲","天辅","天禽","天心","天柱","天任","天英"],"description":"九星"},"men":{"type":"string","enum":["休门","生门","伤门","杜门","景门","死门","惊门","开门"],"description":"八门"},"shen":{"type":"string","enum":["值符","螣蛇","太阴","六合","勾陈","朱雀","九地","九天","白虎","玄武"],"description":"八神（阳遁：勾陈/朱雀；阴遁：白虎/玄武）"},"an_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"],"description":"暗干"}},"required":["di_pan_gan"]}},"ma_xing":{"type":"string","enum":["坎","坤","震","巽","中","乾","兑","艮","离"],"description":"马星宫位"},"kong_wang":{"type":"array","items":{"type":"string","enum":["坎","坤","震","巽","中","乾","兑","艮","离"]},"description":"空亡宫位"}},"required":["jushu","yin_dun","ri_gan","ri_zhi"]},"patterns":{"type":"array"},"gan_interaction":{"type":"array","description":"天干关系"},"men_interaction":{"type":"array","description":"八门关系"},"xing_interaction":{"type":"array","description":"九星关系"},"wang_shuai":{"type":"array","description":"旺衰"},"men_po":{"type":"array","description":"门迫"},"men_zhi":{"type":"array","description":"门制"},"ying_qi":{"type":"object","description":"应期判断"},"ri_gan_gong":{"type":"string","description":"日干落宫（排盘固有）"},"shi_gan_gong":{"type":"string","description":"时干落宫（排盘固有）"},"ri_shi_sheng_ke":{"type":"string","description":"日干宫-时干宫五行生克（确定性派生）"},"kong_wang_affected":{"type":"boolean","description":"日干宫或时干宫是否空亡"},"ma_xing_affected":{"type":"boolean","description":"日干宫或时干宫是否马星"},"zhi_fu_xing_gong":{"type":"string","description":"值符星落宫（排盘固有）"},"zhi_shi_men_gong":{"type":"string","description":"值使门落宫（排盘固有）"},"yong_shen":{"type":"object","description":"用神领域对象（传入yong_shen符号组合时返回）","properties":{"nian_gan_gong":{"type":"string","description":"年命干落宫（本命根基，甲遁六仪，需birth_year）"},"symbols":{"type":"array","description":"用神符号组合落宫状态","items":{"type":"object","properties":{"symbol":{"type":"string","description":"符号名（开门/天辅/六合/戊）"},"palace":{"type":"string","description":"符号落宫"},"tian_gan":{"type":"string","description":"落宫天盘干（十干克应）"},"kong_wang":{"type":"boolean","description":"落宫是否空亡"},"ma_xing":{"type":"boolean","description":"落宫是否马星"}},"required":["symbol","palace"]}}}}},"required":["pan","patterns"]}`),
	},
	{
		Name: "bazhai.chart", Description: "八宅风水。排盘：命卦 + 四吉四凶方 + 流年紫白飞星（基准不再单列 minggua 方法）。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":{"type":"string","format":"date-time","description":"ISO 8601 时间"},"gender":{"type":"string","enum":["male","female"]}},"required":["solar_time","gender"]}`), Handler: bazhaiChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"ming_gua":{"type":"object"},"ba_zhai_dirs":{"type":"object"},"pillar_bagua":{"type":"array"},"liu_nian_xing":{"type":"object","description":"流年紫白飞星（与玄空共用 schema：year/ru_zhong/gong_wei）"}},"required":["ming_gua","ba_zhai_dirs","pillar_bagua"]}`),
	},
	{
		Name: "bazhai.layout", Description: "八宅门主灶配合。chart + 门/主/灶方位 → 各宫与命卦 match（东四西四同组=吉）。确定性计算。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object"},"door_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"门卦"},"master_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"主卧卦"},"stove_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"灶卦"}},"required":["chart","door_gua","master_gua","stove_gua"]}`),
		Handler: bazhaiLayoutHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"group":{"type":"string"},"ming_gua_str":{"type":"string"},"door":{"type":"object","description":"门卦信息: gua_number,gua_name,wuxing,group(东四/西四),match(吉/凶)"},"master":{"type":"object","description":"主(卧室)八卦信息, 同door结构"},"stove":{"type":"object","description":"灶(厨房)卦信息, 同door结构"}},"required":["group","ming_gua_str","door","master","stove"]}`),
	},
	{
		Name: "xuankong.chart", Description: "玄空飞星。返回山向飞星盘。zuo_shan/xiang_shan 为坐向（0-23）。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"zuo_shan":{"type":"integer","minimum":0,"maximum":23,"description":"山向"},"xiang_shan":{"type":"integer","minimum":0,"maximum":23,"description":"朝向"}},"required":["solar_time","zuo_shan","xiang_shan"]}`),
		Handler: xuankongChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yun":{"type":"object","description":"三元九运","properties":{"year":{"type":"integer","description":"当前年份"},"yuan":{"type":"string","description":"上元/中元/下元"},"yun_number":{"type":"integer","description":"运数1-9"},"yun_name":{"type":"string","description":"运名:一运/九运"},"start_year":{"type":"integer","description":"本运起始年"}},"required":["year","yuan","yun_number"]},"gong_wei":{"type":"array"},"wang_shan":{"type":"boolean"},"zuo_shan":{"type":"integer","description":"坐山(0-23)"},"xiang_shan":{"type":"integer","description":"向山(0-23)"},"shan_xing":{"type":"boolean","description":"双星会坐：坐宫山向星皆当令"},"wang_xiang":{"type":"boolean","description":"旺向：向宫向星=当令"},"xiang_xing":{"type":"boolean","description":"双星会向：向宫山向星皆当令"},"xing_jia_hui":{"type":"array","description":"星加会"},"shou_shan_chu_sha":{"type":"object","description":"收山出煞"},"fu_yin":{"type":"boolean","description":"伏吟（运盘）"},"fan_yin":{"type":"boolean","description":"反吟（运盘，恒false）"},"xia_shui":{"type":"boolean","description":"上山下水：向宫山星=当令且坐宫向星=当令"}},"required":["yun","gong_wei","wang_shan"]}`),
	},
	{
		Name: "xuankong.liunian", Description: "玄空流年飞星。chart（可选）+ year → 流年飞星盘 + 宅盘凶星落宫对照（确定性计算）。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"可选：xuankong.chart 返回的宅盘，给则叠加凶星落宫对照"},"year":{"type":"integer","description":"年份"}},"required":["year"]}`),
		Handler: xuankongLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"ru_zhong":{"type":"string","description":"入中星名"},"gong_wei":{"type":"array","description":"流年飞星盘（gong_num/xing/xing_name/wuxing/rating/ru_zhong）"},"house_overlay":{"type":"array","description":"流年凶星落宫对照宅盘该宫三星"}},"required":["year","ru_zhong","gong_wei"]}`),
	},
	{
		Name: "liuyao.qigua", Description: "六爻起卦。摇卦（三枚铜钱起六次），返回原始爻值和动爻位置。",
		Params: mustSchema(`{"type":"object","properties":{"seed":{"type":"integer","description":"可选：固定随机种子（测试用）"}},"required":[]}`),
		Handler: liuyaoQiguaHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yaos":{"type":"array","items":{"type":"integer"}, "description":"六爻值 6-9"},"dong_yao":{"type":"array","items":{"type":"integer"},"description":"动爻位置 1-6"}},"required":["yaos","dong_yao"]}`),
	},
	{
		Name: "liuyao.chart", Description: "六爻装卦。传入起卦结果和问事时辰，装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。lines.liu_qin: 0=父母 1=兄弟 2=官鬼 3=妻财 4=子孙；lines.liu_shou: 0=青龙 1=朱雀 2=勾陈 3=螣蛇 4=白虎 5=玄武；wang_shuai: 0=旺 1=相 2=休 3=囚 4=死",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"yong_shen":{"type":"string","description":"用神六亲（如 妻财/官鬼/父母/兄弟/子孙/世爻），可选，默认世爻"},"yaos":{"type":"array","items":{"type":"integer"},"minItems":6,"maxItems":6,"description":"六爻值（6-9），必填，先调 liuyao.qigua 获取"}},"required":["solar_time","yaos"]}`),
		Handler: liuyaoChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"ben_gua":{"type":"string","enum":["乾","姤","遁","否","观","晋","大有","剥","复","颐","屯","益","震","噬嗑","随","无妄","明夷","贲","既济","家人","丰","离","革","同人","临","损","节","中孚","归妹","睽","兑","履","泰","大畜","需","小畜","大壮","大有","夬","乾","姤","遁","否","观","晋","大有","剥","复","颐","屯","益","震","噬嗑","随","无妄","明夷","贲","既济","家人","丰","离","革","同人"]},"lines":{"type":"array","description":"每爻：六亲/六神/世应 + 确定性状态（yue_po 月破/dong_self 发动/dong_sheng 动爻生/dong_ke 动爻克）","items":{"type":"object","properties":{"position":{"type":"integer"},"type":{"type":"integer"},"gan":{"type":"string"},"zhi":{"type":"string"},"wuxing":{"type":"string"},"liu_qin":{"type":"string","enum":["父母","兄弟","官鬼","妻财","子孙"]},"shi_ying":{"type":"string","description":"世/应"},"liu_shou":{"type":"string","enum":["青龙","朱雀","勾陈","螣蛇","白虎","玄武"]},"yue_po":{"type":"boolean","description":"月破"},"dong_self":{"type":"boolean","description":"本爻发动"},"dong_sheng":{"type":"boolean","description":"有动爻生此爻"},"dong_ke":{"type":"boolean","description":"有动爻克此爻"},"xun_kong":{"type":"boolean","description":"该爻地支值日柱旬空"}},"required":["position","type","gan","zhi","wuxing","liu_qin","shi_ying","liu_shou"]}},"yong_shen":{"type":"object","description":"用神结果","properties":{"name":{"type":"string","description":"用神六亲"},"position":{"type":"integer","description":"爻位1-6（0=未找到）"},"wang_shuai":{"type":"string","description":"用神旺衰"},"yue_po":{"type":"boolean","description":"用神月破"},"xun_kong":{"type":"boolean","description":"用神旬空"},"mu_ku":{"type":"boolean","description":"用神入墓"},"liu_shou":{"type":"string","description":"用神临的六神","enum":["青龙","朱雀","勾陈","螣蛇","白虎","玄武"]},"fu_shen":{"type":"object","description":"飞伏（用神不现时）","properties":{"position":{"type":"integer","description":"爻位"},"liu_qin":{"type":"string","description":"伏神六亲"},"zhi":{"type":"string","description":"伏神地支"}},"required":["position","liu_qin","zhi"]}},"required":["name","position"]},"wang_shuai":{"type":"array"},"yue_jian_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"yue_jian_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"ying_qi":{"type":"object","description":"应期判断结果","properties":{"yong_shen":{"type":"string","description":"用神"},"dong_yao_pos":{"type":"integer","description":"动爻位置"},"ying_time":{"type":"string","description":"应期描述"},"assessment":{"type":"string","description":"综合判断"}},"required":["yong_shen","assessment"]},"bian_gua":{"type":"string","description":"变卦名"},"bian_yao":{"type":"array","description":"变爻"},"dong_yao":{"type":"array","items":{"type":"integer"},"description":"动爻位置"},"gong":{"type":"string","description":"八宫"},"gua_ci":{"type":"object","description":"卦辞爻辞"},"gong_wuxing":{"type":"string","description":"宫五行"},"ri_chen_gan":{"type":"string","description":"日辰天干"},"ri_chen_zhi":{"type":"string","description":"日辰地支"},"ri_chen_relations":{"type":"array","description":"日辰与爻关系"},"xun_kong":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"description":"日柱旬空地支（甲子旬空戌亥…）"},"dong_yao_relations":{"type":"array","description":"动爻与用神的关系","items":{"type":"object","properties":{"position":{"type":"integer","description":"动爻位置"},"relation":{"type":"string","description":"关系类型（生用/克用/比和/冲用/生原神/克原神/生忌神/克忌神）"}},"required":["position","relation"]}},"patterns":{"type":"array","description":"特殊格局","items":{"type":"object","properties":{"type":{"type":"string","description":"格局类型"},"sub_type":{"type":"string","description":"子类型"},"position":{"type":"integer"},"is_true":{"type":"boolean"},"assessment":{"type":"string"}},"required":["type","assessment"]}}},"required":["name","ben_gua","lines","yong_shen"]}`),
	},
	{
		Name: "huangli.days", Description: "黄历查日。返回连续N天的黄历信息（建除、黄道、二十八宿、时辰吉凶等）。",
		Params: mustSchema(`{"type":"object","properties":{"start_date":{"type":"string","description":"起始日期 YYYY-MM-DD"},"count":{"type":"integer","description":"查几天，默认3，最多30"}},"required":["start_date"]}`),
		Handler: huangliDaysHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"date":{"type":"string"},"ri_zhu":{"type":"object","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"nayin":{"type":"string"}},"required":["gan","zhi"]},"nayin":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土"]},"jian_chu":{"type":"string","enum":["建","除","满","平","定","执","破","危","成","收","开","闭"]},"huangdao":{"type":"object","properties":{"name":{"type":"string"},"path":{"type":"string"}},"required":["name","path"]},"xi_shen":{"type":"string"},"cai_shen":{"type":"string"},"fu_shen":{"type":"string"},"gan_ji":{"type":"string"},"zhi_ji":{"type":"string"},"mansion":{"type":"object","properties":{"name":{"type":"string"},"index":{"type":"integer"},"animal":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土","日","月"]},"group":{"type":"string"}},"required":["name","index"]}},"required":["date","ri_zhu"]}}`),
	},
	{
		Name: "time.now", Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Params: mustSchema(`{"type":"object","properties":{"seed":{"type":"integer","description":"可选：固定随机种子（测试用）"}},"required":[]}`),
		Handler: timeNowHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"utc":{"type":"string"},"local":{"type":"string"},"cst":{"type":"string"}},"required":["utc","cst"]}`),
	},
	{
		Name: "tianwen.time", Description: "根据时间和经度计算真太阳时，返回公历、真太阳时、农历三套时间。",
		Params: schemaTimePointParams(),
		Handler: tianwenTimeHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"solar":{"type":"string"},"gregorian":{"type":"string"},"lunar":{"type":"object","description":"农历信息: year/month/day/shichen","properties":{"year":{"type":"integer"},"month":{"type":"integer"},"day":{"type":"integer"},"leap":{"type":"boolean"},"shichen":{"type":"string"}}}},"required":["solar","gregorian","lunar"]}`),
	},
	{
		Name: "city.coords", Description: "根据城市名查询经纬度。支持中英文城市名，全球范围搜索。基于 Nominatim 服务。",
		Params: mustSchema(`{"type":"object","properties":{"city":{"type":"string","description":"城市名称（中英文均可）"}},"required":["city"]}`),
		Handler: cityCoordsHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"longitude":{"type":"number"},"latitude":{"type":"number"},"country":{"type":"string"}},"required":["name","longitude","latitude","country"]}`),
	},
}
