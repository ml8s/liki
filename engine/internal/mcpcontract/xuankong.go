package mcpcontract

import (
	"context"
	"fmt"
	"time"

	"liki-engine/internal/engine/fengshui"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/xuankong"
)

type XuankongChartInput struct {
	PeriodDate DateString    `json:"period_date" jsonschema:"宅运起盘日期，YYYY-MM-DD；年份 1864-2200。这不是命主出生时间。"`
	ZuoShan    MountainIndex `json:"zuo_shan" jsonschema:"坐山，二十四山 index：0=子,1=癸,2=丑,3=艮,4=寅,5=甲,6=卯,7=乙,8=辰,9=巽,10=巳,11=丙,12=午,13=丁,14=未,15=坤,16=申,17=庚,18=酉,19=辛,20=戌,21=乾,22=亥,23=壬。"`
	XiangShan  MountainIndex `json:"xiang_shan" jsonschema:"朝向；必须等于坐山后 12 山（180°相对）。"`
}

type XuankongYunOutput struct {
	Year         XuankongYear     `json:"year" jsonschema:"起盘公历年份。"`
	YearBoundary YearBoundaryName `json:"year_boundary" jsonschema:"当前实现按公历日历年换运；立春年界属另一流派，不在本盘隐含。"`
	Yuan         XuankongYuanName `json:"yuan" jsonschema:"三元元运。"`
	YunNumber    FlyingStarNumber `json:"yun_number" jsonschema:"运数。"`
	YunName      YunName          `json:"yun_name" jsonschema:"运名。"`
	StartYear    int              `json:"start_year" jsonschema:"运起始公历年份。"`
	EndYear      int              `json:"end_year" jsonschema:"运结束公历年份。"`
}

type XuankongStarOutput struct {
	Number FlyingStarNumber `json:"number" jsonschema:"紫白星数。"`
	Color  StarColorName    `json:"color" jsonschema:"紫白星颜色。"`
	Name   FlyingStarName   `json:"name" jsonschema:"紫白星名。"`
	Wuxing ElementName      `json:"wuxing" jsonschema:"星五行。"`
}

type XuankongPalaceOutput struct {
	GongNum   GuaNumber          `json:"gong_num" jsonschema:"洛书宫数。"`
	YunXing   XuankongStarOutput `json:"yun_xing" jsonschema:"运星。"`
	ShanXing  XuankongStarOutput `json:"shan_xing" jsonschema:"山星。"`
	XiangXing XuankongStarOutput `json:"xiang_xing" jsonschema:"向星。"`
}

type XuankongFourSituationOutput struct {
	Name               XuankongSituationName `json:"name" jsonschema:"四大局名称。"`
	SitMountainTimely  bool                  `json:"sit_mountain_timely" jsonschema:"坐宫山星是否当令。"`
	FaceFacingTimely   bool                  `json:"face_facing_timely" jsonschema:"向宫向星是否当令。"`
	SitFacingTimely    bool                  `json:"sit_facing_timely" jsonschema:"坐宫向星是否当令。"`
	FaceMountainTimely bool                  `json:"face_mountain_timely" jsonschema:"向宫山星是否当令。"`
}

type XuankongXingJiaHuiOutput struct {
	ShanNum  FlyingStarNumber `json:"shan_num" jsonschema:"山星数。"`
	XiangNum FlyingStarNumber `json:"xiang_num" jsonschema:"向星数。"`
	Name     string           `json:"name" jsonschema:"星曜加会名称。"`
	Meaning  string           `json:"meaning" jsonschema:"星曜加会通则含义。"`
}

type XuankongShouShanOutput struct {
	ZhengShen  FlyingStarNumber   `json:"zheng_shen" jsonschema:"当运正神星数。"`
	LingShen   FlyingStarNumber   `json:"ling_shen" jsonschema:"当运零神星数。"`
	ShouShan   bool               `json:"shou_shan" jsonschema:"理气收山是否得宜。"`
	ChuSha     bool               `json:"chu_sha" jsonschema:"理气出煞是否得宜。"`
	Assessment XuankongAssessment `json:"assessment" jsonschema:"收山出煞理气评估。"`
}

