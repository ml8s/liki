package mcpcontract

import (
	"context"
	"fmt"

	"liki-engine/internal/engine/fengshui"
	"liki-engine/internal/engine/tianwen"
	"liki-engine/internal/engine/xuankong"
)

// XuankongZhaiInput is the property identity required to recompute a Xuankong
// chart: the period-start moment and opposite sitting/facing mountains.
type XuankongZhaiInput struct {
	YunStartTime StandardTime  `json:"yun_start_time" jsonschema:"宅运起盘地标准日期时间（无 DST），RFC3339 带 offset。"`
	ZuoShan      MountainIndex `json:"zuo_shan" jsonschema:"坐山，二十四山 index：0=子,1=癸,2=丑,3=艮,4=寅,5=甲,6=卯,7=乙,8=辰,9=巽,10=巳,11=丙,12=午,13=丁,14=未,15=坤,16=申,17=庚,18=酉,19=辛,20=戌,21=乾,22=亥,23=壬。"`
	XiangShan    MountainIndex `json:"xiang_shan" jsonschema:"朝向；必须等于坐山后 12 山（180°相对）。"`
}

type XuankongChartInput struct {
	Zhai XuankongZhaiInput `json:"zhai" jsonschema:"玄空宅的事实。"`
}

type XuankongYunOutput struct {
	Year         XuankongYear     `json:"year" jsonschema:"立春年份。"`
	YearBoundary YearBoundaryName `json:"year_boundary" jsonschema:"年界：立春。"`
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

// XuankongChart is the chart tool's public output. It is not a reusable method
// input; callers pass zhai again when another Xuankong fact is needed.
type XuankongChart struct {
	Yun            XuankongYunOutput           `json:"yun" jsonschema:"三元九运。"`
	GongWei        [9]XuankongPalaceOutput     `json:"gong_wei" jsonschema:"九宫运星/山星/向星。"`
	FourSituation  XuankongFourSituationOutput `json:"four_situation" jsonschema:"四大局。"`
	FuYinLayers    []XuankongFuYinLayer        `json:"fu_yin_layers" jsonschema:"伏吟层。"`
	FanYinLayers   []XuankongFanYinLayer       `json:"fan_yin_layers" jsonschema:"反吟层。"`
	XingJiaHui     [9]XuankongXingJiaHuiOutput `json:"xing_jia_hui" jsonschema:"星加会。"`
	ZuoShan        MountainIndex               `json:"zuo_shan" jsonschema:"坐山(0-23)。"`
	ZuoShanName    MountainName                `json:"zuo_shan_name" jsonschema:"坐山名。"`
	XiangShan      MountainIndex               `json:"xiang_shan" jsonschema:"向山(0-23)。"`
	XiangShanName  MountainName                `json:"xiang_shan_name" jsonschema:"向山名。"`
	ShouShanChuSha XuankongShouShanOutput      `json:"shou_shan_chu_sha" jsonschema:"收山出煞理气判断。"`
}

type XuankongLiunianInput struct {
	Zhai             XuankongZhaiInput `json:"zhai" jsonschema:"玄空宅的事实；流年与其宅盘叠加。"`
	LiuNianStartTime StandardTime      `json:"liu_nian_start_time" jsonschema:"流年选择地标准时间（无 DST），RFC3339 带 offset。"`
	LiuNianCount     *SeriesCount      `json:"liu_nian_count,omitempty" jsonschema:"连续查询年数。"`
}

type XuankongLiuNianJie struct {
	LiuNianStartTime StandardTime `json:"liu_nian_start_time" jsonschema:"流年选择器时区标准时间（无 DST）。"`
	LichunYear       int          `json:"lichun_year" jsonschema:"该瞬间所在的立春年。"`
	LichunStart      DateString   `json:"lichun_start" jsonschema:"立春年起始公历日（含）。"`
	LichunEnd        DateString   `json:"lichun_end" jsonschema:"下一立春日（不含）。"`
	BoundaryDay      bool         `json:"boundary_day" jsonschema:"输入瞬间的当地日期是否为立春边界日。"`
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
	Year         XuankongYear                 `json:"year" jsonschema:"立春年份。"`
	RuZhong      FlyingStarName               `json:"ru_zhong" jsonschema:"入中紫白星。"`
	GongWei      [9]XuankongAnnualStarOutput  `json:"gong_wei" jsonschema:"九宫流年紫白星。"`
	HouseOverlay []XuankongHouseOverlayOutput `json:"house_overlay,omitempty" jsonschema:"全部流年星落宫对照宅盘该宫三星；不预设星曜固有吉凶。"`
	NianJie      XuankongLiuNianJie           `json:"nian_jie" jsonschema:"玄空流年立春年界。"`
}

func XuankongChartCompute(_ context.Context, input XuankongChartInput) (XuankongChart, error) {
	chart, err := xuankongComputeChart(input.Zhai)
	if err != nil {
		return XuankongChart{}, fmt.Errorf("xuankong.chart: %w", err)
	}
	output := xuankongChartWire(chart)
	return output, nil
}

func XuankongLiunian(_ context.Context, input XuankongLiunianInput) (XuankongLiuNianSeries, error) {
	items, err := runSeries(input, seriesCountValue(input.LiuNianCount), MaxSeriesCount,
		func(in XuankongLiunianInput) (XuankongLiunianOutput, error) {
			return computeXuankongLiunian(in)
		},
		func(in XuankongLiunianInput) (XuankongLiunianInput, error) {
			at, loc, err := parseStandardTime(in.LiuNianStartTime)
			if err != nil {
				return in, fmt.Errorf("xuankong.liunian: %w", err)
			}
			ly, err := tianwen.LichunYearAt(at)
			if err != nil {
				return in, fmt.Errorf("xuankong.liunian: %w", err)
			}
			in.LiuNianStartTime = standardTimeValue(ly.End.In(loc))
			return in, nil
		},
	)
	if err != nil {
		return XuankongLiuNianSeries{}, err
	}
	return XuankongLiuNianSeries{LiuNian: items}, nil
}

func computeXuankongLiunian(input XuankongLiunianInput) (XuankongLiunianOutput, error) {
	at, loc, err := parseStandardTime(input.LiuNianStartTime)
	if err != nil {
		return XuankongLiunianOutput{}, fmt.Errorf("xuankong.liunian: %w", err)
	}
	ly, err := tianwen.LichunYearAt(at)
	if err != nil {
		return XuankongLiunianOutput{}, fmt.Errorf("xuankong.liunian: %w", err)
	}
	chart, err := xuankongComputeChart(input.Zhai)
	if err != nil {
		return XuankongLiunianOutput{}, fmt.Errorf("xuankong.liunian: %w", err)
	}
	result := xuankong.ComputeLiuNian(ly.Year, &chart)
	output := xuankongLiunianWire(result)
	startDate := tianwen.StandardDateAt(ly.Start, loc)
	output.NianJie = XuankongLiuNianJie{
		LiuNianStartTime: standardTimeValue(at),
		LichunYear:       ly.Year,
		LichunStart:      civilDateString(startDate),
		LichunEnd:        civilDateString(tianwen.StandardDateAt(ly.End, loc)),
		BoundaryDay:      tianwen.StandardDateAt(at, loc) == startDate,
	}
	return output, nil
}

func xuankongComputeChart(input XuankongZhaiInput) (xuankong.Chart, error) {
	at, _, err := parseStandardTime(input.YunStartTime)
	if err != nil {
		return xuankong.Chart{}, err
	}
	if at.UTC().Year() < 1864 || at.UTC().Year() > 2200 {
		return xuankong.Chart{}, fmt.Errorf("year %d out of range (1864-2200)", at.UTC().Year())
	}
	if (int(input.ZuoShan)+12)%24 != int(input.XiangShan) {
		return xuankong.Chart{}, fmt.Errorf(
			"xiang_shan must be opposite zuo_shan by 12 mountains, got %d/%d",
			input.ZuoShan, input.XiangShan,
		)
	}
	return xuankong.ComputeChartAtStandard(at, int(input.ZuoShan), int(input.XiangShan))
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

type XuankongLiuNianSeries struct {
	LiuNian []XuankongLiunianOutput `json:"liu_nian" jsonschema:"玄空流年数组，长度 = count。"`
}
