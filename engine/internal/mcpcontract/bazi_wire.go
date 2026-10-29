// Wire DTOs for the migrated bazi/ziwei/qimen engine tools.
//
// The shapes follow the legacy raw-JSON handlers and the captured engine
// payloads under testdata/actual/. The file started as a generated
// projection; it is ordinary source now, so edit it directly and keep
// TestEngineContractWireFidelity green when a shape changes.

package mcpcontract

import "encoding/json"

// ── bazi.fullchart input ──

type BaziFullChartInput struct {
	Chart BaziChart `json:"chart" jsonschema:"bazi.chart 返回的最小命盘"`
}

// BaziPillarCangGan is the shared hidden-stem projection for all four pillars.
type BaziPillarCangGan struct {
	Main  *LiuyaoGanNameNullable `json:"main" jsonschema:"藏干 中的 main。"`
	Mid   *LiuyaoGanNameNullable `json:"mid" jsonschema:"藏干 中的 mid。"`
	Minor *string                `json:"minor" jsonschema:"藏干 中的 minor。"`
}

// BaziPillarShiShensItem is the shared ten-god projection item for all pillars.
type BaziPillarShiShensItem struct {
	Gan     *LiuyaoGanNameNullable `json:"gan" jsonschema:"天干。"`
	Name    *LiuyaoGanNameNullable `json:"name" jsonschema:"名称。"`
	ShiShen *TenGodNameNullable    `json:"shi_shen" jsonschema:"十神。"`
	Source  *string                `json:"source" jsonschema:"来源。"`
}

// BaziPillarShenShaItem is the shared shen-sha projection item for all pillars.
type BaziPillarShenShaItem struct {
	Category    *ShenShaCategoryNameNullable `json:"category" jsonschema:"类别。"`
	Description *string                      `json:"description" jsonschema:"说明。"`
	Name        *string                      `json:"name" jsonschema:"名称。"`
}

// ── bazi.fullchart output ──

type BaziFullChartOutput struct {
	Nian           BaziFullChartNian                 `json:"nian" jsonschema:"年柱: gan/zhi/na_yin"`
	Yue            BaziFullChartYue                  `json:"yue" jsonschema:"月柱: gan/zhi/na_yin"`
	Ri             BaziFullChartRi                   `json:"ri" jsonschema:"日柱: gan/zhi/na_yin"`
	Shi            BaziFullChartShi                  `json:"shi" jsonschema:"时柱: gan/zhi/na_yin"`
	DaYun          BaziFullChartDaYun                `json:"da_yun" jsonschema:"大运。fullchart 会为每步补 rooted/root_refs；root_refs 说明坐支本气与原局藏干通根。"`
	Gender         GenderName                        `json:"gender" jsonschema:"性别。"`
	BirthYear      *int                              `json:"birth_year,omitempty" jsonschema:"出生公历年份"`
	FuYi           BaziFullChartFuYi                 `json:"fu_yi" jsonschema:"扶抑候选：用/喜/忌与身强弱；最终用神由 counsel 裁决"`
	SanYuan        *BaziFullChartSanYuan             `json:"san_yuan,omitempty" jsonschema:"三元: 胎元/命宫/身宫"`
	TaiXi          *BaziFullChartTaiXi               `json:"tai_xi,omitempty" jsonschema:"胎息（日干五合配、日支六合配）"`
	GongJia        *[]BaziFullChartGongJiaItem       `json:"gong_jia,omitempty" jsonschema:"经典拱局：三合拱 / 三会拱"`
	NayinRel       *[]BaziFullChartNayinRelItem      `json:"nayin_rel,omitempty" jsonschema:"纳音生克"`
	ChangSheng     *[]BaziFullChartChangShengItem    `json:"chang_sheng,omitempty" jsonschema:"日主十二宫全表（长生/沐浴/冠带/临官/帝旺/衰/病/死/墓/绝/胎/养在十二支的分布）"`
	Zodiac         *string                           `json:"zodiac,omitempty" jsonschema:"生肖（年支对应动物，如酉→鸡）"`
	SanQiName      *string                           `json:"san_qi_name,omitempty" jsonschema:"三奇贵人"`
	GanHe          *[]BaziFullChartGanHeItem         `json:"gan_he,omitempty" jsonschema:"天干五合"`
	GanChong       *[]BaziFullChartGanChongItem      `json:"gan_chong,omitempty" jsonschema:"天干相冲（甲庚/乙辛/丙壬/丁癸）"`
	ZhiLiuHe       *[]BaziFullChartZhiLiuHeItem      `json:"zhi_liu_he,omitempty" jsonschema:"地支六合"`
	SanHePartial   *[]BaziFullChartSanHePartialItem  `json:"san_he_partial,omitempty" jsonschema:"半合局（三合缺一支且必须含旺支）"`
	SanHe          *[]BaziFullChartSanHeItem         `json:"san_he,omitempty" jsonschema:"三合局"`
	SanHui         *[]BaziFullChartSanHuiItem        `json:"san_hui,omitempty" jsonschema:"三会方"`
	LiuChong       *[]BaziFullChartLiuChongItem      `json:"liu_chong,omitempty" jsonschema:"六冲"`
	LiuHai         *[]BaziFullChartLiuHaiItem        `json:"liu_hai,omitempty" jsonschema:"六害"`
	LiuXing        *[]BaziFullChartLiuXingItem       `json:"liu_xing,omitempty" jsonschema:"相刑"`
	LiuPo          *[]BaziFullChartLiuPoItem         `json:"liu_po,omitempty" jsonschema:"六破"`
	AnHe           *[]BaziFullChartAnHeItem          `json:"an_he,omitempty" jsonschema:"地支暗合"`
	ShenShaSchool  BaziFullChartShenShaSchool        `json:"shen_sha_school" jsonschema:"shen_sha_school。"`
	LuRoots        []BaziFullChartLuRootsItem        `json:"lu_roots" jsonschema:"透干十神得十干禄"`
	RelationGroups []BaziFullChartRelationGroupsItem `json:"relation_groups" jsonschema:"完整合会冲刑组；Python 只读"`
	TenGodStates   []BaziFullChartTenGodStatesItem   `json:"ten_god_states" jsonschema:"engine 十神显隐 / 通根 / 得令 / 旺弱状态"`
	ElementStates  []BaziFullChartElementStatesItem  `json:"element_states" jsonschema:"engine 五行季节旺弱、组合旺弱与生克方向"`
	AtomicFacts    BaziFullChartAtomicFacts          `json:"atomic_facts" jsonschema:"engine 确定性命理原子事实；Python 因子层只读"`
	TiaoHou        BaziFullChartTiaoHou              `json:"tiao_hou" jsonschema:"调候派：主辅调候五行候选 + 季节 + 条件"`
	GeJu           BaziFullChartGeJu                 `json:"ge_ju" jsonschema:"格局派：月令格神 + 顺逆用 + 结构（非用神）"`
	DayXun         string                            `json:"day_xun" jsonschema:"日柱所在旬"`
	DayXunKong     string                            `json:"day_xun_kong" jsonschema:"日柱旬空二支；各柱 is_void 均按此判断"`
}

type BaziFullChartNian struct {
	Gan            LiuyaoGanName            `json:"gan" jsonschema:"天干。"`
	Zhi            LiuyaoZhiName            `json:"zhi" jsonschema:"地支。"`
	NaYin          string                   `json:"na_yin" jsonschema:"干支纳音。"`
	CangGan        BaziPillarCangGan        `json:"cang_gan" jsonschema:"藏干"`
	ShiShens       []BaziPillarShiShensItem `json:"shi_shens" jsonschema:"十神"`
	DayMasterTrend ChangShengStageName      `json:"day_master_trend" jsonschema:"日主在此柱地支的十二长生（地势）"`
	Xun            string                   `json:"xun" jsonschema:"该柱所在旬"`
	XunKong        string                   `json:"xun_kong" jsonschema:"该柱所在旬的旬空二支；is_void 另指该柱值日柱旬空"`
	SelfSitting    ChangShengStageName      `json:"self_sitting" jsonschema:"自坐长生（该柱天干在自己的地支上的十二长生阶段）"`
	ShenSha        []BaziPillarShenShaItem  `json:"shen_sha" jsonschema:"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"`
	IsVoid         bool                     `json:"is_void" jsonschema:"旬空"`
	IsSelfHe       bool                     `json:"is_self_he" jsonschema:"自合"`
	IsKuiGang      bool                     `json:"is_kui_gang" jsonschema:"魁罡"`
	SelfHeName     string                   `json:"self_he_name" jsonschema:"自合名；缺失时为 无自合"`
}

type BaziFullChartYue struct {
	Gan            LiuyaoGanName            `json:"gan" jsonschema:"天干。"`
	Zhi            LiuyaoZhiName            `json:"zhi" jsonschema:"地支。"`
	NaYin          string                   `json:"na_yin" jsonschema:"干支纳音。"`
	CangGan        BaziPillarCangGan        `json:"cang_gan" jsonschema:"藏干"`
	ShiShens       []BaziPillarShiShensItem `json:"shi_shens" jsonschema:"十神"`
	DayMasterTrend ChangShengStageName      `json:"day_master_trend" jsonschema:"日主在此柱地支的十二长生（地势）"`
	Xun            string                   `json:"xun" jsonschema:"该柱所在旬"`
	XunKong        string                   `json:"xun_kong" jsonschema:"该柱所在旬的旬空二支；is_void 另指该柱值日柱旬空"`
	SelfSitting    ChangShengStageName      `json:"self_sitting" jsonschema:"自坐长生（该柱天干在自己的地支上的十二长生阶段）"`
	ShenSha        []BaziPillarShenShaItem  `json:"shen_sha" jsonschema:"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"`
	IsVoid         bool                     `json:"is_void" jsonschema:"旬空"`
	IsSelfHe       bool                     `json:"is_self_he" jsonschema:"自合"`
	IsKuiGang      bool                     `json:"is_kui_gang" jsonschema:"魁罡"`
	SelfHeName     string                   `json:"self_he_name" jsonschema:"自合名；缺失时为 无自合"`
}

