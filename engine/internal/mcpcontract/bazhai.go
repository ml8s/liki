package mcpcontract

import (
	"context"

	"liki-engine/internal/engine/bazhai"
	"liki-engine/internal/engine/ganzhi"
)

type BazhaiChartInput struct {
	BirthYear BazhaiBirthYear `json:"birth_year" jsonschema:"出生公历年份。"`
	Gender    GenderName      `json:"gender" jsonschema:"性别。"`
}

type BazhaiGuaOutput struct {
	Index   GuaNumber    `json:"index" jsonschema:"洛书宫数。"`
	Name    EightGuaName `json:"name" jsonschema:"卦名。"`
	Wuxing  ElementName  `json:"wuxing" jsonschema:"卦五行。"`
	YinYang YinYangName  `json:"yin_yang" jsonschema:"卦阴阳。"`
}

type BazhaiMingGuaOutput struct {
	Gua          BazhaiGuaOutput  `json:"gua" jsonschema:"命卦。"`
	Group        LifeGroupName    `json:"group" jsonschema:"东四命或西四命。"`
	YearBoundary YearBoundaryName `json:"year_boundary" jsonschema:"当前实现按公历日历年换运；立春年界属另一流派，不在本盘隐含。"`
}

type BazhaiDirectionsOutput struct {
	ShengQi [1]DirectionName `json:"sheng_qi" jsonschema:"生气方。"`
	TianYi  [1]DirectionName `json:"tian_yi" jsonschema:"天医方。"`
	YanNian [1]DirectionName `json:"yan_nian" jsonschema:"延年方。"`
	FuWei   [1]DirectionName `json:"fu_wei" jsonschema:"伏位方。"`
	HuoHai  [1]DirectionName `json:"huo_hai" jsonschema:"祸害方。"`
	WuGui   [1]DirectionName `json:"wu_gui" jsonschema:"五鬼方。"`
	LiuSha  [1]DirectionName `json:"liu_sha" jsonschema:"六煞方。"`
	JueMing [1]DirectionName `json:"jue_ming" jsonschema:"绝命方。"`
}

type BazhaiAnnualStarOutput struct {
	GongNum  GuaNumber        `json:"gong_num" jsonschema:"洛书宫数。"`
	Xing     FlyingStarNumber `json:"xing" jsonschema:"紫白星数。"`
	XingName FlyingStarName   `json:"xing_name" jsonschema:"紫白星名。"`
	Wuxing   ElementName      `json:"wuxing" jsonschema:"星五行。"`
	RuZhong  bool             `json:"ru_zhong" jsonschema:"是否入中。"`
}

type BazhaiYearStarsOutput struct {
	Year         BazhaiBirthYear           `json:"year" jsonschema:"流年公历年份。"`
	RuZhong      FlyingStarName            `json:"ru_zhong" jsonschema:"入中紫白星。"`
	YearBoundary YearBoundaryName          `json:"year_boundary" jsonschema:"当前实现按公历日历年换运。"`
	Palaces      [9]BazhaiAnnualStarOutput `json:"gong_wei" jsonschema:"九宫紫白星。"`
}

type BazhaiChartOutput struct {
	MingGua    BazhaiMingGuaOutput    `json:"ming_gua" jsonschema:"命卦。"`
	BaZhaiDirs BazhaiDirectionsOutput `json:"ba_zhai_dirs" jsonschema:"四吉四凶方向。"`
	YearStars  BazhaiYearStarsOutput  `json:"liu_nian_xing" jsonschema:"出生年紫白飞星。"`
}

type BazhaiLayoutInput struct {
	MingGua   EightGuaName `json:"ming_gua" jsonschema:"命卦。"`
	DoorGua   EightGuaName `json:"door_gua" jsonschema:"门卦。"`
	MasterGua EightGuaName `json:"master_gua" jsonschema:"主卧卦。"`
	StoveGua  EightGuaName `json:"stove_gua" jsonschema:"灶卦。"`
}

type BazhaiDoorStoveOutput struct {
	Direction DirectionName    `json:"direction" jsonschema:"方位。"`
	GuaName   EightGuaName     `json:"gua_name" jsonschema:"卦名。"`
	Wuxing    ElementName      `json:"wuxing" jsonschema:"卦五行。"`
	YouXing   YouxingName      `json:"youxing" jsonschema:"游年九星。"`
	Rating    BazhaiRatingName `json:"rating" jsonschema:"吉凶等级。"`
	Group     TrigramGroupName `json:"group" jsonschema:"东四卦或西四卦。"`
	Match     BazhaiMatchName  `json:"match" jsonschema:"与命卦同组吉，异组凶。"`
}