// XuankongChart is both the chart tool's public output and the audited
// optional input to liunian. Its digest makes a caller-modified chart fail.
type XuankongChart struct {
	Yun            XuankongYunOutput           `json:"yun" jsonschema:"三元九运。"`
	GongWei        [9]XuankongPalaceOutput     `json:"gong_wei" jsonschema:"九宫运星/山星/向星。"`
	FourSituation  XuankongFourSituationOutput `json:"four_situation" jsonschema:"四大局。"`
	FuYinLayers    []XuankongFuYinLayer        `json:"fu_yin_layers" jsonschema:"伏吟层。"`
	FanYinLayers   []XuankongFanYinLayer       `json:"fan_yin_layers" jsonschema:"反吟层。"`
	XingJiaHui     [9]XuankongXingJiaHuiOutput `json:"xing_jia_hui" jsonschema:"星加会。"`
	ChartDigest    ChartDigest                 `json:"chart_digest" jsonschema:"canonical SHA-256 chart digest。"`
	ZuoShan        MountainIndex               `json:"zuo_shan" jsonschema:"坐山(0-23)。"`
	ZuoShanName    MountainName                `json:"zuo_shan_name" jsonschema:"坐山名。"`
	XiangShan      MountainIndex               `json:"xiang_shan" jsonschema:"向山(0-23)。"`
	XiangShanName  MountainName                `json:"xiang_shan_name" jsonschema:"向山名。"`
	ShouShanChuSha XuankongShouShanOutput      `json:"shou_shan_chu_sha" jsonschema:"收山出煞理气判断。"`
}

type XuankongLiunianInput struct {
	Chart *XuankongChart `json:"chart,omitempty" jsonschema:"可选 xuankong.chart 宅盘；传入时 digest 必须匹配。"`
	Year  XuankongYear   `json:"year" jsonschema:"目标公历年份。"`
	Count SeriesCount    `json:"count,omitempty" jsonschema:"连续查询年数，单条或批量items数组"`
}

type XuankongAnnualStarOutput struct {
	GongNum  GuaNumber        `json:"gong_num" jsonschema:"洛书宫数。"`
	Xing     FlyingStarNumber `json:"xing" jsonschema:"紫白星数。"`
	XingName FlyingStarName   `json:"xing_name" jsonschema:"紫白星名。"`
	Wuxing   ElementName      `json:"wuxing" jsonschema:"星五行。"`
	RuZhong  bool             `json:"ru_zhong" jsonschema:"是否入中。"`
}

type XuankongHouseOverlayOutput struct {
	GongNum     GuaNumber      `json:"gong_num" jsonschema:"洛书宫数。"`
	Star        FlyingStarName `json:"star" jsonschema:"流年星名。"`
	PalaceStars string         `json:"palace_stars" jsonschema:"宅盘该宫运星/山星/向星。"`
}

type XuankongLiunianOutput struct {
	Year         XuankongYear                 `json:"year" jsonschema:"目标公历年份。"`
	RuZhong      FlyingStarName               `json:"ru_zhong" jsonschema:"入中紫白星。"`
	GongWei      [9]XuankongAnnualStarOutput  `json:"gong_wei" jsonschema:"九宫流年紫白星。"`
	HouseOverlay []XuankongHouseOverlayOutput `json:"house_overlay,omitempty" jsonschema:"传入宅盘时，全部流年星落宫对照宅盘该宫三星；不预设星曜固有吉凶。"`
}

func XuankongChartCompute(_ context.Context, input XuankongChartInput) (XuankongChart, error) {
	periodDate, err := time.Parse("2006-01-02", string(input.PeriodDate))
	if err != nil {
		return XuankongChart{}, fmt.Errorf("xuankong.chart: invalid period_date %q: %w", input.PeriodDate, err)
	}
	if periodDate.Year() < 1864 || periodDate.Year() > 2200 {
		return XuankongChart{}, fmt.Errorf("xuankong.chart: period_date year %d out of range (1864-2200)", periodDate.Year())
	}
	if (int(input.ZuoShan)+12)%24 != int(input.XiangShan) {
		return XuankongChart{}, fmt.Errorf(
			"xuankong.chart: xiang_shan must be opposite zuo_shan by 12 mountains, got %d/%d",
			input.ZuoShan, input.XiangShan,
		)
	}
	chart := xuankong.ComputeChart(
		time.Time(periodDate), int(input.ZuoShan), int(input.XiangShan),
	)
	return xuankongChartWire(chart), nil
}

func XuankongLiunian(_ context.Context, input XuankongLiunianInput) (any, error) {
	return runSeries(input, int(input.Count), MaxSeriesCount,
		func(in XuankongLiunianInput) (XuankongLiunianOutput, error) {
			return computeXuankongLiunian(in)
		},
		func(in XuankongLiunianInput) (XuankongLiunianInput, error) {
			in.Year++
			return in, nil
		},
	)
}