type BaziFullChartRi struct {
	Gan            LiuyaoGanName            `json:"gan" jsonschema:"天干。"`
	Zhi            LiuyaoZhiName            `json:"zhi" jsonschema:"地支。"`
	NaYin          string                   `json:"na_yin" jsonschema:"干支纳音。"`
	CangGan        BaziPillarCangGan        `json:"cang_gan" jsonschema:"藏干"`
	ShiShens       []BaziPillarShiShensItem `json:"shi_shens" jsonschema:"十神"`
	DayMasterTrend ChangShengStageName      `json:"day_master_trend" jsonschema:"日主在此柱地支的十二长生（地势）"`
	Xun            string                   `json:"xun" jsonschema:"该柱所在旬"`
	XunKong        string                   `json:"xun_kong" jsonschema:"该柱所在旬的旬空二支；is_void 另指该柱值日柱旬空"`
	SelfSitting    ChangShengStageName      `json:"self_sitting" jsonschema:"自坐长生（该柱天干在自己的地支上的十二长生阶段）"`
	ShenSha        []BaziPillarShenShaItem  `json:"shen_sha" jsonschema:"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"`
	IsVoid         bool                     `json:"is_void" jsonschema:"旬空"`
	IsSelfHe       bool                     `json:"is_self_he" jsonschema:"自合"`
	IsKuiGang      bool                     `json:"is_kui_gang" jsonschema:"魁罡"`
	SelfHeName     string                   `json:"self_he_name" jsonschema:"自合名；缺失时为 无自合"`
}

type BaziFullChartShi struct {
	Gan            LiuyaoGanName            `json:"gan" jsonschema:"天干。"`
	Zhi            LiuyaoZhiName            `json:"zhi" jsonschema:"地支。"`
	NaYin          string                   `json:"na_yin" jsonschema:"干支纳音。"`
	CangGan        BaziPillarCangGan        `json:"cang_gan" jsonschema:"藏干"`
	ShiShens       []BaziPillarShiShensItem `json:"shi_shens" jsonschema:"十神"`
	DayMasterTrend ChangShengStageName      `json:"day_master_trend" jsonschema:"日主在此柱地支的十二长生（地势）"`
	Xun            string                   `json:"xun" jsonschema:"该柱所在旬"`
	XunKong        string                   `json:"xun_kong" jsonschema:"该柱所在旬的旬空二支；is_void 另指该柱值日柱旬空"`
	SelfSitting    ChangShengStageName      `json:"self_sitting" jsonschema:"自坐长生（该柱天干在自己的地支上的十二长生阶段）"`
	ShenSha        []BaziPillarShenShaItem  `json:"shen_sha" jsonschema:"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"`
	IsVoid         bool                     `json:"is_void" jsonschema:"旬空"`
	IsSelfHe       bool                     `json:"is_self_he" jsonschema:"自合"`
	IsKuiGang      bool                     `json:"is_kui_gang" jsonschema:"魁罡"`
	SelfHeName     string                   `json:"self_he_name" jsonschema:"自合名；缺失时为 无自合"`
}

type BaziFullChartDaYun struct {
	Steps           []BaziFullChartDaYunStepsItem `json:"steps" jsonschema:"大运。fullchart 会为每步补 rooted/root_refs；root_refs 说明坐支本气与原局藏干通根。列表。"`
	Direction       *string                       `json:"direction,omitempty" jsonschema:"大运。fullchart 会为每步补 rooted/root_refs；root_refs 说明坐支本气与原局藏干通根。 中的 direction。"`
	StartDate       *string                       `json:"start_date,omitempty" jsonschema:"起始公历日 YYYY-MM-DD。"`
	StartDayAfter   *int                          `json:"start_day_after,omitempty" jsonschema:"大运。fullchart 会为每步补 rooted/root_refs；root_refs 说明坐支本气与原局藏干通根。 中的 start_day_after。"`
	StartMonthAfter *int                          `json:"start_month_after,omitempty" jsonschema:"大运。fullchart 会为每步补 rooted/root_refs；root_refs 说明坐支本气与原局藏干通根。 中的 start_month_after。"`
	StartYearAfter  *int                          `json:"start_year_after,omitempty" jsonschema:"大运。fullchart 会为每步补 rooted/root_refs；root_refs 说明坐支本气与原局藏干通根。 中的 start_year_after。"`
}

type BaziFullChartDaYunStepsItem struct {
	Gan       LiuyaoGanName                              `json:"gan" jsonschema:"天干。"`
	Zhi       LiuyaoZhiName                              `json:"zhi" jsonschema:"地支。"`
	Name      string                                     `json:"name" jsonschema:"名称。"`
	Wuxing    ElementName                                `json:"wuxing" jsonschema:"五行。"`
	ShiShen   string                                     `json:"shi_shen" jsonschema:"十神。"`
	Rooted    bool                                       `json:"rooted" jsonschema:"是否通根。"`
	RootRefs  *[]BaziFullChartDaYunStepsItemRootRefsItem `json:"root_refs,omitempty" jsonschema:"通根依据。"`
	EndDate   *string                                    `json:"end_date,omitempty" jsonschema:"结束公历日 YYYY-MM-DD。"`
	EndYear   *int                                       `json:"end_year,omitempty" jsonschema:"结束公历年。"`
	StartDate *string                                    `json:"start_date,omitempty" jsonschema:"起始公历日 YYYY-MM-DD。"`
	StartYear *int                                       `json:"start_year,omitempty" jsonschema:"起始公历年。"`
}

type BaziFullChartDaYunStepsItemRootRefsItem struct {
	Source RootRefsSourceName `json:"source" jsonschema:"来源。"`
	Pillar *PillarName        `json:"pillar,omitempty" jsonschema:"柱位。"`
	Stem   LiuyaoGanName      `json:"stem" jsonschema:"天干。"`
	Branch LiuyaoZhiName      `json:"branch" jsonschema:"地支。"`
}

type BaziFullChartFuYi struct {
	WuxingCount map[string]int             `json:"wuxing_count" jsonschema:"五行数量"`
	WangShuai   map[string]LiuyaoWangShuai `json:"wang_shuai" jsonschema:"五行旺相休囚死"`
	Yong        CandidateWuxingName        `json:"yong" jsonschema:"扶抑候选用神五行；缺省为 无候选"`
	Xi          CandidateWuxingName        `json:"xi" jsonschema:"扶抑候选喜神五行；缺省为 无候选"`
	Ji          CandidateWuxingName        `json:"ji" jsonschema:"扶抑候选忌神五行；缺省为 无候选"`
	Qiangruo    string                     `json:"qiangruo" jsonschema:"身强弱"`
	Pattern     *string                    `json:"pattern,omitempty" jsonschema:"从格标签"`
	Model       string                     `json:"model" jsonschema:"扶抑候选模型标识。"`
	Basis       BaziFullChartFuYiBasis     `json:"basis" jsonschema:"强弱表输入"`
}

type BaziFullChartFuYiBasis struct {
	RootType       string                                     `json:"root_type" jsonschema:"强弱表输入 中的 root_type。"`
	Season         LiuyaoWangShuai                            `json:"season" jsonschema:"季节状态。"`
	YinBiCount     int                                        `json:"yin_bi_count" jsonschema:"强弱表输入 中的 yin_bi_count。"`
	DayMasterRoots []BaziFullChartFuYiBasisDayMasterRootsItem `json:"day_master_roots" jsonschema:"强弱表输入列表。"`
	RelationFacts  []BaziFullChartFuYiBasisRelationFactsItem  `json:"relation_facts" jsonschema:"关系事实。"`
	DayBranch      *string                                    `json:"day_branch,omitempty" jsonschema:"日支。"`
}

type BaziFullChartFuYiBasisDayMasterRootsItem struct {
	Branch *LiuyaoZhiNameNullable `json:"branch" jsonschema:"地支。"`
	Pillar *string                `json:"pillar" jsonschema:"柱位。"`
	Source *string                `json:"source" jsonschema:"来源。"`
	Stem   *LiuyaoGanNameNullable `json:"stem" jsonschema:"天干。"`
}

type BaziFullChartFuYiBasisRelationFactsItem struct {
	Branches []LiuyaoZhiName `json:"branches" jsonschema:"地支列表。"`
	Field    *string         `json:"field" jsonschema:"字段。"`
	Group    *string         `json:"group" jsonschema:"分组。"`
	Pillars  []string        `json:"pillars" jsonschema:"涉及柱位。"`
	Targets  []string        `json:"targets" jsonschema:"作用对象。"`
}

type BaziFullChartSanYuan struct {
	TaiYuan  BaziFullChartSanYuanTaiYuan  `json:"tai_yuan" jsonschema:"胎元（月干+1 月支+3）"`
	MingGong BaziFullChartSanYuanMingGong `json:"ming_gong" jsonschema:"命宫"`
	ShenGong BaziFullChartSanYuanShenGong `json:"shen_gong" jsonschema:"身宫"`
}

type BaziFullChartSanYuanTaiYuan struct {
	Gan LiuyaoGanName `json:"gan" jsonschema:"天干。"`
	Zhi LiuyaoZhiName `json:"zhi" jsonschema:"地支。"`
}