type BazhaiLayoutOutput struct {
	Group   LifeGroupName         `json:"group" jsonschema:"东四命或西四命。"`
	MingGua EightGuaName          `json:"ming_gua" jsonschema:"命卦。"`
	Door    BazhaiDoorStoveOutput `json:"door" jsonschema:"门位判断。"`
	Master  BazhaiDoorStoveOutput `json:"master" jsonschema:"主卧判断。"`
	Stove   BazhaiDoorStoveOutput `json:"stove" jsonschema:"灶位判断。"`
}

func BazhaiChart(_ context.Context, input BazhaiChartInput) (BazhaiChartOutput, error) {
	chart := bazhai.ComputeChart(ganzhi.Gender(input.Gender), int(input.BirthYear))
	return bazhaiChartWire(chart), nil
}

func BazhaiLayout(_ context.Context, input BazhaiLayoutInput) (BazhaiLayoutOutput, error) {
	layout, err := bazhai.ComputeLayout(
		string(input.MingGua), string(input.DoorGua),
		string(input.MasterGua), string(input.StoveGua),
	)
	if err != nil {
		return BazhaiLayoutOutput{}, err
	}
	return BazhaiLayoutOutput{
		Group:   LifeGroupName(layout.Group),
		MingGua: EightGuaName(layout.MingGua),
		Door:    bazhaiDoorStoveWire(layout.Door),
		Master:  bazhaiDoorStoveWire(layout.Master),
		Stove:   bazhaiDoorStoveWire(layout.Stove),
	}, nil
}

func bazhaiChartWire(chart bazhai.Chart) BazhaiChartOutput {
	output := BazhaiChartOutput{
		MingGua: BazhaiMingGuaOutput{
			Gua: BazhaiGuaOutput{
				Index:   GuaNumber(chart.MingGua.Gua.Index),
				Name:    EightGuaName(chart.MingGua.Gua.Name),
				Wuxing:  ElementName(chart.MingGua.Gua.Wuxing),
				YinYang: YinYangName(chart.MingGua.Gua.YinYang),
			},
			Group:        LifeGroupName(chart.MingGua.Group),
			YearBoundary: YearBoundaryName(chart.MingGua.YearBoundary),
		},
		BaZhaiDirs: BazhaiDirectionsOutput{
			ShengQi: directionArray(chart.BaZhaiDirs.ShengQi),
			TianYi:  directionArray(chart.BaZhaiDirs.TianYi),
			YanNian: directionArray(chart.BaZhaiDirs.YanNian),
			FuWei:   directionArray(chart.BaZhaiDirs.FuWei),
			HuoHai:  directionArray(chart.BaZhaiDirs.HuoHai),
			WuGui:   directionArray(chart.BaZhaiDirs.WuGui),
			LiuSha:  directionArray(chart.BaZhaiDirs.LiuSha),
			JueMing: directionArray(chart.BaZhaiDirs.JueMing),
		},
		YearStars: BazhaiYearStarsOutput{
			Year:         BazhaiBirthYear(chart.YearStars.Year),
			RuZhong:      FlyingStarName(chart.YearStars.RuZhong),
			YearBoundary: YearBoundaryName(chart.YearStars.YearBoundary),
		},
	}
	for i, star := range chart.YearStars.Palaces {
		output.YearStars.Palaces[i] = BazhaiAnnualStarOutput{
			GongNum:  GuaNumber(star.GongNum),
			Xing:     FlyingStarNumber(star.Xing),
			XingName: FlyingStarName(star.XingName),
			Wuxing:   ElementName(star.Wuxing),
			RuZhong:  star.RuZhong,
		}
	}
	return output
}

func directionArray(values []string) [1]DirectionName {
	var output [1]DirectionName
	if len(values) > 0 {
		output[0] = DirectionName(values[0])
	}
	return output
}

func bazhaiDoorStoveWire(value bazhai.DoorStoveInfo) BazhaiDoorStoveOutput {
	return BazhaiDoorStoveOutput{
		Direction: DirectionName(value.Direction),
		GuaName:   EightGuaName(value.GuaName),
		Wuxing:    ElementName(value.Wuxing),
		YouXing:   YouxingName(value.YouXing),
		Rating:    BazhaiRatingName(value.Rating),
		Group:     TrigramGroupName(value.Group),
		Match:     BazhaiMatchName(value.Match),
	}
}