func computeXuankongLiunian(input XuankongLiunianInput) (XuankongLiunianOutput, error) {
	var chart *xuankong.Chart
	if input.Chart != nil {
		domainChart, err := xuankongChartDomain(*input.Chart)
		if err != nil {
			return XuankongLiunianOutput{}, err
		}
		chart = &domainChart
	}
	result := xuankong.ComputeLiuNian(int(input.Year), chart)
	return xuankongLiunianWire(result), nil
}

func xuankongStarWire(star fengshui.FlyingStar) XuankongStarOutput {
	return XuankongStarOutput{
		Number: FlyingStarNumber(star.Number),
		Color:  StarColorName(star.Color),
		Name:   FlyingStarName(star.Name),
		Wuxing: ElementName(star.Element.String()),
	}
}

func xuankongChartWire(chart xuankong.Chart) XuankongChart {
	output := XuankongChart{
		Yun: XuankongYunOutput{
			Year:         XuankongYear(chart.Yun.Year),
			YearBoundary: YearBoundaryName(chart.Yun.YearBoundary),
			Yuan:         XuankongYuanName(chart.Yun.Yuan),
			YunNumber:    FlyingStarNumber(chart.Yun.YunNumber),
			YunName:      YunName(chart.Yun.YunName),
			StartYear:    chart.Yun.StartYear,
			EndYear:      chart.Yun.EndYear,
		},
		FourSituation: XuankongFourSituationOutput{
			Name:               XuankongSituationName(chart.FourSituation.Name),
			SitMountainTimely:  chart.FourSituation.SitMountainTimely,
			FaceFacingTimely:   chart.FourSituation.FaceFacingTimely,
			SitFacingTimely:    chart.FourSituation.SitFacingTimely,
			FaceMountainTimely: chart.FourSituation.FaceMountainTimely,
		},
		ChartDigest:   ChartDigest(chart.ChartDigest),
		ZuoShan:       MountainIndex(chart.SitMountain),
		ZuoShanName:   MountainName(chart.ZuoShanName),
		XiangShan:     MountainIndex(chart.FaceMountain),
		XiangShanName: MountainName(chart.XiangShanName),
		ShouShanChuSha: XuankongShouShanOutput{
			ZhengShen:  FlyingStarNumber(chart.ShouShanChuSha.ZhengShen),
			LingShen:   FlyingStarNumber(chart.ShouShanChuSha.LingShen),
			ShouShan:   chart.ShouShanChuSha.ShouShanOK,
			ChuSha:     chart.ShouShanChuSha.ChuShaOK,
			Assessment: XuankongAssessment(chart.ShouShanChuSha.Assessment),
		},
	}
	for i, palace := range chart.Palaces {
		output.GongWei[i] = XuankongPalaceOutput{
			GongNum:   GuaNumber(palace.PalaceNum),
			YunXing:   xuankongStarWire(palace.PeriodStar),
			ShanXing:  xuankongStarWire(palace.MountainStar),
			XiangXing: xuankongStarWire(palace.FacingStar),
		}
	}
	for i, item := range chart.XingJiaHui {
		output.XingJiaHui[i] = XuankongXingJiaHuiOutput{
			ShanNum:  FlyingStarNumber(item.ShanNum),
			XiangNum: FlyingStarNumber(item.XiangNum),
			Name:     item.Name,
			Meaning:  item.Meaning,
		}
	}
	for _, layer := range chart.FuYinLayers {
		output.FuYinLayers = append(output.FuYinLayers, XuankongFuYinLayer(layer))
	}
	for _, layer := range chart.FanYinLayers {
		output.FanYinLayers = append(output.FanYinLayers, XuankongFanYinLayer(layer))
	}
	return output
}