type BaziFullChartSanYuanMingGong struct {
	Gan LiuyaoGanName `json:"gan" jsonschema:"天干。"`
	Zhi LiuyaoZhiName `json:"zhi" jsonschema:"地支。"`
}

type BaziFullChartSanYuanShenGong struct {
	Gan LiuyaoGanName `json:"gan" jsonschema:"天干。"`
	Zhi LiuyaoZhiName `json:"zhi" jsonschema:"地支。"`
}

type BaziFullChartTaiXi struct {
	Gan LiuyaoGanName `json:"gan" jsonschema:"天干。"`
	Zhi LiuyaoZhiName `json:"zhi" jsonschema:"地支。"`
}

type BaziFullChartGongJiaItem struct {
	PillarA *int             `json:"pillar_a,omitempty" jsonschema:"pillar_a。"`
	PillarB *int             `json:"pillar_b,omitempty" jsonschema:"pillar_b。"`
	Type    *GongJiaTypeName `json:"type,omitempty" jsonschema:"类型。"`
	Wuxing  *ElementName     `json:"wuxing,omitempty" jsonschema:"五行。"`
	Zhi     *LiuyaoZhiName   `json:"zhi,omitempty" jsonschema:"地支。"`
}

type BaziFullChartNayinRelItem struct {
	A        *string `json:"a" jsonschema:"a。"`
	B        *string `json:"b" jsonschema:"b。"`
	Relation *string `json:"relation" jsonschema:"关系。"`
}

type BaziFullChartChangShengItem struct {
	Index *LiuyaoZhiNameNullable       `json:"index" jsonschema:"序号。"`
	Name  *ChangShengStageNameNullable `json:"name" jsonschema:"名称。"`
}

type BaziFullChartGanHeItem struct {
	Contested *bool   `json:"contested" jsonschema:"contested。"`
	GanA      *string `json:"gan_a" jsonschema:"甲干。"`
	GanB      *string `json:"gan_b" jsonschema:"乙干。"`
	HeElement *string `json:"he_element" jsonschema:"he_element。"`
	PillarA   *int    `json:"pillar_a" jsonschema:"pillar_a。"`
	PillarB   *int    `json:"pillar_b" jsonschema:"pillar_b。"`
	Position  *string `json:"position" jsonschema:"position。"`
}

type BaziFullChartGanChongItem struct {
	GanA     LiuyaoGanName        `json:"gan_a" jsonschema:"甲干。"`
	GanB     LiuyaoGanName        `json:"gan_b" jsonschema:"乙干。"`
	PillarA  int                  `json:"pillar_a" jsonschema:"pillar_a。"`
	PillarB  int                  `json:"pillar_b" jsonschema:"pillar_b。"`
	Position GanChongPositionName `json:"position" jsonschema:"枚举值：adjacent、remote、separated。"`
}

