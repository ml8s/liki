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
		SolarTime string `json:"solar_time"`
		Kind      string `json:"kind"`
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
	result := qimen.ComputeChart(st, kind)
	return wrapResult("qimen", result)
}

func qimenJudgmentHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
		Event string          `json:"event"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qimen.judgment: %w", err)
	}
	if p.Event == "" {
		p.Event = "general"
	}
	event, err := qimen.ParseEventKind(p.Event)
	if err != nil {
		return nil, fmt.Errorf("qimen.judgment: %w", err)
	}

	// Unmarshal the chart.
	var chart qimen.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("qimen.judgment: invalid chart: %w", err)
	}

	result := qimen.ComputeJudgment(chart, event)
	return wrapResult("qimen_judgment", result)
}

func qimenSelectHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Event     string `json:"event"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Count     int    `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qimen.select: %w", err)
	}
	if p.Event == "" {
		p.Event = "general"
	}
	event, err := qimen.ParseEventKind(p.Event)
	if err != nil {
		return nil, fmt.Errorf("qimen.select: %w", err)
	}
	if p.Count <= 0 {
		p.Count = 3
	}

	start, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return nil, fmt.Errorf("qimen.select: invalid start_date: %w", err)
	}
	end, err := time.Parse("2006-01-02", p.EndDate)
	if err != nil {
		return nil, fmt.Errorf("qimen.select: invalid end_date: %w", err)
	}

	slots := qimen.ComputeTimeSelection(start, end, event, p.Count, 116.4, 8)
	return wrapResult("qimen_select", slots)
}


func liuyaoJudgmentHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
		Event string          `json:"event"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.judgment: %w", err)
	}
	var chart liuyao.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("liuyao.judgment: invalid chart: %w", err)
	}
	result := liuyao.ComputeJudgment(chart, p.Event)
	return wrapResult("liuyao_judgment", result)
}
// ── bazhai ──

func bazhaiJudgmentHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart     json.RawMessage `json:"chart"`
		DoorGua   string          `json:"door_gua"`
		MasterGua string          `json:"master_gua"`
		StoveGua  string          `json:"stove_gua"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("bazhai.judgment: %w", err)
	}
	var chart bazhai.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("bazhai.judgment: parse chart: %w", err)
	}
	result := bazhai.ComputeJudgment(chart, p.DoorGua, p.MasterGua, p.StoveGua)
	return wrapResult("bazhai_judgment", result)
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

func bazhaiMingguaHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Gender    ganzhi.Gender `json:"gender"`
		BirthYear int            `json:"birth_year"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("bazhai.minggua: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("bazhai.minggua: %w", err)
	}
	result := bazhai.ComputeMingGua(p.Gender, p.BirthYear)
	return wrapResult("minggua", result)
}

// ── xuankong ──

func xuankongAnnualHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year int `json:"year"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("xuankong.annual: %w", err)
	}
	if p.Year < 1900 || p.Year > 2200 {
		return nil, fmt.Errorf("xuankong.annual: year %d out of range (1900-2200)", p.Year)
	}
	result := xuankong.ComputeAnnual(p.Year)
	return wrapResult("xuankong_annual", result)
}

func xuankongChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime    string `json:"solar_time"`
		SitMountain  *int   `json:"sit_mountain"`
		FaceMountain *int   `json:"face_mountain"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("xuankong.chart: %w", err)
	}
	if p.SitMountain == nil {
		return nil, fmt.Errorf("xuankong.chart: sit_mountain is required")
	}
	if p.FaceMountain == nil {
		return nil, fmt.Errorf("xuankong.chart: face_mountain is required")
	}
	if *p.SitMountain < 0 || *p.SitMountain > 23 {
		return nil, fmt.Errorf("xuankong.chart: sit_mountain must be 0-23, got %d", *p.SitMountain)
	}
	if *p.FaceMountain < 0 || *p.FaceMountain > 23 {
		return nil, fmt.Errorf("xuankong.chart: face_mountain must be 0-23, got %d", *p.FaceMountain)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("xuankong.chart: %w", err)
	}
	result := xuankong.ComputeChart(st, *p.SitMountain, *p.FaceMountain)
	return wrapResult("xuankong", result)
}

func xuankongSanyuanHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year int `json:"year"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("xuankong.sanyuan: %w", err)
	}
	if p.Year <= 0 {
		return nil, fmt.Errorf("xuankong.sanyuan: year must be positive, got %d", p.Year)
	}
	result := xuankong.ComputeSanYuanYun(p.Year)
	return wrapResult("sanyuan", result)
}

// ── liuyao ──

func liuyaoQiguaHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct{}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.qigua: %w", err)
	}
	result := liuyao.Qigua()
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

func cityHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	result, err := city.SearchCity(ctx, raw)
	if err != nil {
		return nil, err
	}
	return wrapResult("city", result)
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
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"kind":{"type":"string","enum":["shi","ri","yue","nian"],"description":"奇门类型，默认 shi"}},"required":["solar_time"]}`),
		Handler: qimenChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"pan":{"type":"object","description":"奇门排盘结果","properties":{"jushu":{"type":"integer","description":"局数"},"yin_dun":{"type":"boolean","description":"阴遁/阳遁"},"ri_gan":{"type":"string","description":"日干","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"ri_zhi":{"type":"string","description":"日支","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"drive_gan":{"type":"string","description":"时干","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"drive_zhi":{"type":"string","description":"时支","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"zhi_fu_xing":{"type":"string","enum":["天蓬","天芮","天冲","天辅","天禽","天心","天柱","天任","天英"],"description":"值符星"},"zhi_shi_men":{"type":"string","enum":["休门","生门","伤门","杜门","景门","死门","惊门","开门"],"description":"值使门"},"gong_wei":{"type":"array","description":"九宫排盘结果","items":{"type":"object","properties":{"di_pan_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"],"description":"地盘干"},"tian_pan_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"],"description":"天盘干"},"xing":{"type":"string","enum":["天蓬","天芮","天冲","天辅","天禽","天心","天柱","天任","天英"],"description":"九星"},"door":{"type":"string","enum":["休门","生门","伤门","杜门","景门","死门","惊门","开门"],"description":"八门"},"spirit":{"type":"string","enum":["值符","螣蛇","太阴","六合","勾陈","朱雀","九地","九天","白虎","玄武"],"description":"八神（阳遁：勾陈/朱雀；阴遁：白虎/玄武）"},"hidden_stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"],"description":"暗干"}},"required":["di_pan_gan","tian_pan_gan"]}},"ma_xing":{"type":"string","enum":["坎","坤","震","巽","中","乾","兑","艮","离"],"description":"马星宫位"},"kong_wang":{"type":"array","items":{"type":"string","enum":["坎","坤","震","巽","中","乾","兑","艮","离"]},"description":"空亡宫位"}},"required":["jushu","yin_dun","ri_gan","ri_zhi"]},"patterns":{"type":"array"}},"required":["pan","patterns"]}`),
	},
	{
		Name: "qimen.judgment", Description: "奇门断事。基于排盘结果和事件类型进行综合分析判断。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"qimen.chart 返回的完整 chart 对象"},"event":{"type":"string","enum":["general","career","wealth","relationship","study","travel","health","legal"],"description":"事件类型，默认 general"}},"required":["chart"]}`),
		Handler: qimenJudgmentHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"subject_palace":{"type":"integer"},"event_palace":{"type":"integer"},"rating":{"type":"string"},"advice":{"type":"string"}},"required":["subject_palace","event_palace","rating","advice"]}`),
	},
	{
		Name: "qimen.select", Description: "奇门择吉。在指定日期范围内查找最佳时辰。",
		Params: mustSchema(`{"type":"object","properties":{"event":{"type":"string","enum":["general","career","wealth","relationship","study","travel","health","legal"],"description":"事件类型，默认 general","examples":["travel"]},"start_date":{"type":"string","description":"开始日期 YYYY-MM-DD","examples":["2026-07-20"]},"end_date":{"type":"string","description":"结束日期 YYYY-MM-DD","examples":["2026-07-22"]},"count":{"type":"integer","description":"推荐数量，默认 3","examples":[5]}},"required":["start_date","end_date"]}`),
		Handler: qimenSelectHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"date":{"type":"string"},"time":{"type":"string"},"shi_chen":{"type":"string"},"rating":{"type":"string"},"advice":{"type":"string"}},"required":["date","time","shi_chen","rating","advice"]}}`),
	},
	{
		Name: "liuyao.judgment", Description: "六爻断卦。对装卦结果进行综合分析判断，返回用神状态、评级和断语。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"liuyao.chart 返回的完整 chart 对象（非旧版 chart_data 格式）。必传。"},"event":{"type":"string","enum":["general","career","wealth","relationship","study","health","legal","travel"],"description":"事件类型，默认general"}},"required":["chart"]}`),
		Handler: liuyaoJudgmentHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yong_shen":{"type":"object","properties":{"name":{"type":"string"},"position":{"type":"integer"},"month":{"type":"string"},"day":{"type":"string"},"ri_chen_power":{"type":"string"},"is_dong":{"type":"boolean"}},"required":["name","position","month","day"]},"rating":{"type":"string"},"advice":{"type":"string"}},"required":["yong_shen","rating","advice"]}`),
	},
	{
		Name: "bazhai.judgment", Description: "八宅门主灶论断。分析门/主/灶与命卦配合吉凶。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object"},"door_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"门卦"},"master_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"主卧卦"},"stove_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"灶卦"}},"required":["chart","door_gua","master_gua","stove_gua"]}`),
		Handler: bazhaiJudgmentHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"group":{"type":"string"},"ming_gua_str":{"type":"string"},"door":{"type":"object","description":"门卦信息: gua_number,gua_name,wuxing,group(东四/西四),match(吉/凶)"},"master":{"type":"object","description":"主(卧室)八卦信息, 同door结构"},"stove":{"type":"object","description":"灶(厨房)卦信息, 同door结构"},"rating":{"type":"string"},"summary":{"type":"string"}},"required":["group","rating","summary"]}`),
	},
	{
		Name: "bazhai.chart", Description: "八宅风水。综合命卦与飞星分析。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":{"type":"string","format":"date-time","description":"ISO 8601 时间"},"gender":{"type":"string","enum":["male","female"]}},"required":["solar_time","gender"]}`), Handler: bazhaiChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"ming_gua":{"type":"object"},"ba_zhai_dirs":{"type":"object"},"pillar_bagua":{"type":"array"}},"required":["ming_gua","ba_zhai_dirs","pillar_bagua"]}`),
	},
	{
		Name: "bazhai.minggua", Description: "命卦查询。返回东四命/西四命 + 命卦 + 四吉四凶方。",
		Params: mustSchema(`{"type":"object","properties":{"gender":` + schemaGender + `,"birth_year":{"type":"integer","description":"出生年份"}},"required":["gender","birth_year"]}`),
		Handler: bazhaiMingguaHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"gua":{"type":"object","properties":{"index":{"type":"integer","description":"洛书数1-9"},"name":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"卦名"},"wuxing":{"type":"string","enum":["金","木","水","火","土"]},"yin_yang":{"type":"string","enum":["阳","阴"]}},"required":["index","name","wuxing","yin_yang"]},"group":{"type":"string"}},"required":["gua","group"]}`),
	},
	{
		Name: "xuankong.annual", Description: "玄空流年飞星。返回每年入中星、各宫飞星分布、吉凶评级。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"年份"}},"required":["year"]}`),
		Handler: xuankongAnnualHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"ru_zhong":{"type":"string","enum":["贪狼","巨门","禄存","文曲","廉贞","武曲","破军","左辅","右弼"]},"xing_yao":{"type":"array"}},"required":["year","ru_zhong","xing_yao"]}`),
	},
	{
		Name: "xuankong.sanyuan", Description: "三元九运查询。返回当前三元九运的时间表。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"年份"}},"required":["year"]}`),
		Handler: xuankongSanyuanHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"yuan":{"type":"string"},"yun_number":{"type":"integer"}},"required":["year","yuan","yun_number"]}`),
	},
	{
		Name: "xuankong.chart", Description: "玄空飞星。返回山向飞星盘。sit_mountain/face_mountain 为坐向（0-23）。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"sit_mountain":{"type":"integer","minimum":0,"maximum":23,"description":"山向"},"face_mountain":{"type":"integer","minimum":0,"maximum":23,"description":"朝向"}},"required":["solar_time","sit_mountain","face_mountain"]}`),
		Handler: xuankongChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yun":{"type":"object","description":"三元九运","properties":{"year":{"type":"integer","description":"当前年份"},"yuan":{"type":"string","description":"上元/中元/下元"},"yun_number":{"type":"integer","description":"运数1-9"},"yun_name":{"type":"string","description":"运名:一运/九运"},"start_year":{"type":"integer","description":"本运起始年"}},"required":["year","yuan","yun_number"]},"gong_wei":{"type":"array"},"wang_shan":{"type":"boolean"}},"required":["yun","gong_wei","wang_shan"]}`),
	},
	{
		Name: "liuyao.qigua", Description: "六爻起卦。摇卦（三枚铜钱起六次），返回原始爻值和动爻位置。",
		Params: mustSchema(`{"type":"object","properties":{},"required":[]}`),
		Handler: liuyaoQiguaHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yaos":{"type":"array","items":{"type":"integer"}, "description":"六爻值 6-9"},"dong_yao":{"type":"array","items":{"type":"integer"},"description":"动爻位置 1-6"}},"required":["yaos","dong_yao"]}`),
	},
	{
		Name: "liuyao.chart", Description: "六爻装卦。传入起卦结果和问事时辰，装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。lines.liu_qin: 0=父母 1=兄弟 2=官鬼 3=妻财 4=子孙；lines.liu_shou: 0=青龙 1=朱雀 2=勾陈 3=螣蛇 4=白虎 5=玄武；wang_shuai: 0=旺 1=相 2=休 3=囚 4=死",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"yong_shen":{"type":"string","description":"用神六亲（如 妻财/官鬼/父母/兄弟/子孙/世爻），可选，默认世爻"},"yaos":{"type":"array","items":{"type":"integer"},"minItems":6,"maxItems":6,"description":"六爻值（6-9），必填，先调 liuyao.qigua 获取"}},"required":["solar_time","yaos"]}`),
		Handler: liuyaoChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"ben_gua":{"type":"string","enum":["乾","姤","遁","否","观","晋","大有","剥","复","颐","屯","益","震","噬嗑","随","无妄","明夷","贲","既济","家人","丰","离","革","同人","临","损","节","中孚","归妹","睽","兑","履","泰","大畜","需","小畜","大壮","大有","夬","乾","姤","遁","否","观","晋","大有","剥","复","颐","屯","益","震","噬嗑","随","无妄","明夷","贲","既济","家人","丰","离","革","同人"]},"lines":{"type":"array"},"yong_shen":{"type":"object","description":"用神状态","properties":{"name":{"type":"string","description":"用神六亲"},"position":{"type":"integer","description":"爻位1-6"},"month":{"type":"string","description":"月令旺衰","enum":["旺","相","休","囚","死"],"examples":["旺"]},"day":{"type":"string","description":"日建关系:生/扶/克/冲/合/平"},"ri_chen_power":{"type":"string","description":"日建力","enum":["旺","平","衰"],"examples":["旺"]},"is_dong":{"type":"boolean","description":"是否为动爻"},"is_yue_po":{"type":"boolean","description":"月破"},"is_chi_shi":{"type":"boolean","description":"持世"},"dong_sheng":{"type":"boolean","description":"动爻生用神"},"dong_ke":{"type":"boolean","description":"动爻克用神"}}},"wang_shuai":{"type":"array"},"yue_jian_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"yue_jian_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"ying_qi":{"type":"object","description":"应期判断结果","properties":{"yong_shen":{"type":"string","description":"用神"},"dong_yao_pos":{"type":"integer","description":"动爻位置"},"ying_time":{"type":"string","description":"应期描述"},"assessment":{"type":"string","description":"综合判断"}},"required":["yong_shen","assessment"]}},"required":["name","ben_gua","lines","yong_shen"]}`),
	},
	{
		Name: "huangli.days", Description: "黄历查日。返回连续N天的黄历信息（建除、黄道、二十八宿、时辰吉凶等）。",
		Params: mustSchema(`{"type":"object","properties":{"start_date":{"type":"string","description":"起始日期 YYYY-MM-DD"},"count":{"type":"integer","description":"查几天，默认3，最多30"}},"required":["start_date"]}`),
		Handler: huangliDaysHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"date":{"type":"string"},"ri_zhu":{"type":"object","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"nayin":{"type":"string"}},"required":["gan","zhi"]},"nayin":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土"]},"jian_chu":{"type":"string","enum":["建","除","满","平","定","执","破","危","成","收","开","闭"]},"huangdao":{"type":"object","properties":{"name":{"type":"string"},"path":{"type":"string"}},"required":["name","path"]},"xi_shen":{"type":"string"},"cai_shen":{"type":"string"},"fu_shen":{"type":"string"},"gan_ji":{"type":"string"},"zhi_ji":{"type":"string"},"mansion":{"type":"object","properties":{"name":{"type":"string"},"index":{"type":"integer"},"animal":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土","日","月"]},"group":{"type":"string"}},"required":["name","index"]}},"required":["date","ri_zhu"]}}`),
	},
	{
		Name: "time.now", Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Params: mustSchema(`{"type":"object","properties":{},"required":[]}`),
		Handler: timeNowHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"utc":{"type":"string"},"local":{"type":"string"},"cst":{"type":"string"}},"required":["utc","cst"]}`),
	},
	{
		Name: "tianwen.time", Description: "根据时间和经度计算真太阳时，返回公历、真太阳时、农历三套时间。",
		Params: schemaTimePointParams(),
		Handler: tianwenTimeHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"solar":{"type":"string"},"gregorian":{"type":"string"},"lunar":{"type":"object","properties":{"year":{"type":"integer"},"month":{"type":"integer"},"day":{"type":"integer"},"leap":{"type":"boolean"},"shichen":{"type":"string"}}}},"required":["solar","gregorian","lunar"]}`),
	},
	{
		Name: "city", Description: "根据城市名查询经纬度。基于 Nominatim 服务。",
		Params: mustSchema(`{"type":"object","properties":{"city":{"type":"string","description":"城市名称"}},"required":["city"]}`),
		Handler: cityHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"longitude":{"type":"number"},"latitude":{"type":"number"},"country":{"type":"string"}},"required":["name","longitude","latitude","country"]}`),
	},
}