func xuankongChartDomain(input XuankongChart) (xuankong.Chart, error) {
	var chart xuankong.Chart
	chart.Yun = xuankong.SanYuanYun{
		Year:         int(input.Yun.Year),
		Yuan:         string(input.Yun.Yuan),
		YunNumber:    int(input.Yun.YunNumber),
		YunName:      string(input.Yun.YunName),
		StartYear:    input.Yun.StartYear,
		EndYear:      input.Yun.EndYear,
		YearBoundary: string(input.Yun.YearBoundary),
	}
	for i, palace := range input.GongWei {
		yunXing, err := xuankongStarDomain(palace.YunXing)
		if err != nil {
			return xuankong.Chart{}, fmt.Errorf("xuankong.liunian: palace %d yun_xing: %w", palace.GongNum, err)
		}
		shanXing, err := xuankongStarDomain(palace.ShanXing)
		if err != nil {
			return xuankong.Chart{}, fmt.Errorf("xuankong.liunian: palace %d shan_xing: %w", palace.GongNum, err)
		}
		xiangXing, err := xuankongStarDomain(palace.XiangXing)
		if err != nil {
			return xuankong.Chart{}, fmt.Errorf("xuankong.liunian: palace %d xiang_xing: %w", palace.GongNum, err)
		}
		chart.Palaces[i].PalaceNum = int(palace.GongNum)
		chart.Palaces[i].PeriodStar = yunXing
		chart.Palaces[i].MountainStar = shanXing
		chart.Palaces[i].FacingStar = xiangXing
	}
	chart.FourSituation = xuankong.FourSituation{
		Name:               string(input.FourSituation.Name),
		SitMountainTimely:  input.FourSituation.SitMountainTimely,
		FaceFacingTimely:   input.FourSituation.FaceFacingTimely,
		SitFacingTimely:    input.FourSituation.SitFacingTimely,
		FaceMountainTimely: input.FourSituation.FaceMountainTimely,
	}
	chart.FuYinLayers = []string{}
	chart.FanYinLayers = []string{}
	for _, layer := range input.FuYinLayers {
		chart.FuYinLayers = append(chart.FuYinLayers, string(layer))
	}
	for _, layer := range input.FanYinLayers {
		chart.FanYinLayers = append(chart.FanYinLayers, string(layer))
	}
	for i, item := range input.XingJiaHui {
		chart.XingJiaHui[i].ShanNum = int(item.ShanNum)
		chart.XingJiaHui[i].XiangNum = int(item.XiangNum)
		chart.XingJiaHui[i].Name = item.Name
		chart.XingJiaHui[i].Meaning = item.Meaning
	}
	chart.SitMountain = int(input.ZuoShan)
	chart.ZuoShanName = string(input.ZuoShanName)
	chart.FaceMountain = int(input.XiangShan)
	chart.XiangShanName = string(input.XiangShanName)
	chart.ShouShanChuSha.ZhengShen = int(input.ShouShanChuSha.ZhengShen)
	chart.ShouShanChuSha.LingShen = int(input.ShouShanChuSha.LingShen)
	chart.ShouShanChuSha.ShouShanOK = input.ShouShanChuSha.ShouShan
	chart.ShouShanChuSha.ChuShaOK = input.ShouShanChuSha.ChuSha
	chart.ShouShanChuSha.Assessment = string(input.ShouShanChuSha.Assessment)
	chart.ChartDigest = string(input.ChartDigest)
	if err := chart.ValidateChartDigest(); err != nil {
		return xuankong.Chart{}, fmt.Errorf("xuankong.liunian: %w", err)
	}
	return chart, nil
}

func xuankongStarDomain(star XuankongStarOutput) (fengshui.FlyingStar, error) {
	element, err := ganzhi.ParseWuxing(string(star.Wuxing))
	if err != nil {
		return fengshui.FlyingStar{}, err
	}
	return fengshui.FlyingStar{
		Number:  int(star.Number),
		Color:   string(star.Color),
		Name:    string(star.Name),
		Element: element,
	}, nil
}

func xuankongLiunianWire(result xuankong.LiuNianResult) XuankongLiunianOutput {
	output := XuankongLiunianOutput{
		Year:    XuankongYear(result.Year),
		RuZhong: FlyingStarName(result.RuZhong),
	}
	for i, star := range result.GongWei {
		if i >= len(output.GongWei) {
			break
		}
		output.GongWei[i] = XuankongAnnualStarOutput{
			GongNum:  GuaNumber(star.GongNum),
			Xing:     FlyingStarNumber(star.Xing),
			XingName: FlyingStarName(star.XingName),
			Wuxing:   ElementName(star.Wuxing),
			RuZhong:  star.RuZhong,
		}
	}
	for _, overlay := range result.HouseOverlay {
		output.HouseOverlay = append(output.HouseOverlay, XuankongHouseOverlayOutput{
			GongNum:     GuaNumber(overlay.GongNum),
			Star:        FlyingStarName(overlay.Star),
			PalaceStars: overlay.PalaceStars,
		})
	}
	return output
}