type BaziFullChartZhiLiuHeItem struct {
	PillarA *int                   `json:"pillar_a" jsonschema:"pillar_a。"`
	PillarB *int                   `json:"pillar_b" jsonschema:"pillar_b。"`
	Wuxing  *string                `json:"wuxing" jsonschema:"五行。"`
	ZhiA    *LiuyaoZhiNameNullable `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB    *LiuyaoZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziFullChartSanHePartialItem struct {
	Type     SanHePartialTypeName `json:"type" jsonschema:"类型。"`
	Name     string               `json:"name" jsonschema:"名称。"`
	Wuxing   ElementName          `json:"wuxing" jsonschema:"五行。"`
	Branches []LiuyaoZhiName      `json:"branches" jsonschema:"地支列表。"`
	Pillars  []PillarName         `json:"pillars" jsonschema:"涉及柱位。"`
}

type BaziFullChartSanHeItem struct {
	Type     *SanHeTypeName   `json:"type,omitempty" jsonschema:"类型。"`
	Name     *string          `json:"name,omitempty" jsonschema:"名称。"`
	Wuxing   *ElementName     `json:"wuxing,omitempty" jsonschema:"五行。"`
	Branches *[]LiuyaoZhiName `json:"branches,omitempty" jsonschema:"地支列表。"`
	Pillars  *[]PillarName    `json:"pillars,omitempty" jsonschema:"涉及柱位。"`
}

type BaziFullChartSanHuiItem struct {
	Type     SanHuiTypeName  `json:"type" jsonschema:"类型。"`
	Name     string          `json:"name" jsonschema:"名称。"`
	Wuxing   ElementName     `json:"wuxing" jsonschema:"五行。"`
	Branches []LiuyaoZhiName `json:"branches" jsonschema:"地支列表。"`
	Pillars  []PillarName    `json:"pillars" jsonschema:"涉及柱位。"`
}

type BaziFullChartLiuChongItem struct {
	PillarA *int    `json:"pillar_a" jsonschema:"pillar_a。"`
	PillarB *int    `json:"pillar_b" jsonschema:"pillar_b。"`
	ZhiA    *string `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB    *string `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziFullChartLiuHaiItem struct {
	PillarA *int    `json:"pillar_a" jsonschema:"pillar_a。"`
	PillarB *int    `json:"pillar_b" jsonschema:"pillar_b。"`
	ZhiA    *string `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB    *string `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziFullChartLiuXingItem struct {
	PillarA *int                   `json:"pillar_a" jsonschema:"pillar_a。"`
	PillarB *int                   `json:"pillar_b" jsonschema:"pillar_b。"`
	ZhiA    *LiuyaoZhiNameNullable `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB    *LiuyaoZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziFullChartLiuPoItem struct {
	ZhiA    LiuyaoZhiName   `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB    LiuyaoZhiName   `json:"zhi_b" jsonschema:"乙支。"`
	PillarA json.RawMessage `json:"pillar_a" jsonschema:"pillar_a。"`
	PillarB json.RawMessage `json:"pillar_b" jsonschema:"pillar_b。"`
	Element *string         `json:"element,omitempty" jsonschema:"element。"`
}

type BaziFullChartAnHeItem struct {
	ZhiA    LiuyaoZhiName   `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB    LiuyaoZhiName   `json:"zhi_b" jsonschema:"乙支。"`
	PillarA json.RawMessage `json:"pillar_a" jsonschema:"pillar_a。"`
	PillarB json.RawMessage `json:"pillar_b" jsonschema:"pillar_b。"`
	Element *string         `json:"element,omitempty" jsonschema:"element。"`
}

type BaziFullChartShenShaSchool struct {
	DualReference []DualReferenceName     `json:"dual_reference" jsonschema:"dual_reference列表。"`
	Policy        ShenShaSchoolPolicyName `json:"policy" jsonschema:"策略。"`
}

type BaziFullChartLuRootsItem struct {
	ShiShen string `json:"shi_shen" jsonschema:"十神。"`
	Gan     string `json:"gan" jsonschema:"天干。"`
	Zhi     string `json:"zhi" jsonschema:"地支。"`
}

type BaziFullChartRelationGroupsItem struct {
	Field RelationGroupsFieldName `json:"field" jsonschema:"字段。"`
	Group string                  `json:"group" jsonschema:"分组。"`
}

type BaziFullChartTenGodStatesItem struct {
	ShiShen     TenGodName   `json:"shi_shen" jsonschema:"十神。"`
	Wuxing      ElementName  `json:"wuxing" jsonschema:"五行。"`
	Transparent bool         `json:"transparent" jsonschema:"transparent。"`
	Hidden      bool         `json:"hidden" jsonschema:"hidden。"`
	Rooted      bool         `json:"rooted" jsonschema:"是否通根。"`
	Timely      bool         `json:"timely" jsonschema:"是否得令。"`
	Count       int          `json:"count" jsonschema:"数量。"`
	Strength    StrengthName `json:"strength" jsonschema:"强弱。"`
}

type BaziFullChartElementStatesItem struct {
	Wuxing             ElementName     `json:"wuxing" jsonschema:"五行。"`
	Season             LiuyaoWangShuai `json:"season" jsonschema:"季节状态。"`
	SeasonStrength     StrengthName    `json:"season_strength" jsonschema:"枚举值。"`
	Strength           StrengthName    `json:"strength" jsonschema:"强弱。"`
	Transparent        bool            `json:"transparent" jsonschema:"transparent。"`
	Rooted             bool            `json:"rooted" jsonschema:"是否通根。"`
	Controls           ElementName     `json:"controls" jsonschema:"枚举值。"`
	ControlledBy       ElementName     `json:"controlled_by" jsonschema:"枚举值。"`
	ControllerStrength StrengthName    `json:"controller_strength" jsonschema:"枚举值。"`
	Reasons            []string        `json:"reasons" jsonschema:"reasons列表。"`
}

type BaziFullChartAtomicFacts struct {
	DayMasterElement      ElementName                               `json:"day_master_element" jsonschema:"日主五行。"`
	DayMasterStem         LiuyaoGanName                             `json:"day_master_stem" jsonschema:"日主天干。"`
	DayBranch             LiuyaoZhiName                             `json:"day_branch" jsonschema:"日支。"`
	MonthLongevity        ChangShengStageName                       `json:"month_longevity" jsonschema:"月令十二长生。"`
	YearStemTenGod        YearStemTenGodName                        `json:"year_stem_ten_god" jsonschema:"枚举值：、七杀、伤官、偏印、偏财、劫财、正印、正官、正财、比肩、食神。"`
	MonthMainTenGod       YearStemTenGodName                        `json:"month_main_ten_god" jsonschema:"枚举值：、七杀、伤官、偏印、偏财、劫财、正印、正官、正财、比肩、食神。"`
	HourStemTenGod        YearStemTenGodName                        `json:"hour_stem_ten_god" jsonschema:"枚举值：、七杀、伤官、偏印、偏财、劫财、正印、正官、正财、比肩、食神。"`
	PatternGodTransparent bool                                      `json:"pattern_god_transparent" jsonschema:"engine 确定性命理原子事实；Python 因子层只读 中的 pattern_god_transparent。"`
	PillarPunishments     BaziFullChartAtomicFactsPillarPunishments `json:"pillar_punishments" jsonschema:"engine 确定性命理原子事实；Python 因子层只读 中的 pillar_punishments。"`
	OfficerKillingCleaned bool                                      `json:"officer_killing_cleaned" jsonschema:"engine 确定性命理原子事实；Python 因子层只读 中的 officer_killing_cleaned。"`
	WealthTombPresent     bool                                      `json:"wealth_tomb_present" jsonschema:"engine 确定性命理原子事实；Python 因子层只读 中的 wealth_tomb_present。"`
	WealthStarInTomb      bool                                      `json:"wealth_star_in_tomb" jsonschema:"engine 确定性命理原子事实；Python 因子层只读 中的 wealth_star_in_tomb。"`
	SpousePalaceState     SpousePalaceStateName                     `json:"spouse_palace_state" jsonschema:"配偶宫状态。"`
	DayBranchType         DayBranchTypeName                         `json:"day_branch_type" jsonschema:"日支类型。"`
	YearOfficerKilling    bool                                      `json:"year_officer_killing" jsonschema:"engine 确定性命理原子事实；Python 因子层只读 中的 year_officer_killing。"`
}

type BaziFullChartAtomicFactsPillarPunishments struct {
	Nian bool `json:"nian" jsonschema:"nian。"`
	Yue  bool `json:"yue" jsonschema:"yue。"`
	Ri   bool `json:"ri" jsonschema:"ri。"`
	Shi  bool `json:"shi" jsonschema:"shi。"`
}

type BaziFullChartTiaoHou struct {
	PrimaryWuxing      ElementName                   `json:"primary_wuxing" jsonschema:"调候派：主辅调候五行候选 + 季节 + 条件 中的 primary_wuxing。"`
	SecondaryWuxing    SecondaryWuxingName           `json:"secondary_wuxing" jsonschema:"调候派：主辅调候五行候选 + 季节 + 条件 中的 secondary_wuxing；缺失时为 无辅调候。"`
	Season             string                        `json:"season" jsonschema:"季节状态。"`
	Detail             string                        `json:"detail" jsonschema:"细节。"`
	Model              string                        `json:"model" jsonschema:"调候派：主辅调候五行候选 + 季节 + 条件 中的 model。"`
	Primary            BaziFullChartTiaoHouPrimary   `json:"primary" jsonschema:"主调候。"`
	Secondary          BaziFullChartTiaoHouSecondary `json:"secondary" jsonschema:"辅调候。"`
	SecondaryCondition *string                       `json:"secondary_condition,omitempty" jsonschema:"调候派：主辅调候五行候选 + 季节 + 条件 中的 secondary_condition。"`
}

type BaziFullChartTiaoHouPrimary struct {
	Stem          LiuyaoGanName                                  `json:"stem" jsonschema:"天干。"`
	Transparent   bool                                           `json:"transparent" jsonschema:"transparent。"`
	Hidden        bool                                           `json:"hidden" jsonschema:"hidden。"`
	Occurrences   []BaziFullChartTiaoHouPrimaryOccurrencesItem   `json:"occurrences" jsonschema:"出现位置。"`
	RelationFacts []BaziFullChartTiaoHouPrimaryRelationFactsItem `json:"relation_facts" jsonschema:"关系事实。"`
}

type BaziFullChartTiaoHouPrimaryOccurrencesItem struct {
	Branch *string                `json:"branch" jsonschema:"地支。"`
	Pillar *string                `json:"pillar" jsonschema:"柱位。"`
	Source *string                `json:"source" jsonschema:"来源。"`
	Stem   *LiuyaoGanNameNullable `json:"stem" jsonschema:"天干。"`
}

type BaziFullChartTiaoHouPrimaryRelationFactsItem struct {
	Branches []LiuyaoZhiName `json:"branches" jsonschema:"地支列表。"`
	Field    *string         `json:"field" jsonschema:"字段。"`
	Group    *string         `json:"group" jsonschema:"分组。"`
	Pillars  []string        `json:"pillars" jsonschema:"涉及柱位。"`
	Targets  []string        `json:"targets" jsonschema:"作用对象。"`
}

type BaziFullChartTiaoHouSecondary struct {
	Hidden        *bool                                            `json:"hidden" jsonschema:"hidden。"`
	Occurrences   []BaziFullChartTiaoHouSecondaryOccurrencesItem   `json:"occurrences" jsonschema:"出现位置。"`
	RelationFacts []BaziFullChartTiaoHouSecondaryRelationFactsItem `json:"relation_facts" jsonschema:"关系事实。"`
	Stem          *LiuyaoGanNameNullable                           `json:"stem" jsonschema:"天干。"`
	Transparent   *bool                                            `json:"transparent" jsonschema:"transparent。"`
}

type BaziFullChartTiaoHouSecondaryOccurrencesItem struct {
	Branch *LiuyaoZhiNameNullable `json:"branch" jsonschema:"地支。"`
	Pillar *string                `json:"pillar" jsonschema:"柱位。"`
	Source *string                `json:"source" jsonschema:"来源。"`
	Stem   *string                `json:"stem" jsonschema:"天干。"`
}

type BaziFullChartTiaoHouSecondaryRelationFactsItem struct {
	Branches []LiuyaoZhiName `json:"branches" jsonschema:"地支列表。"`
	Field    *string         `json:"field" jsonschema:"字段。"`
	Group    *string         `json:"group" jsonschema:"分组。"`
	Pillars  []string        `json:"pillars" jsonschema:"涉及柱位。"`
	Targets  []string        `json:"targets" jsonschema:"作用对象。"`
}

type BaziFullChartGeJu struct {
	GeJu             string                     `json:"ge_ju" jsonschema:"格局派：月令格神 + 顺逆用 + 结构（非用神） 中的 ge_ju。"`
	YongFa           string                     `json:"yong_fa" jsonschema:"格局派：月令格神 + 顺逆用 + 结构（非用神） 中的 yong_fa。"`
	PatternGod       *LiuyaoGanName             `json:"pattern_god,omitempty" jsonschema:"格局派：月令格神 + 顺逆用 + 结构（非用神） 中的 pattern_god。"`
	PatternGodTenGod *TenGodName                `json:"pattern_god_ten_god,omitempty" jsonschema:"格局派：月令格神 + 顺逆用 + 结构（非用神） 中的 pattern_god_ten_god。"`
	PatternGodSource string                     `json:"pattern_god_source" jsonschema:"格局派：月令格神 + 顺逆用 + 结构（非用神） 中的 pattern_god_source。"`
	Structure        BaziFullChartGeJuStructure `json:"structure" jsonschema:"格局结构。"`
}

type BaziFullChartGeJuStructure struct {
	Pattern       BaziFullChartGeJuStructurePattern             `json:"pattern" jsonschema:"格神。"`
	Controller    BaziFullChartGeJuStructureController          `json:"controller" jsonschema:"克神。"`
	Generator     BaziFullChartGeJuStructureGenerator           `json:"generator" jsonschema:"生神。"`
	RelationFacts []BaziFullChartGeJuStructureRelationFactsItem `json:"relation_facts" jsonschema:"关系事实。"`
}

type BaziFullChartGeJuStructurePattern struct {
	Roots              []BaziFullChartGeJuStructurePatternRootsItem `json:"roots" jsonschema:"通根位置。"`
	Season             *string                                      `json:"season" jsonschema:"季节状态。"`
	Strength           *string                                      `json:"strength" jsonschema:"强弱。"`
	Timely             *bool                                        `json:"timely" jsonschema:"是否得令。"`
	TransparentPillars []string                                     `json:"transparent_pillars" jsonschema:"透干所在柱。"`
	Wuxing             ElementName                                  `json:"wuxing" jsonschema:"五行。"`
}

type BaziFullChartGeJuStructurePatternRootsItem struct {
	Branch *LiuyaoZhiNameNullable `json:"branch" jsonschema:"地支。"`
	Pillar *string                `json:"pillar" jsonschema:"柱位。"`
	Source *string                `json:"source" jsonschema:"来源。"`
	Stem   *LiuyaoGanNameNullable `json:"stem" jsonschema:"天干。"`
}

type BaziFullChartGeJuStructureController struct {
	Roots              []BaziFullChartGeJuStructureControllerRootsItem `json:"roots" jsonschema:"通根位置。"`
	Season             *string                                         `json:"season" jsonschema:"季节状态。"`
	Strength           *StrengthNameNullable                           `json:"strength" jsonschema:"强弱。"`
	Timely             *bool                                           `json:"timely" jsonschema:"是否得令。"`
	TransparentPillars []string                                        `json:"transparent_pillars" jsonschema:"透干所在柱。"`
	Wuxing             ElementName                                     `json:"wuxing" jsonschema:"五行。"`
}

type BaziFullChartGeJuStructureControllerRootsItem struct {
	Branch *LiuyaoZhiNameNullable `json:"branch" jsonschema:"地支。"`
	Pillar *string                `json:"pillar" jsonschema:"柱位。"`
	Source *string                `json:"source" jsonschema:"来源。"`
	Stem   *LiuyaoGanNameNullable `json:"stem" jsonschema:"天干。"`
}

type BaziFullChartGeJuStructureGenerator struct {
	Roots              []BaziFullChartGeJuStructureGeneratorRootsItem `json:"roots" jsonschema:"通根位置。"`
	Season             *string                                        `json:"season" jsonschema:"季节状态。"`
	Strength           *string                                        `json:"strength" jsonschema:"强弱。"`
	Timely             *bool                                          `json:"timely" jsonschema:"是否得令。"`
	TransparentPillars []string                                       `json:"transparent_pillars" jsonschema:"透干所在柱。"`
	Wuxing             ElementName                                    `json:"wuxing" jsonschema:"五行。"`
}

type BaziFullChartGeJuStructureGeneratorRootsItem struct {
	Branch *LiuyaoZhiNameNullable `json:"branch" jsonschema:"地支。"`
	Pillar *string                `json:"pillar" jsonschema:"柱位。"`
	Source *string                `json:"source" jsonschema:"来源。"`
	Stem   *LiuyaoGanNameNullable `json:"stem" jsonschema:"天干。"`
}

type BaziFullChartGeJuStructureRelationFactsItem struct {
	Branches []LiuyaoZhiName `json:"branches" jsonschema:"地支列表。"`
	Field    *string         `json:"field" jsonschema:"字段。"`
	Group    *string         `json:"group" jsonschema:"分组。"`
	Pillars  []string        `json:"pillars" jsonschema:"涉及柱位。"`
	Targets  []string        `json:"targets" jsonschema:"作用对象。"`
}

// ── bazi.bond input ──

type BaziBondInput struct {
	A BaziBondInputA `json:"a" jsonschema:"a。"`
	B BaziBondInputB `json:"b" jsonschema:"b。"`
}

type BaziBondInputA struct {
	Chart BaziChart `json:"chart" jsonschema:"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"`
}

type BaziBondInputB struct {
	Chart BaziChart `json:"chart" jsonschema:"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"`
}

// ── bazi.bond output ──

type BaziBondOutput struct {
	DayMaster    BaziBondDayMaster    `json:"day_master" jsonschema:"day_master。"`
	SpousePalace BaziBondSpousePalace `json:"spouse_palace" jsonschema:"spouse_palace。"`
	ZhuCross     BaziBondZhuCross     `json:"zhu_cross" jsonschema:"zhu_cross。"`
	ShiShenCross BaziBondShiShenCross `json:"shi_shen_cross" jsonschema:"shi_shen_cross。"`
	WuxingCross  BaziBondWuxingCross  `json:"wuxing_cross" jsonschema:"wuxing_cross。"`
	NayinCross   BaziBondNayinCross   `json:"nayin_cross" jsonschema:"nayin_cross。"`
	SpouseStar   BaziBondSpouseStar   `json:"spouse_star" jsonschema:"spouse_star。"`
	ShenshaCross BaziBondShenshaCross `json:"shensha_cross" jsonschema:"shensha_cross。"`
}

type BaziBondDayMaster struct {
	A         BaziBondDayMasterA         `json:"a" jsonschema:"a。"`
	B         BaziBondDayMasterB         `json:"b" jsonschema:"b。"`
	GanGuanXi BaziBondDayMasterGanGuanXi `json:"gan_guan_xi" jsonschema:"gan_guan_xi。"`
}

type BaziBondDayMasterA struct {
	Gan    string      `json:"gan" jsonschema:"天干。"`
	Wuxing ElementName `json:"wuxing" jsonschema:"五行。"`
	Gender GenderName  `json:"gender" jsonschema:"性别。"`
}

type BaziBondDayMasterB struct {
	Gan    string      `json:"gan" jsonschema:"天干。"`
	Wuxing ElementName `json:"wuxing" jsonschema:"五行。"`
	Gender GenderName  `json:"gender" jsonschema:"性别。"`
}

type BaziBondDayMasterGanGuanXi struct {
	GanA     string `json:"gan_a" jsonschema:"甲干。"`
	GanB     string `json:"gan_b" jsonschema:"乙干。"`
	Type     string `json:"type" jsonschema:"类型。"`
	Relation string `json:"relation" jsonschema:"关系。"`
}

type BaziBondSpousePalace struct {
	AZhi     string                       `json:"a_zhi" jsonschema:"a_zhi。"`
	BZhi     string                       `json:"b_zhi" jsonschema:"b_zhi。"`
	Relation BaziBondSpousePalaceRelation `json:"relation" jsonschema:"关系。"`
}

type BaziBondSpousePalaceRelation struct {
	ZhiA   string `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   string `json:"zhi_b" jsonschema:"乙支。"`
	Type   string `json:"type" jsonschema:"类型。"`
	Detail string `json:"detail" jsonschema:"细节。"`
}

type BaziBondZhuCross struct {
	Pairs []BaziBondZhuCrossPairsItem `json:"pairs" jsonschema:"pairs列表。"`
}

type BaziBondZhuCrossPairsItem struct {
	AZhu      PillarName                         `json:"a_zhu" jsonschema:"枚举值：nian、ri、shi、yue。"`
	BZhu      PillarName                         `json:"b_zhu" jsonschema:"枚举值：nian、ri、shi、yue。"`
	AGan      LiuyaoGanName                      `json:"a_gan" jsonschema:"a_gan。"`
	BGan      LiuyaoGanName                      `json:"b_gan" jsonschema:"b_gan。"`
	AZhi      string                             `json:"a_zhi" jsonschema:"a_zhi。"`
	BZhi      LiuyaoZhiName                      `json:"b_zhi" jsonschema:"b_zhi。"`
	GanGuanXi BaziBondZhuCrossPairsItemGanGuanXi `json:"gan_guan_xi" jsonschema:"gan_guan_xi。"`
	ZhiGuanXi BaziBondZhuCrossPairsItemZhiGuanXi `json:"zhi_guan_xi" jsonschema:"zhi_guan_xi。"`
}

type BaziBondZhuCrossPairsItemGanGuanXi struct {
	GanA     LiuyaoGanName `json:"gan_a" jsonschema:"甲干。"`
	GanB     LiuyaoGanName `json:"gan_b" jsonschema:"乙干。"`
	Type     string        `json:"type" jsonschema:"类型。"`
	Relation string        `json:"relation" jsonschema:"关系。"`
}

type BaziBondZhuCrossPairsItemZhiGuanXi struct {
	ZhiA   string        `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   LiuyaoZhiName `json:"zhi_b" jsonschema:"乙支。"`
	Type   string        `json:"type" jsonschema:"类型。"`
	Detail string        `json:"detail" jsonschema:"细节。"`
}

type BaziBondShiShenCross struct {
	AToB BaziBondShiShenCrossAToB `json:"a_to_b" jsonschema:"a_to_b。"`
	BToA BaziBondShiShenCrossBToA `json:"b_to_a" jsonschema:"b_to_a。"`
}

type BaziBondShiShenCrossAToB struct {
	NianGan TenGodName `json:"nian_gan" jsonschema:"年干。"`
	YueGan  TenGodName `json:"yue_gan" jsonschema:"月干。"`
	RiGan   TenGodName `json:"ri_gan" jsonschema:"日干。"`
	ShiGan  TenGodName `json:"shi_gan" jsonschema:"时干。"`
}

type BaziBondShiShenCrossBToA struct {
	NianGan TenGodName `json:"nian_gan" jsonschema:"年干。"`
	YueGan  TenGodName `json:"yue_gan" jsonschema:"月干。"`
	RiGan   TenGodName `json:"ri_gan" jsonschema:"日干。"`
	ShiGan  TenGodName `json:"shi_gan" jsonschema:"时干。"`
}

type BaziBondWuxingCross struct {
	A           BaziBondWuxingCrossA           `json:"a" jsonschema:"a。"`
	B           BaziBondWuxingCrossB           `json:"b" jsonschema:"b。"`
	Combined    BaziBondWuxingCrossCombined    `json:"combined" jsonschema:"combined。"`
	FuYiInOther BaziBondWuxingCrossFuYiInOther `json:"fu_yi_in_other" jsonschema:"fu_yi_in_other。"`
}

type BaziBondWuxingCrossA struct {
	Mu   int `json:"木" jsonschema:"木。"`
	Huo  int `json:"火" jsonschema:"火。"`
	Tu   int `json:"土" jsonschema:"土。"`
	Jin  int `json:"金" jsonschema:"金。"`
	Shui int `json:"水" jsonschema:"水。"`
}

type BaziBondWuxingCrossB struct {
	Mu   int `json:"木" jsonschema:"木。"`
	Huo  int `json:"火" jsonschema:"火。"`
	Tu   int `json:"土" jsonschema:"土。"`
	Jin  int `json:"金" jsonschema:"金。"`
	Shui int `json:"水" jsonschema:"水。"`
}

type BaziBondWuxingCrossCombined struct {
	Mu   int `json:"木" jsonschema:"木。"`
	Huo  int `json:"火" jsonschema:"火。"`
	Tu   int `json:"土" jsonschema:"土。"`
	Jin  int `json:"金" jsonschema:"金。"`
	Shui int `json:"水" jsonschema:"水。"`
}

type BaziBondWuxingCrossFuYiInOther struct {
	A BaziBondWuxingCrossFuYiInOtherA `json:"a" jsonschema:"a。"`
	B BaziBondWuxingCrossFuYiInOtherB `json:"b" jsonschema:"b。"`
}

type BaziBondWuxingCrossFuYiInOtherA struct {
	Model       string              `json:"model" jsonschema:"model。"`
	Yong        CandidateWuxingName `json:"yong" jsonschema:"用神；缺省为 无候选。"`
	Xi          CandidateWuxingName `json:"xi" jsonschema:"喜神；缺省为 无候选。"`
	Ji          CandidateWuxingName `json:"ji" jsonschema:"忌神；缺省为 无候选。"`
	YongInOther int                 `json:"yong_in_other" jsonschema:"yong_in_other。"`
	XiInOther   int                 `json:"xi_in_other" jsonschema:"xi_in_other。"`
	JiInOther   int                 `json:"ji_in_other" jsonschema:"ji_in_other。"`
}

type BaziBondWuxingCrossFuYiInOtherB struct {
	Model       string              `json:"model" jsonschema:"model。"`
	Yong        CandidateWuxingName `json:"yong" jsonschema:"用神；缺省为 无候选。"`
	Xi          CandidateWuxingName `json:"xi" jsonschema:"喜神；缺省为 无候选。"`
	Ji          CandidateWuxingName `json:"ji" jsonschema:"忌神；缺省为 无候选。"`
	YongInOther int                 `json:"yong_in_other" jsonschema:"yong_in_other。"`
	XiInOther   int                 `json:"xi_in_other" jsonschema:"xi_in_other。"`
	JiInOther   int                 `json:"ji_in_other" jsonschema:"ji_in_other。"`
}

type BaziBondNayinCross struct {
	Pairs        []BaziBondNayinCrossPairsItem  `json:"pairs" jsonschema:"pairs列表。"`
	WuxingCounts BaziBondNayinCrossWuxingCounts `json:"wuxing_counts" jsonschema:"wuxing_counts。"`
}

type BaziBondNayinCrossPairsItem struct {
	AZhu     PillarName `json:"a_zhu" jsonschema:"枚举值：nian、ri、shi、yue。"`
	BZhu     PillarName `json:"b_zhu" jsonschema:"枚举值：nian、ri、shi、yue。"`
	ANaYin   string     `json:"a_na_yin" jsonschema:"a_na_yin。"`
	BNaYin   string     `json:"b_na_yin" jsonschema:"b_na_yin。"`
	Relation string     `json:"relation" jsonschema:"关系。"`
}

type BaziBondNayinCrossWuxingCounts struct {
	A map[string]int `json:"a" jsonschema:"a 映射表。"`
	B map[string]int `json:"b" jsonschema:"b 映射表。"`
}

type BaziBondSpouseStar struct {
	A BaziBondSpouseStarA `json:"a" jsonschema:"a。"`
	B BaziBondSpouseStarB `json:"b" jsonschema:"b。"`
}

type BaziBondSpouseStarA struct {
	Gender    GenderName                   `json:"gender" jsonschema:"性别。"`
	Primary   BaziBondSpouseStarAPrimary   `json:"primary" jsonschema:"主调候。"`
	Secondary BaziBondSpouseStarASecondary `json:"secondary" jsonschema:"辅调候。"`
}

type BaziBondSpouseStarAPrimary struct {
	TenGod      string                                      `json:"ten_god" jsonschema:"ten_god。"`
	Occurrences []BaziBondSpouseStarAPrimaryOccurrencesItem `json:"occurrences" jsonschema:"出现位置。"`
}

type BaziBondSpouseStarAPrimaryOccurrencesItem struct {
	Pillar PillarName `json:"pillar" jsonschema:"柱位。"`
	Branch string     `json:"branch" jsonschema:"地支。"`
	Stem   string     `json:"stem" jsonschema:"天干。"`
	Source string     `json:"source" jsonschema:"来源。"`
	IsVoid bool       `json:"is_void" jsonschema:"是否旬空。"`
}

type BaziBondSpouseStarASecondary struct {
	TenGod      string                                        `json:"ten_god" jsonschema:"ten_god。"`
	Occurrences []BaziBondSpouseStarASecondaryOccurrencesItem `json:"occurrences" jsonschema:"出现位置。"`
}

type BaziBondSpouseStarASecondaryOccurrencesItem struct {
	Pillar PillarName `json:"pillar" jsonschema:"柱位。"`
	Branch string     `json:"branch" jsonschema:"地支。"`
	Stem   string     `json:"stem" jsonschema:"天干。"`
	Source string     `json:"source" jsonschema:"来源。"`
	IsVoid bool       `json:"is_void" jsonschema:"是否旬空。"`
}

type BaziBondSpouseStarB struct {
	Gender    GenderName                   `json:"gender" jsonschema:"性别。"`
	Primary   BaziBondSpouseStarBPrimary   `json:"primary" jsonschema:"主调候。"`
	Secondary BaziBondSpouseStarBSecondary `json:"secondary" jsonschema:"辅调候。"`
}

type BaziBondSpouseStarBPrimary struct {
	TenGod      string                                      `json:"ten_god" jsonschema:"ten_god。"`
	Occurrences []BaziBondSpouseStarBPrimaryOccurrencesItem `json:"occurrences" jsonschema:"出现位置。"`
}

type BaziBondSpouseStarBPrimaryOccurrencesItem struct {
	Pillar PillarName `json:"pillar" jsonschema:"柱位。"`
	Branch string     `json:"branch" jsonschema:"地支。"`
	Stem   string     `json:"stem" jsonschema:"天干。"`
	Source string     `json:"source" jsonschema:"来源。"`
	IsVoid bool       `json:"is_void" jsonschema:"是否旬空。"`
}

type BaziBondSpouseStarBSecondary struct {
	TenGod      string                                        `json:"ten_god" jsonschema:"ten_god。"`
	Occurrences []BaziBondSpouseStarBSecondaryOccurrencesItem `json:"occurrences" jsonschema:"出现位置。"`
}

type BaziBondSpouseStarBSecondaryOccurrencesItem struct {
	Pillar PillarName `json:"pillar" jsonschema:"柱位。"`
	Branch string     `json:"branch" jsonschema:"地支。"`
	Stem   string     `json:"stem" jsonschema:"天干。"`
	Source string     `json:"source" jsonschema:"来源。"`
	IsVoid *bool      `json:"is_void,omitempty" jsonschema:"是否旬空。"`
}

type BaziBondShenshaCross struct {
	TianYi   BaziBondShenshaCrossTianYi   `json:"tian_yi" jsonschema:"tian_yi。"`
	Lu       BaziBondShenshaCrossLu       `json:"lu" jsonschema:"lu。"`
	TaoHua   BaziBondShenshaCrossTaoHua   `json:"tao_hua" jsonschema:"tao_hua。"`
	YiMa     BaziBondShenshaCrossYiMa     `json:"yi_ma" jsonschema:"yi_ma。"`
	KongWang BaziBondShenshaCrossKongWang `json:"kong_wang" jsonschema:"空亡。"`
	KuiGang  BaziBondShenshaCrossKuiGang  `json:"kui_gang" jsonschema:"kui_gang。"`
	RiDe     BaziBondShenshaCrossRiDe     `json:"ri_de" jsonschema:"ri_de。"`
	RiGui    BaziBondShenshaCrossRiGui    `json:"ri_gui" jsonschema:"ri_gui。"`
}

type BaziBondShenshaCrossTianYi struct {
	APresentInB bool `json:"a_present_in_b" jsonschema:"a_present_in_b。"`
	BPresentInA bool `json:"b_present_in_a" jsonschema:"b_present_in_a。"`
}

type BaziBondShenshaCrossLu struct {
	APresentInB bool `json:"a_present_in_b" jsonschema:"a_present_in_b。"`
	BPresentInA bool `json:"b_present_in_a" jsonschema:"b_present_in_a。"`
}

type BaziBondShenshaCrossTaoHua struct {
	APresentInB bool `json:"a_present_in_b" jsonschema:"a_present_in_b。"`
	BPresentInA bool `json:"b_present_in_a" jsonschema:"b_present_in_a。"`
}

type BaziBondShenshaCrossYiMa struct {
	APresentInB bool `json:"a_present_in_b" jsonschema:"a_present_in_b。"`
	BPresentInA bool `json:"b_present_in_a" jsonschema:"b_present_in_a。"`
}

type BaziBondShenshaCrossKongWang struct {
	APresentInB bool `json:"a_present_in_b" jsonschema:"a_present_in_b。"`
	BPresentInA bool `json:"b_present_in_a" jsonschema:"b_present_in_a。"`
}

type BaziBondShenshaCrossKuiGang struct {
	APresentInB bool `json:"a_present_in_b" jsonschema:"a_present_in_b。"`
	BPresentInA bool `json:"b_present_in_a" jsonschema:"b_present_in_a。"`
}

type BaziBondShenshaCrossRiDe struct {
	APresentInB bool `json:"a_present_in_b" jsonschema:"a_present_in_b。"`
	BPresentInA bool `json:"b_present_in_a" jsonschema:"b_present_in_a。"`
}

type BaziBondShenshaCrossRiGui struct {
	APresentInB bool `json:"a_present_in_b" jsonschema:"a_present_in_b。"`
	BPresentInA bool `json:"b_present_in_a" jsonschema:"b_present_in_a。"`
}

// ── bazi.liunian input ──

type BaziLiuNianInput struct {
	StandardTime StandardTime `json:"standard_time" jsonschema:"目标时区标准时间（无 DST），RFC3339 带 offset。"`
	Count        SeriesCount  `json:"count,omitempty" jsonschema:"连续查询年数，单条或批量items数组"`
	Chart        BaziChart    `json:"chart" jsonschema:"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"`
}

// ── bazi.liunian output ──

type BaziLiuNianResolved struct {
	SelectorStandardTime StandardTime `json:"selector_standard_time" jsonschema:"选择器时区标准时间（无 DST）。"`
	SelectorDate         DateString   `json:"selector_date" jsonschema:"按当地标准日取的选择器日期。"`
	LichunYear           int          `json:"lichun_year" jsonschema:"该日期所在的立春年。"`
	LichunStart          DateString   `json:"lichun_start" jsonschema:"立春年起始公历日（含）。"`
	LichunEnd            DateString   `json:"lichun_end" jsonschema:"下一立春日（不含）。"`
	BoundaryDay          bool         `json:"boundary_day" jsonschema:"输入日期是否为立春边界日（归入新年）。"`
}

type BaziLiuNianOutput struct {
	Year              int                                 `json:"year" jsonschema:"年。"`
	NianName          string                              `json:"nian_name" jsonschema:"nian_name。"`
	ShiShen           TenGodName                          `json:"shi_shen" jsonschema:"十神。"`
	NianGan           *string                             `json:"nian_gan,omitempty" jsonschema:"流年天干"`
	NianZhi           *string                             `json:"nian_zhi,omitempty" jsonschema:"流年地支"`
	Wuxing            *string                             `json:"wuxing,omitempty" jsonschema:"五行。"`
	NaYin             *string                             `json:"na_yin,omitempty" jsonschema:"干支纳音。"`
	Sheng             *int                                `json:"sheng,omitempty" jsonschema:"生数"`
	Ke                *int                                `json:"ke,omitempty" jsonschema:"克数"`
	Shensha           *[]BaziLiuNianShenshaItem           `json:"shensha,omitempty" jsonschema:"流年神煞（动态：桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天乙贵人/羊刃——年支+日支双查；值年：病符/丧门/吊客/大耗——命局四柱逢煞支即应）"`
	FuyinFanyin       *[]BaziLiuNianFuyinFanyinItem       `json:"fuyin_fanyin,omitempty" jsonschema:"伏吟反吟"`
	DayunInteractions *[]BaziLiuNianDayunInteractionsItem `json:"dayun_interactions,omitempty" jsonschema:"查询年份所在大运与流年交互（chart 需含 birth_year；未起运/大运过完/缺 birth_year 时为空数组）"`
	NatalInteractions *[]BaziLiuNianNatalInteractionsItem `json:"natal_interactions,omitempty" jsonschema:"流年与命局交互"`
	AtomicFacts       BaziLiuNianAtomicFacts              `json:"atomic_facts" jsonschema:"engine 流年原子事实"`
	Resolved          BaziLiuNianResolved                 `json:"resolved" jsonschema:"时间推导结果。"`
}

type BaziLiuNianShenshaItem struct {
	Name        ShenshaName         `json:"name" jsonschema:"名称。"`
	Category    ShenShaCategoryName `json:"category" jsonschema:"类别。"`
	Description string              `json:"description" jsonschema:"说明。"`
}

type BaziLiuNianFuyinFanyinItem struct {
	Detail     *string `json:"detail" jsonschema:"细节。"`
	NatalIndex *int    `json:"natal_index" jsonschema:"natal_index。"`
	Type       *string `json:"type" jsonschema:"类型。"`
}

type BaziLiuNianDayunInteractionsItem struct {
	GanRels     []BaziLiuNianDayunInteractionsItemGanRelsItem `json:"gan_rels" jsonschema:"gan_rels列表。"`
	PillarLabel *string                                       `json:"pillar_label" jsonschema:"pillar_label。"`
	ZhiRels     []BaziLiuNianDayunInteractionsItemZhiRelsItem `json:"zhi_rels" jsonschema:"zhi_rels列表。"`
}

type BaziLiuNianDayunInteractionsItemGanRelsItem struct {
	GanA     *LiuyaoGanNameNullable `json:"gan_a" jsonschema:"甲干。"`
	GanB     *LiuyaoGanNameNullable `json:"gan_b" jsonschema:"乙干。"`
	Relation *string                `json:"relation" jsonschema:"关系。"`
	Type     *string                `json:"type" jsonschema:"类型。"`
}

type BaziLiuNianDayunInteractionsItemZhiRelsItem struct {
	Detail *string                `json:"detail" jsonschema:"细节。"`
	Type   *string                `json:"type" jsonschema:"类型。"`
	ZhiA   *LiuyaoZhiNameNullable `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   *LiuyaoZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziLiuNianNatalInteractionsItem struct {
	GanRels     []BaziLiuNianNatalInteractionsItemGanRelsItem `json:"gan_rels" jsonschema:"gan_rels列表。"`
	PillarLabel *string                                       `json:"pillar_label" jsonschema:"pillar_label。"`
	ZhiRels     []BaziLiuNianNatalInteractionsItemZhiRelsItem `json:"zhi_rels" jsonschema:"zhi_rels列表。"`
}

type BaziLiuNianNatalInteractionsItemGanRelsItem struct {
	GanA     *string                `json:"gan_a" jsonschema:"甲干。"`
	GanB     *LiuyaoGanNameNullable `json:"gan_b" jsonschema:"乙干。"`
	Relation *string                `json:"relation" jsonschema:"关系。"`
	Type     *string                `json:"type" jsonschema:"类型。"`
}

type BaziLiuNianNatalInteractionsItemZhiRelsItem struct {
	Detail *string                `json:"detail" jsonschema:"细节。"`
	Type   *string                `json:"type" jsonschema:"类型。"`
	ZhiA   *string                `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   *LiuyaoZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziLiuNianAtomicFacts struct {
	ControlsTargets           BaziLiuNianAtomicFactsControlsTargets    `json:"controls_targets" jsonschema:"engine 流年原子事实 中的 controls_targets。"`
	ControlsElements          []ElementName                            `json:"controls_elements" jsonschema:"engine 流年原子事实列表。"`
	WealthBreaksSeal          bool                                     `json:"wealth_breaks_seal" jsonschema:"engine 流年原子事实 中的 wealth_breaks_seal。"`
	YearBranchRelations       map[string]YearBranchRelationsName       `json:"year_branch_relations" jsonschema:"engine 流年原子事实 映射表。"`
	YearBranchControlledBy    []ElementName                            `json:"year_branch_controlled_by" jsonschema:"engine 流年原子事实列表。"`
	YearGanControlsDayGan     bool                                     `json:"year_gan_controls_day_gan" jsonschema:"engine 流年原子事实 中的 year_gan_controls_day_gan。"`
	DayGanControlsYearGan     bool                                     `json:"day_gan_controls_year_gan" jsonschema:"engine 流年原子事实 中的 day_gan_controls_year_gan。"`
	GanCombines               bool                                     `json:"gan_combines" jsonschema:"engine 流年原子事实 中的 gan_combines。"`
	DayunGanCombines          bool                                     `json:"dayun_gan_combines" jsonschema:"engine 流年原子事实 中的 dayun_gan_combines。"`
	DayunZhiClashesYear       bool                                     `json:"dayun_zhi_clashes_year" jsonschema:"engine 流年原子事实 中的 dayun_zhi_clashes_year。"`
	DayVoidBranches           []LiuyaoZhiName                          `json:"day_void_branches" jsonschema:"engine 流年原子事实列表。"`
	YearEqualsDayPillar       bool                                     `json:"year_equals_day_pillar" jsonschema:"engine 流年原子事实 中的 year_equals_day_pillar。"`
	DayunEqualsYearPillar     bool                                     `json:"dayun_equals_year_pillar" jsonschema:"engine 流年原子事实 中的 dayun_equals_year_pillar。"`
	YearGanEqualsNatalYearGan bool                                     `json:"year_gan_equals_natal_year_gan" jsonschema:"engine 流年原子事实 中的 year_gan_equals_natal_year_gan。"`
	Combinations              []BaziLiuNianAtomicFactsCombinationsItem `json:"combinations" jsonschema:"engine 流年原子事实列表。"`
}

type BaziLiuNianAtomicFactsControlsTargets struct {
	DayMaster      bool `json:"day_master" jsonschema:"day_master。"`
	WealthStar     bool `json:"wealth_star" jsonschema:"wealth_star。"`
	OfficerKilling bool `json:"officer_killing" jsonschema:"officer_killing。"`
	SealStar       bool `json:"seal_star" jsonschema:"seal_star。"`
	FoodInjury     bool `json:"food_injury" jsonschema:"food_injury。"`
}

type BaziLiuNianAtomicFactsCombinationsItem struct {
	Kind         YearBranchRelationsName `json:"kind" jsonschema:"类别。"`
	Group        string                  `json:"group" jsonschema:"分组。"`
	Branches     []LiuyaoZhiName         `json:"branches" jsonschema:"地支列表。"`
	IncludesYear bool                    `json:"includes_year" jsonschema:"includes_year。"`
}

// ── bazi.liuyue input ──

type BaziLiuYueInput struct {
	StandardTime StandardTime `json:"standard_time" jsonschema:"目标时区标准时间（无 DST），RFC3339 带 offset。"`
	Count        SeriesCount  `json:"count,omitempty" jsonschema:"连续查询月数，单条或批量items数组"`
	Chart        BaziChart    `json:"chart" jsonschema:"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"`
}

// ── bazi.liuyue output ──

type BaziLiuYueResolved struct {
	SelectorStandardTime StandardTime `json:"selector_standard_time" jsonschema:"选择器时区标准时间（无 DST）。"`
	SelectorDate         DateString   `json:"selector_date" jsonschema:"按当地标准日取的选择器日期。"`
	JieqiStart           DateString   `json:"jieqi_start" jsonschema:"节气月起日（节，含）。"`
	JieqiEnd             DateString   `json:"jieqi_end" jsonschema:"下一节（不含）。"`
	BoundaryDay          bool         `json:"boundary_day" jsonschema:"输入日期是否为节边界日（归入新月）。"`
}

type BaziLiuYueOutput struct {
	MonthName string                   `json:"month_name" jsonschema:"节气月名。"`
	ShiShen   TenGodName               `json:"shi_shen" jsonschema:"十神。"`
	YueGan    *string                  `json:"yue_gan,omitempty" jsonschema:"流月天干"`
	YueZhi    *string                  `json:"yue_zhi,omitempty" jsonschema:"流月地支"`
	Wuxing    *string                  `json:"wuxing,omitempty" jsonschema:"五行。"`
	Sheng     *int                     `json:"sheng,omitempty" jsonschema:"sheng。"`
	Ke        *int                     `json:"ke,omitempty" jsonschema:"ke。"`
	Shensha   *[]BaziLiuYueShenshaItem `json:"shensha,omitempty" jsonschema:"流月神煞（动态：桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天乙贵人/羊刃——年支+日支双查）"`
	GanRels   *[]BaziLiuYueGanRelsItem `json:"gan_rels,omitempty" jsonschema:"天干关系"`
	ZhiRels   *[]BaziLiuYueZhiRelsItem `json:"zhi_rels,omitempty" jsonschema:"地支关系"`
	Resolved  BaziLiuYueResolved       `json:"resolved" jsonschema:"时间推导结果。"`
}

type BaziLiuYueShenshaItem struct {
	Name        ShenshaName         `json:"name" jsonschema:"名称。"`
	Category    ShenShaCategoryName `json:"category" jsonschema:"类别。"`
	Description string              `json:"description" jsonschema:"说明。"`
}

type BaziLiuYueGanRelsItem struct {
	GanA     *string                `json:"gan_a" jsonschema:"甲干。"`
	GanB     *LiuyaoGanNameNullable `json:"gan_b" jsonschema:"乙干。"`
	Relation *string                `json:"relation" jsonschema:"关系。"`
	Type     *string                `json:"type" jsonschema:"类型。"`
}

type BaziLiuYueZhiRelsItem struct {
	Detail *string                `json:"detail" jsonschema:"细节。"`
	Type   *string                `json:"type" jsonschema:"类型。"`
	ZhiA   *string                `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   *LiuyaoZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

// ── bazi.liuri input ──

type BaziLiuRiInput struct {
	StandardTime StandardTime `json:"standard_time" jsonschema:"目标时区标准时间（无 DST），RFC3339 带 offset。"`
	Longitude    float64      `json:"longitude" jsonschema:"目标地经度（-180..180）；真太阳时校正用于日柱。"`
	Count        SeriesCount  `json:"count,omitempty" jsonschema:"连续查询日数，单条或批量items数组"`
	Chart        BaziChart    `json:"chart" jsonschema:"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"`
}

type BaziLiuRiResolved struct {
	StandardTime StandardTime `json:"standard_time" jsonschema:"时区标准时间（无 DST）。"`
	LocalDate    DateString   `json:"local_date" jsonschema:"目标地标准日。"`
	LastDate     DateString   `json:"last_date" jsonschema:"真太阳时日（日柱所用日期）。"`
	BeijingDate  DateString   `json:"beijing_date" jsonschema:"北京时间民用日。"`
}

// ── bazi.liuri output ──

type BaziLiuRiOutput struct {
	Date        string                      `json:"date" jsonschema:"date。"`
	DayName     string                      `json:"day_name" jsonschema:"day_name。"`
	ShiShen     TenGodName                  `json:"shi_shen" jsonschema:"十神。"`
	RiGan       *string                     `json:"ri_gan,omitempty" jsonschema:"流日天干"`
	RiZhi       *string                     `json:"ri_zhi,omitempty" jsonschema:"流日地支"`
	DayNayin    *string                     `json:"day_nayin,omitempty" jsonschema:"day_nayin。"`
	Shensha     *[]BaziLiuRiShenshaItem     `json:"shensha,omitempty" jsonschema:"流日神煞（动态：天乙贵人/文昌/驿马/桃花/华盖——按流日干支查表）"`
	GanRels     *[]BaziLiuRiGanRelsItem     `json:"gan_rels,omitempty" jsonschema:"天干关系"`
	ZhiRels     *[]BaziLiuRiZhiRelsItem     `json:"zhi_rels,omitempty" jsonschema:"地支关系"`
	DayunRels   *[]BaziLiuRiDayunRelsItem   `json:"dayun_rels,omitempty" jsonschema:"与大运关系"`
	LiunianRels *[]BaziLiuRiLiunianRelsItem `json:"liunian_rels,omitempty" jsonschema:"与流年关系"`
	Resolved    BaziLiuRiResolved           `json:"resolved" jsonschema:"时间推导结果。"`
}

type BaziLiuRiShenshaItem struct {
	Name        ShenshaName         `json:"name" jsonschema:"名称。"`
	Category    ShenShaCategoryName `json:"category" jsonschema:"类别。"`
	Description string              `json:"description" jsonschema:"说明。"`
}

type BaziLiuRiGanRelsItem struct {
	GanA     *string                `json:"gan_a" jsonschema:"甲干。"`
	GanB     *LiuyaoGanNameNullable `json:"gan_b" jsonschema:"乙干。"`
	Relation *string                `json:"relation" jsonschema:"关系。"`
	Type     *string                `json:"type" jsonschema:"类型。"`
}

type BaziLiuRiZhiRelsItem struct {
	Detail *string                `json:"detail" jsonschema:"细节。"`
	Type   *string                `json:"type" jsonschema:"类型。"`
	ZhiA   *string                `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   *LiuyaoZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziLiuRiDayunRelsItem struct {
	Detail *string                `json:"detail" jsonschema:"细节。"`
	Type   *string                `json:"type" jsonschema:"类型。"`
	ZhiA   *string                `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   *LiuyaoZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziLiuRiLiunianRelsItem struct {
	Detail *string `json:"detail" jsonschema:"细节。"`
	Type   *string `json:"type" jsonschema:"类型。"`
	ZhiA   *string `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   *string `json:"zhi_b" jsonschema:"乙支。"`
}

// ── bazi.liushi input ──

type BaziLiuShiInput struct {
	StandardTime StandardTime `json:"standard_time" jsonschema:"占问地标准时间（无 DST），RFC3339 带 offset。"`
	Longitude    float64      `json:"longitude" jsonschema:"占问地经度（-180..180）；真太阳时校正用于日柱/时柱。"`
	Count        SeriesCount  `json:"count,omitempty" jsonschema:"连续查询时辰数，单条或批量items数组"`
	Chart        BaziChart    `json:"chart" jsonschema:"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"`
}

// ── bazi.liushi output ──

type BaziLiuShiResolved struct {
	StandardTime StandardTime `json:"standard_time" jsonschema:"时区标准时间（无 DST）。"`
	LocalDate    DateString   `json:"local_date" jsonschema:"占问地标准日。"`
	Shichen      string       `json:"shichen" jsonschema:"真太阳时时辰。"`
}

type BaziLiuShiOutput struct {
	Time     string                   `json:"time" jsonschema:"time。"`
	HourName string                   `json:"hour_name" jsonschema:"hour_name。"`
	ShiShen  TenGodName               `json:"shi_shen" jsonschema:"十神。"`
	ShiGan   *string                  `json:"shi_gan,omitempty" jsonschema:"流时天干"`
	ShiZhi   *string                  `json:"shi_zhi,omitempty" jsonschema:"流时地支"`
	GanRels  *[]BaziLiuShiGanRelsItem `json:"gan_rels,omitempty" jsonschema:"天干关系"`
	ZhiRels  *[]BaziLiuShiZhiRelsItem `json:"zhi_rels,omitempty" jsonschema:"地支关系"`
	Resolved BaziLiuShiResolved       `json:"resolved" jsonschema:"时间推导结果。"`
}

type BaziLiuShiGanRelsItem struct {
	GanA     *string                `json:"gan_a" jsonschema:"甲干。"`
	GanB     *LiuyaoGanNameNullable `json:"gan_b" jsonschema:"乙干。"`
	Relation *string                `json:"relation" jsonschema:"关系。"`
	Type     *string                `json:"type" jsonschema:"类型。"`
}

type BaziLiuShiZhiRelsItem struct {
	Detail *string                `json:"detail" jsonschema:"细节。"`
	Type   *string                `json:"type" jsonschema:"类型。"`
	ZhiA   *string                `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB   *LiuyaoZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

// ── bazi.xiaoyun input ──

type BaziXiaoYunInput struct {
	Chart BaziChart   `json:"chart" jsonschema:"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"`
	Count SeriesCount `json:"count,omitempty" jsonschema:"返回岁数，单岁或多岁"`
}

// ── bazi.xiaoyun output ──

type BaziXiaoYunOutput struct {
	Items []BaziXiaoYunItem `json:"items" jsonschema:"各学派小运列表。"`
}

type BaziXiaoYunItem struct {
	Method string                    `json:"method" jsonschema:"流派标识"`
	Basis  string                    `json:"basis" jsonschema:"起法依据"`
	Zhus   []BaziXiaoYunItemZhusItem `json:"zhus" jsonschema:"zhus列表。"`
}

type BaziXiaoYunItemZhusItem struct {
	Age     int           `json:"age" jsonschema:"age。"`
	Gan     LiuyaoGanName `json:"gan" jsonschema:"天干。"`
	Zhi     LiuyaoZhiName `json:"zhi" jsonschema:"地支。"`
	Name    string        `json:"name" jsonschema:"名称。"`
	ShiShen *TenGodName   `json:"shi_shen,omitempty" jsonschema:"十神。"`
}
