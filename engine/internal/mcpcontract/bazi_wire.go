// Engine wire DTOs for the MCP contract layer.
//
// These types are ordinary source now, not generated artifacts. Captured
// fixtures under testdata/actual protect compatibility with the published
// engine payload; edit them directly and keep TestEngineContractWireFidelity
// green when a shape changes.
package mcpcontract

import "liki-engine/internal/engine/bazi"

// BaziPillarCangGan is the shared hidden-stem projection for all four pillars.
type BaziPillarCangGan struct {
	Main  *GanNameNullable `json:"main" jsonschema:"藏干 中的 main。"`
	Mid   *GanNameNullable `json:"mid" jsonschema:"藏干 中的 mid。"`
	Minor *string          `json:"minor" jsonschema:"藏干 中的 minor。"`
}

// BaziPillarShiShensItem is the shared ten-god projection item for all pillars.
type BaziPillarShiShensItem struct {
	Gan     *GanNameNullable    `json:"gan" jsonschema:"天干。"`
	Name    *GanNameNullable    `json:"name" jsonschema:"名称。"`
	ShiShen *TenGodNameNullable `json:"shi_shen" jsonschema:"十神。"`
	Source  *string             `json:"source" jsonschema:"来源。"`
}

// BaziPillarShenShaItem is the shared shen-sha projection item for all pillars.
type BaziPillarShenShaItem struct {
	Category    *ShenShaCategoryNameNullable `json:"category" jsonschema:"类别。"`
	Description *string                      `json:"description" jsonschema:"说明。"`
	Name        *string                      `json:"name" jsonschema:"名称。"`
}

// ── bazi.chart output ──

type BaziChartOutput struct {
	ZiShiRule     string                       `json:"zi_shi_rule" jsonschema:"子时换日规则。"`
	Nian          BaziChartPillar              `json:"nian" jsonschema:"年柱: gan/zhi/na_yin"`
	Yue           BaziChartPillar              `json:"yue" jsonschema:"月柱: gan/zhi/na_yin"`
	Ri            BaziChartPillar              `json:"ri" jsonschema:"日柱: gan/zhi/na_yin"`
	Shi           BaziChartPillar              `json:"shi" jsonschema:"时柱: gan/zhi/na_yin"`
	DaYun         BaziChartDaYun               `json:"da_yun" jsonschema:"大运。root_refs 说明坐支本气与原局藏干通根。"`
	BirthYear     int                          `json:"birth_year" jsonschema:"出生公历年份"`
	FuYi          BaziChartFuYi                `json:"fu_yi" jsonschema:"扶抑候选：用/喜/忌与身强弱；最终用神由 counsel 裁决"`
	SanYuan       *BaziChartSanYuan            `json:"san_yuan,omitempty" jsonschema:"三元: 胎元/命宫/身宫"`
	TaiXi         *BaziChartTaiXi              `json:"tai_xi,omitempty" jsonschema:"胎息（日干五合配、日支六合配）"`
	GongJia       *[]BaziChartGongJiaItem      `json:"gong_jia,omitempty" jsonschema:"经典拱局：三合拱 / 三会拱"`
	NaYinRel      *[]BaziChartNayinRelItem     `json:"na_yin_rel,omitempty" jsonschema:"纳音生克"`
	ChangSheng    *[]BaziChartChangShengItem   `json:"chang_sheng,omitempty" jsonschema:"日主十二宫全表（长生/沐浴/冠带/临官/帝旺/衰/病/死/墓/绝/胎/养在十二支的分布）"`
	Zodiac        *string                      `json:"zodiac,omitempty" jsonschema:"生肖（年支对应动物，如酉→鸡）"`
	SanQiName     *string                      `json:"san_qi_name,omitempty" jsonschema:"三奇贵人"`
	ShenShaSchool BaziChartShenShaSchool       `json:"shen_sha_school" jsonschema:"shen_sha_school。"`
	LuRoots       []BaziChartLuRootsItem       `json:"lu_roots" jsonschema:"透干十神得十干禄"`
	GuanXi        []BaziChartGuanXiItem        `json:"guan_xi" jsonschema:"canonical 命局关系事实；每条 occurrence 使用稳定 id。"`
	TenGodStates  []BaziChartTenGodStatesItem  `json:"ten_god_states" jsonschema:"engine 十神显隐 / 通根 / 得令 / 旺弱状态"`
	ElementStates []BaziChartElementStatesItem `json:"element_states" jsonschema:"engine 五行季节旺弱、组合旺弱与生克方向"`
	GuanSha       BaziChartGuanSha             `json:"guan_sha" jsonschema:"官杀取清结论。"`
	CaiKu         BaziChartCaiKu               `json:"cai_ku" jsonschema:"财库结论。"`
	TiaoHou       BaziChartTiaoHou             `json:"tiao_hou" jsonschema:"调候候选：主辅调候五行候选 + 季节 + 条件"`
	GeJu          BaziChartGeJu                `json:"ge_ju" jsonschema:"月令结构证据：月令格神 + 顺逆用 + 结构（非用神）"`
	DayXun        string                       `json:"ri_xun" jsonschema:"日柱所在旬"`
	DayXunKong    string                       `json:"ri_xun_kong" jsonschema:"日柱旬空二支；各柱 is_void 均按此判断"`
}

type BaziChartPillar struct {
	Gan            GanName                  `json:"gan" jsonschema:"天干。"`
	Zhi            ZhiName                  `json:"zhi" jsonschema:"地支。"`
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

type BaziChartDaYun struct {
	Steps           []BaziChartDaYunStepsItem `json:"steps" jsonschema:"大运。root_refs 说明坐支本气与原局藏干通根。列表。"`
	Direction       *string                   `json:"direction,omitempty" jsonschema:"大运。大运方向。"`
	StartDate       *string                   `json:"start_date,omitempty" jsonschema:"起始公历日 YYYY-MM-DD。"`
	StartDayAfter   *int                      `json:"start_day_after,omitempty" jsonschema:"大运。出生后余日数。"`
	StartMonthAfter *int                      `json:"start_month_after,omitempty" jsonschema:"大运。出生后余月数。"`
	StartYearAfter  *int                      `json:"start_year_after,omitempty" jsonschema:"大运。出生后整年数。"`
}

type BaziChartDaYunStepsItem struct {
	Gan       GanName                                `json:"gan" jsonschema:"天干。"`
	Zhi       ZhiName                                `json:"zhi" jsonschema:"地支。"`
	Name      string                                 `json:"name" jsonschema:"名称。"`
	Wuxing    ElementName                            `json:"wuxing" jsonschema:"五行。"`
	ShiShen   string                                 `json:"shi_shen" jsonschema:"十神。"`
	Rooted    bool                                   `json:"rooted" jsonschema:"是否通根。"`
	RootRefs  *[]BaziChartDaYunStepsItemRootRefsItem `json:"root_refs,omitempty" jsonschema:"通根依据。"`
	EndDate   *string                                `json:"end_date,omitempty" jsonschema:"结束公历日 YYYY-MM-DD。"`
	EndYear   *int                                   `json:"end_year,omitempty" jsonschema:"结束公历年。"`
	StartDate *string                                `json:"start_date,omitempty" jsonschema:"起始公历日 YYYY-MM-DD。"`
	StartYear *int                                   `json:"start_year,omitempty" jsonschema:"起始公历年。"`
}

type BaziChartDaYunStepsItemRootRefsItem struct {
	Source RootRefsSourceName `json:"source" jsonschema:"来源。"`
	Pillar *PillarName        `json:"pillar,omitempty" jsonschema:"柱位。"`
	Stem   GanName            `json:"stem" jsonschema:"天干。"`
	Branch ZhiName            `json:"branch" jsonschema:"地支。"`
}

type BaziChartFuYi struct {
	WuxingCount map[string]int           `json:"wuxing_count" jsonschema:"五行数量"`
	WangShuai   map[string]WangShuaiName `json:"wang_shuai" jsonschema:"五行旺相休囚死"`
	Yong        CandidateWuxingName      `json:"yong" jsonschema:"扶抑候选用神五行；缺省为 无候选"`
	Xi          CandidateWuxingName      `json:"xi" jsonschema:"扶抑候选喜神五行；缺省为 无候选"`
	Ji          CandidateWuxingName      `json:"ji" jsonschema:"扶抑候选忌神五行；缺省为 无候选"`
	QiangRuo    string                   `json:"qiang_ruo" jsonschema:"身强弱"`
	Pattern     *string                  `json:"pattern,omitempty" jsonschema:"从格标签"`
	ModelID     string                   `json:"fu_yi_model_id" jsonschema:"扶抑候选模型 ID。"`
	Basis       BaziChartFuYiBasis       `json:"basis" jsonschema:"强弱表输入"`
}

type BaziChartFuYiBasis struct {
	RiGanGen      string                  `json:"ri_gan_gen" jsonschema:"日主通根类型。"`
	YinBiShuLiang int                     `json:"yin_bi_shu_liang" jsonschema:"印比数量。"`
	RiGanTongGen  []BaziStemOccurrenceRef `json:"ri_gan_tong_gen" jsonschema:"日主通根位置。"`
	GuanXiLinks   []BaziChartRelationLink `json:"guan_xi_links" jsonschema:"canonical guan_xi 引用与本证据目标。"`
}

type BaziChartRelationLink struct {
	ID      string   `json:"id" jsonschema:"guan_xi occurrence ID。"`
	Targets []string `json:"targets" jsonschema:"证据Projection目标。"`
}

type BaziStemOccurrenceRef struct {
	Branch *ZhiNameNullable `json:"branch" jsonschema:"地支。"`
	Pillar *string          `json:"pillar" jsonschema:"柱位。"`
	Source *string          `json:"source" jsonschema:"来源。"`
	Stem   *GanNameNullable `json:"stem" jsonschema:"天干。"`
}

type BaziChartSanYuan struct {
	TaiYuan  BaziGanZhiPair `json:"tai_yuan" jsonschema:"胎元（月干+1 月支+3）"`
	MingGong BaziGanZhiPair `json:"ming_gong" jsonschema:"命宫"`
	ShenGong BaziGanZhiPair `json:"shen_gong" jsonschema:"身宫"`
}

type BaziGanZhiPair struct {
	Gan GanName `json:"gan" jsonschema:"天干。"`
	Zhi ZhiName `json:"zhi" jsonschema:"地支。"`
}

type BaziChartTaiXi = BaziGanZhiPair

type BaziChartGongJiaItem struct {
	PillarA *int             `json:"pillar_a,omitempty" jsonschema:"pillar_a。"`
	PillarB *int             `json:"pillar_b,omitempty" jsonschema:"pillar_b。"`
	Type    *GongJiaTypeName `json:"type,omitempty" jsonschema:"类型。"`
	Wuxing  *ElementName     `json:"wuxing,omitempty" jsonschema:"五行。"`
	Zhi     *ZhiName         `json:"zhi,omitempty" jsonschema:"地支。"`
}

type BaziChartNayinRelItem struct {
	A        *string `json:"a" jsonschema:"a。"`
	B        *string `json:"b" jsonschema:"b。"`
	Relation *string `json:"relation" jsonschema:"关系。"`
}

type BaziChartChangShengItem struct {
	Index *ZhiNameNullable             `json:"index" jsonschema:"序号。"`
	Name  *ChangShengStageNameNullable `json:"name" jsonschema:"名称。"`
}

type BaziChartShenShaSchool struct {
	DualReference []DualReferenceName     `json:"dual_reference" jsonschema:"dual_reference列表。"`
	Policy        ShenShaSchoolPolicyName `json:"policy" jsonschema:"策略。"`
}

type BaziChartLuRootsItem struct {
	ShiShen string `json:"shi_shen" jsonschema:"十神。"`
	Gan     string `json:"gan" jsonschema:"天干。"`
	Zhi     string `json:"zhi" jsonschema:"地支。"`
}

type BaziChartGuanXiMember struct {
	Gan    GanName    `json:"gan,omitempty" jsonschema:"参与天干。"`
	Pillar PillarName `json:"pillar" jsonschema:"所在柱。"`
	Zhi    ZhiName    `json:"zhi,omitempty" jsonschema:"参与地支。"`
}

type BaziChartGuanXiItem struct {
	Contested bool                    `json:"contested,omitempty" jsonschema:"天干五合是否争合。"`
	ID        string                  `json:"id" jsonschema:"稳定关系 ID：relation_kind:group。"`
	Kind      RelationKindName        `json:"relation_kind" jsonschema:"关系类型。"`
	Members   []BaziChartGuanXiMember `json:"members" jsonschema:"宫位解析后的参与干支。"`
	Group     string                  `json:"group" jsonschema:"稳定匹配组。"`
	Position  string                  `json:"position,omitempty" jsonschema:"天干关系位置。"`
	Wuxing    ElementName             `json:"wuxing,omitempty" jsonschema:"关系化局五行。"`
}

type BaziChartTenGodStatesItem struct {
	ShiShen     TenGodName   `json:"shi_shen" jsonschema:"十神。"`
	Wuxing      ElementName  `json:"wuxing" jsonschema:"五行。"`
	Transparent bool         `json:"transparent" jsonschema:"是否透干。"`
	Hidden      bool         `json:"hidden" jsonschema:"是否藏干。"`
	Rooted      bool         `json:"rooted" jsonschema:"是否通根。"`
	Timely      bool         `json:"timely" jsonschema:"是否得令。"`
	Count       int          `json:"count" jsonschema:"数量。"`
	Strength    StrengthName `json:"strength" jsonschema:"强弱。"`
}

type BaziChartElementStatesItem struct {
	Wuxing             ElementName   `json:"wuxing" jsonschema:"五行。"`
	Season             WangShuaiName `json:"season" jsonschema:"季节状态。"`
	SeasonStrength     StrengthName  `json:"season_strength" jsonschema:"月令强弱。"`
	Strength           StrengthName  `json:"strength" jsonschema:"强弱。"`
	Transparent        bool          `json:"transparent" jsonschema:"是否透干。"`
	Rooted             bool          `json:"rooted" jsonschema:"是否通根。"`
	Controls           ElementName   `json:"controls" jsonschema:"所克五行。"`
	ControlledBy       ElementName   `json:"controlled_by" jsonschema:"所生/所克来源五行。"`
	ControllerStrength StrengthName  `json:"controller_strength" jsonschema:"克神强弱。"`
	Reasons            []string      `json:"reasons" jsonschema:"推导依据列表。"`
}

type BaziChartGuanSha struct {
	QuQing bool `json:"qu_qing" jsonschema:"官杀混杂是否取清。"`
}

type BaziChartCaiKu struct {
	Xian     bool `json:"xian" jsonschema:"日主财星墓库支是否现于命局。"`
	XingRuKu bool `json:"xing_ru_ku" jsonschema:"财星是否入墓库。"`
}

type BaziChartTiaoHou struct {
	PrimaryWuxing      ElementName               `json:"primary_wuxing" jsonschema:"调候候选：主辅调候五行候选 + 季节 + 条件 中的 primary_wuxing。"`
	SecondaryWuxing    SecondaryWuxingName       `json:"secondary_wuxing" jsonschema:"调候候选：主辅调候五行候选 + 季节 + 条件 中的 secondary_wuxing；缺失时为 无辅调候。"`
	Season             string                    `json:"season" jsonschema:"季节状态。"`
	TiaoHouExplanation string                    `json:"tiao_hou_explanation" jsonschema:"调候口径说明。"`
	ModelID            string                    `json:"tiao_hou_model_id" jsonschema:"调候候选模型 ID。"`
	Primary            BaziChartTiaoHouPrimary   `json:"primary" jsonschema:"主调候。"`
	Secondary          BaziChartTiaoHouSecondary `json:"secondary" jsonschema:"辅调候。"`
	SecondaryCondition *string                   `json:"secondary_condition,omitempty" jsonschema:"调候候选：主辅调候五行候选 + 季节 + 条件 中的 secondary_condition。"`
}

type BaziChartTiaoHouPrimary struct {
	Stem        GanName                                  `json:"stem" jsonschema:"天干。"`
	Transparent bool                                     `json:"transparent" jsonschema:"是否透干。"`
	Hidden      bool                                     `json:"hidden" jsonschema:"是否藏干。"`
	Occurrences []BaziChartTiaoHouPrimaryOccurrencesItem `json:"occurrences" jsonschema:"出现位置。"`
	GuanXiLinks []BaziChartRelationLink                  `json:"guan_xi_links" jsonschema:"canonical guan_xi 引用与本证据目标。"`
}

type BaziChartTiaoHouPrimaryOccurrencesItem struct {
	Branch *string          `json:"branch" jsonschema:"地支。"`
	Pillar *string          `json:"pillar" jsonschema:"柱位。"`
	Source *string          `json:"source" jsonschema:"来源。"`
	Stem   *GanNameNullable `json:"stem" jsonschema:"天干。"`
}

type BaziChartTiaoHouSecondary struct {
	Hidden      *bool                                      `json:"hidden" jsonschema:"是否藏干。"`
	Occurrences []BaziChartTiaoHouSecondaryOccurrencesItem `json:"occurrences" jsonschema:"出现位置。"`
	GuanXiLinks []BaziChartRelationLink                    `json:"guan_xi_links" jsonschema:"canonical guan_xi 引用与本证据目标。"`
	Stem        *GanNameNullable                           `json:"stem" jsonschema:"天干。"`
	Transparent *bool                                      `json:"transparent" jsonschema:"是否透干。"`
}

type BaziChartTiaoHouSecondaryOccurrencesItem struct {
	Branch *ZhiNameNullable `json:"branch" jsonschema:"地支。"`
	Pillar *string          `json:"pillar" jsonschema:"柱位。"`
	Source *string          `json:"source" jsonschema:"来源。"`
	Stem   *string          `json:"stem" jsonschema:"天干。"`
}

type BaziChartGeJu struct {
	GeJu             string                 `json:"ge_ju" jsonschema:"月令结构证据：月令格神 + 顺逆用 + 结构（非用神） 中的 ge_ju。"`
	YongFa           string                 `json:"yong_fa" jsonschema:"月令结构证据：月令格神 + 顺逆用 + 结构（非用神） 中的 yong_fa。"`
	PatternGod       *GanName               `json:"pattern_god,omitempty" jsonschema:"月令结构证据：月令格神 + 顺逆用 + 结构（非用神） 中的 pattern_god。"`
	PatternGodTenGod *TenGodName            `json:"pattern_god_ten_god,omitempty" jsonschema:"月令结构证据：月令格神 + 顺逆用 + 结构（非用神） 中的 pattern_god_ten_god。"`
	PatternGodSource string                 `json:"pattern_god_source" jsonschema:"月令结构证据：月令格神 + 顺逆用 + 结构（非用神） 中的 pattern_god_source。"`
	GeShenTou        bool                   `json:"ge_shen_tou" jsonschema:"格神是否透干。"`
	Structure        BaziChartGeJuStructure `json:"structure" jsonschema:"格局结构。"`
}

type BaziChartGeJuStructure struct {
	Pattern     BaziChartGeJuStructurePattern    `json:"pattern" jsonschema:"格神。"`
	Controller  BaziChartGeJuStructureController `json:"controller" jsonschema:"克神。"`
	Generator   BaziChartGeJuStructureGenerator  `json:"generator" jsonschema:"生神。"`
	GuanXiLinks []BaziChartRelationLink          `json:"guan_xi_links" jsonschema:"canonical guan_xi 引用与本证据目标。"`
}

type BaziChartElementEvidence struct {
	Roots              []BaziStemOccurrenceRef `json:"roots" jsonschema:"通根位置。"`
	Season             *string                 `json:"season" jsonschema:"季节状态。"`
	Strength           *string                 `json:"strength" jsonschema:"强弱。"`
	Timely             *bool                   `json:"timely" jsonschema:"是否得令。"`
	TransparentPillars []string                `json:"transparent_pillars" jsonschema:"透干所在柱。"`
	Wuxing             ElementName             `json:"wuxing" jsonschema:"五行。"`
}

type BaziChartGeJuStructureController struct {
	Roots              []BaziStemOccurrenceRef `json:"roots" jsonschema:"通根位置。"`
	Season             *string                 `json:"season" jsonschema:"季节状态。"`
	Strength           *StrengthNameNullable   `json:"strength" jsonschema:"强弱。"`
	Timely             *bool                   `json:"timely" jsonschema:"是否得令。"`
	TransparentPillars []string                `json:"transparent_pillars" jsonschema:"透干所在柱。"`
	Wuxing             ElementName             `json:"wuxing" jsonschema:"五行。"`
}

type BaziChartGeJuStructurePattern = BaziChartElementEvidence

type BaziChartGeJuStructureGenerator = BaziChartElementEvidence

// ── bazi.bond input ──

type BaziBondInput struct {
	A BaziBondSideInput `json:"a" jsonschema:"甲方：已归一化的出生主体。"`
	B BaziBondSideInput `json:"b" jsonschema:"乙方：已归一化的出生主体。"`
}

type BaziBondSideInput struct {
	Subject NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

// ── bazi.bond output ──

type BaziBondOutput struct {
	DayMaster    BaziBondDayMaster    `json:"day_master" jsonschema:"day_master。"`
	SpousePalace BaziBondSpousePalace `json:"spouse_palace" jsonschema:"spouse_palace。"`
	ZhuCross     BaziBondZhuCross     `json:"zhu_cross" jsonschema:"zhu_cross。"`
	ShiShenCross BaziBondShiShenCross `json:"shi_shen_cross" jsonschema:"shi_shen_cross。"`
	WuxingCross  BaziBondWuxingCross  `json:"wuxing_cross" jsonschema:"wuxing_cross。"`
	NaYinCross   BaziBondNaYinCross   `json:"na_yin_cross" jsonschema:"na_yin_cross。"`
	SpouseStar   BaziBondSpouseStar   `json:"spouse_star" jsonschema:"spouse_star。"`
	ShenshaCross BaziBondShenshaCross `json:"shen_sha_cross" jsonschema:"shen_sha_cross。"`
}

type BaziBondDayMaster struct {
	A        BaziBondSideIdentity `json:"a" jsonschema:"甲方日主。"`
	B        BaziBondSideIdentity `json:"b" jsonschema:"乙方日主。"`
	Relation BaziBondRelation     `json:"relation" jsonschema:"日主关系。"`
}

type BaziBondSideIdentity struct {
	Gan    string      `json:"gan" jsonschema:"天干。"`
	Wuxing ElementName `json:"wuxing" jsonschema:"五行。"`
	Gender GenderName  `json:"gender" jsonschema:"性别。"`
}

type BaziBondRelation struct {
	Kind     string `json:"kind" jsonschema:"关系类型。"`
	Relation string `json:"relation" jsonschema:"关系描述。"`
}

type BaziBondSpousePalace struct {
	A        BaziBondZhiSide           `json:"a" jsonschema:"甲方日支。"`
	B        BaziBondZhiSide           `json:"b" jsonschema:"乙方日支。"`
	Relation BaziBondExplainedRelation `json:"relation" jsonschema:"日支关系。"`
}

type BaziBondZhiSide struct {
	Zhi ZhiName `json:"zhi" jsonschema:"日支。"`
}

type BaziBondExplainedRelation struct {
	Kind        string `json:"kind" jsonschema:"关系类型。"`
	Explanation string `json:"explanation" jsonschema:"关系解释。"`
}

type BaziBondZhuCross struct {
	Pairs []BaziBondZhuCrossPairsItem `json:"pairs" jsonschema:"pairs列表。"`
}

type BaziBondZhuCrossPairsItem struct {
	A           BaziBondZhuSide     `json:"a" jsonschema:"甲方柱。"`
	B           BaziBondZhuSide     `json:"b" jsonschema:"乙方柱。"`
	GanRelation BaziBondRelation    `json:"gan_relation" jsonschema:"天干关系。"`
	ZhiRelation BaziBondZhiRelation `json:"zhi_relation" jsonschema:"地支关系。"`
}

type BaziBondZhuSide struct {
	Zhu PillarName `json:"zhu" jsonschema:"柱位。"`
	Gan GanName    `json:"gan" jsonschema:"天干。"`
	Zhi ZhiName    `json:"zhi" jsonschema:"地支。"`
}

type BaziBondZhiRelation struct {
	Kind        string `json:"kind" jsonschema:"关系类型。"`
	Explanation string `json:"explanation" jsonschema:"关系解释。"`
}

type BaziBondShiShenCross struct {
	AToB BaziBondTenGodPillars `json:"a_to_b" jsonschema:"a_to_b。"`
	BToA BaziBondTenGodPillars `json:"b_to_a" jsonschema:"b_to_a。"`
}

type BaziBondTenGodPillars struct {
	NianGan TenGodName `json:"nian_gan" jsonschema:"年干。"`
	YueGan  TenGodName `json:"yue_gan" jsonschema:"月干。"`
	RiGan   TenGodName `json:"ri_gan" jsonschema:"日干。"`
	ShiGan  TenGodName `json:"shi_gan" jsonschema:"时干。"`
}

type BaziBondWuxingCross struct {
	A           BaziBondElementCounts          `json:"a" jsonschema:"a。"`
	B           BaziBondElementCounts          `json:"b" jsonschema:"b。"`
	Combined    BaziBondElementCounts          `json:"combined" jsonschema:"combined。"`
	FuYiInOther BaziBondWuxingCrossFuYiInOther `json:"fu_yi_in_other" jsonschema:"fu_yi_in_other。"`
}

type BaziBondElementCounts struct {
	Mu   int `json:"木" jsonschema:"木。"`
	Huo  int `json:"火" jsonschema:"火。"`
	Tu   int `json:"土" jsonschema:"土。"`
	Jin  int `json:"金" jsonschema:"金。"`
	Shui int `json:"水" jsonschema:"水。"`
}

type BaziBondWuxingCrossFuYiInOther struct {
	A BaziBondFuYiInOtherSide `json:"a" jsonschema:"a。"`
	B BaziBondFuYiInOtherSide `json:"b" jsonschema:"b。"`
}

type BaziBondFuYiInOtherSide struct {
	ModelID     string              `json:"fu_yi_model_id" jsonschema:"扶抑候选模型 ID。"`
	Yong        CandidateWuxingName `json:"yong" jsonschema:"用神；缺省为 无候选。"`
	Xi          CandidateWuxingName `json:"xi" jsonschema:"喜神；缺省为 无候选。"`
	Ji          CandidateWuxingName `json:"ji" jsonschema:"忌神；缺省为 无候选。"`
	YongInOther int                 `json:"yong_in_other" jsonschema:"yong_in_other。"`
	XiInOther   int                 `json:"xi_in_other" jsonschema:"xi_in_other。"`
	JiInOther   int                 `json:"ji_in_other" jsonschema:"ji_in_other。"`
}

type BaziBondNaYinCross struct {
	Pairs        []BaziBondNaYinCrossPairsItem  `json:"pairs" jsonschema:"pairs列表。"`
	WuxingCounts BaziBondNaYinCrossWuxingCounts `json:"wuxing_counts" jsonschema:"wuxing_counts。"`
}

type BaziBondNaYinCrossPairsItem struct {
	A        BaziBondNaYinSide `json:"a" jsonschema:"甲方柱。"`
	B        BaziBondNaYinSide `json:"b" jsonschema:"乙方柱。"`
	Relation string            `json:"relation" jsonschema:"纳音关系。"`
}

type BaziBondNaYinSide struct {
	Zhu   PillarName `json:"zhu" jsonschema:"柱位。"`
	NaYin string     `json:"na_yin" jsonschema:"纳音。"`
}

type BaziBondNaYinCrossWuxingCounts struct {
	A map[string]int `json:"a" jsonschema:"a 映射表。"`
	B map[string]int `json:"b" jsonschema:"b 映射表。"`
}

type BaziBondSpouseStar struct {
	A BaziBondSpouseStarSide `json:"a" jsonschema:"甲方配偶星。"`
	B BaziBondSpouseStarSide `json:"b" jsonschema:"乙方配偶星。"`
}

type BaziBondSpouseStarSide struct {
	Gender    GenderName                  `json:"gender" jsonschema:"性别。"`
	Primary   BaziBondSpouseStarCandidate `json:"primary" jsonschema:"主调候。"`
	Secondary BaziBondSpouseStarCandidate `json:"secondary" jsonschema:"辅调候。"`
}

type BaziBondSpouseStarOccurrence struct {
	Pillar PillarName `json:"pillar" jsonschema:"柱位。"`
	Branch string     `json:"branch" jsonschema:"地支。"`
	Stem   string     `json:"stem" jsonschema:"天干。"`
	Source string     `json:"source" jsonschema:"来源。"`
	IsVoid bool       `json:"is_void" jsonschema:"是否旬空。"`
}

type BaziBondSpouseStarCandidate struct {
	TenGod      string                         `json:"ten_god" jsonschema:"ten_god。"`
	Occurrences []BaziBondSpouseStarOccurrence `json:"occurrences" jsonschema:"出现位置。"`
}

type BaziBondShenshaCross struct {
	TianYi   BaziBondMutualFact `json:"tian_yi" jsonschema:"tian_yi。"`
	Lu       BaziBondMutualFact `json:"lu" jsonschema:"lu。"`
	TaoHua   BaziBondMutualFact `json:"tao_hua" jsonschema:"tao_hua。"`
	YiMa     BaziBondMutualFact `json:"yi_ma" jsonschema:"yi_ma。"`
	KongWang BaziBondMutualFact `json:"kong_wang" jsonschema:"空亡。"`
	KuiGang  BaziBondMutualFact `json:"kui_gang" jsonschema:"kui_gang。"`
	RiDe     BaziBondMutualFact `json:"ri_de" jsonschema:"ri_de。"`
	RiGui    BaziBondMutualFact `json:"ri_gui" jsonschema:"ri_gui。"`
}

type BaziBondMutualFact struct {
	AToB bool `json:"a_to_b" jsonschema:"甲方神煞是否见于乙方。"`
	BToA bool `json:"b_to_a" jsonschema:"乙方神煞是否见于甲方。"`
}

func newBaziBondOutput(value bazi.Bond) BaziBondOutput {
	pairs := make([]BaziBondZhuCrossPairsItem, 0, len(value.ZhuCross.Pairs))
	for _, pair := range value.ZhuCross.Pairs {
		pairs = append(pairs, BaziBondZhuCrossPairsItem{
			A: BaziBondZhuSide{Zhu: PillarName(pair.AZhu), Gan: GanName(pair.AGan), Zhi: ZhiName(pair.AZhi)},
			B: BaziBondZhuSide{Zhu: PillarName(pair.BZhu), Gan: GanName(pair.BGan), Zhi: ZhiName(pair.BZhi)},
			GanRelation: BaziBondRelation{
				Kind: pair.GanRel.Kind, Relation: pair.GanRel.Relation,
			},
			ZhiRelation: BaziBondZhiRelation{
				Kind: pair.ZhiRel.Kind, Explanation: pair.ZhiRel.Detail,
			},
		})
	}
	naYinPairs := make([]BaziBondNaYinCrossPairsItem, 0, len(value.NaYinCross.Pairs))
	for _, pair := range value.NaYinCross.Pairs {
		naYinPairs = append(naYinPairs, BaziBondNaYinCrossPairsItem{
			A:        BaziBondNaYinSide{Zhu: PillarName(pair.AZhu), NaYin: pair.ANaYin},
			B:        BaziBondNaYinSide{Zhu: PillarName(pair.BZhu), NaYin: pair.BNaYin},
			Relation: pair.Relation,
		})
	}
	tenGodPillars := func(source map[string]string) BaziBondTenGodPillars {
		return BaziBondTenGodPillars{
			NianGan: TenGodName(source["nian_gan"]),
			YueGan:  TenGodName(source["yue_gan"]),
			RiGan:   TenGodName(source["ri_gan"]),
			ShiGan:  TenGodName(source["shi_gan"]),
		}
	}
	return BaziBondOutput{
		DayMaster: BaziBondDayMaster{
			A: BaziBondSideIdentity{
				Gan: value.DayMaster.A.Gan, Wuxing: ElementName(value.DayMaster.A.Wuxing), Gender: GenderName(value.DayMaster.A.Gender),
			},
			B: BaziBondSideIdentity{
				Gan: value.DayMaster.B.Gan, Wuxing: ElementName(value.DayMaster.B.Wuxing), Gender: GenderName(value.DayMaster.B.Gender),
			},
			Relation: BaziBondRelation{
				Kind:     value.DayMaster.GanRel.Kind,
				Relation: value.DayMaster.GanRel.Relation,
			},
		},
		SpousePalace: BaziBondSpousePalace{
			A: BaziBondZhiSide{Zhi: ZhiName(value.SpousePalace.AZhi)},
			B: BaziBondZhiSide{Zhi: ZhiName(value.SpousePalace.BZhi)},
			Relation: BaziBondExplainedRelation{
				Kind:        value.SpousePalace.Relation.Kind,
				Explanation: value.SpousePalace.Relation.Detail,
			},
		},
		ZhuCross: BaziBondZhuCross{Pairs: pairs},
		ShiShenCross: BaziBondShiShenCross{
			AToB: tenGodPillars(value.ShiShenCross.AToB),
			BToA: tenGodPillars(value.ShiShenCross.BToA),
		},
		WuxingCross: BaziBondWuxingCross{
			A:        newBaziBondElementCounts(value.ElementCross.A),
			B:        newBaziBondElementCounts(value.ElementCross.B),
			Combined: newBaziBondElementCounts(value.ElementCross.Combined),
			FuYiInOther: BaziBondWuxingCrossFuYiInOther{
				A: BaziBondFuYiInOtherSide{
					ModelID: value.ElementCross.FuYi.A.ModelID, Yong: CandidateWuxingName(value.ElementCross.FuYi.A.Yong),
					Xi: CandidateWuxingName(value.ElementCross.FuYi.A.Xi), Ji: CandidateWuxingName(value.ElementCross.FuYi.A.Ji),
					YongInOther: value.ElementCross.FuYi.A.YongInOther, XiInOther: value.ElementCross.FuYi.A.XiInOther, JiInOther: value.ElementCross.FuYi.A.JiInOther,
				},
				B: BaziBondFuYiInOtherSide{
					ModelID: value.ElementCross.FuYi.B.ModelID, Yong: CandidateWuxingName(value.ElementCross.FuYi.B.Yong),
					Xi: CandidateWuxingName(value.ElementCross.FuYi.B.Xi), Ji: CandidateWuxingName(value.ElementCross.FuYi.B.Ji),
					YongInOther: value.ElementCross.FuYi.B.YongInOther, XiInOther: value.ElementCross.FuYi.B.XiInOther, JiInOther: value.ElementCross.FuYi.B.JiInOther,
				},
			},
		},
		NaYinCross: BaziBondNaYinCross{
			Pairs: naYinPairs,
			WuxingCounts: BaziBondNaYinCrossWuxingCounts{
				A: value.NaYinCross.Elements.A,
				B: value.NaYinCross.Elements.B,
			},
		},
		SpouseStar: BaziBondSpouseStar{
			A: newBaziBondSpouseStarSide(value.SpouseStar.A),
			B: newBaziBondSpouseStarSide(value.SpouseStar.B),
		},
		ShenshaCross: newBaziBondShenshaCross(value.ShenshaCross),
	}
}

func newBaziBondShenshaCross(value bazi.ShenshaCross) BaziBondShenshaCross {
	side := func(mutual bazi.ShenshaMutual) BaziBondMutualFact {
		return BaziBondMutualFact{AToB: mutual.AInB, BToA: mutual.BInA}
	}
	return BaziBondShenshaCross{
		TianYi:   side(value.TianYi),
		Lu:       side(value.Lu),
		TaoHua:   side(value.TaoHua),
		YiMa:     side(value.YiMa),
		KongWang: side(value.KongWang),
		KuiGang:  side(value.KuiGang),
		RiDe:     side(value.RiDe),
		RiGui:    side(value.RiGui),
	}
}

func newBaziBondSpouseStarSide(value bazi.SpouseStarFact) BaziBondSpouseStarSide {
	return BaziBondSpouseStarSide{
		Gender:    GenderName(value.Gender),
		Primary:   newBaziBondSpouseStarCandidate(value.Primary),
		Secondary: newBaziBondSpouseStarCandidate(value.Secondary),
	}
}

func newBaziBondElementCounts(value map[string]int) BaziBondElementCounts {
	return BaziBondElementCounts{
		Mu: value["木"], Huo: value["火"], Tu: value["土"], Jin: value["金"], Shui: value["水"],
	}
}

func newBaziBondSpouseStarCandidate(value bazi.SpouseStarGroup) BaziBondSpouseStarCandidate {
	occurrences := make([]BaziBondSpouseStarOccurrence, 0, len(value.Occurrences))
	for _, occurrence := range value.Occurrences {
		occurrences = append(occurrences, BaziBondSpouseStarOccurrence{
			Pillar: PillarName(occurrence.Pillar), Branch: occurrence.Branch, Stem: occurrence.Stem,
			Source: occurrence.Source, IsVoid: occurrence.IsVoid,
		})
	}
	return BaziBondSpouseStarCandidate{TenGod: value.TenGod, Occurrences: occurrences}
}

// ── bazi.liunian input ──

type BaziLiuNianInput struct {
	LiuNianStartTime StandardTime `json:"liu_nian_start_time" jsonschema:"流年选择起点：local standard time，RFC3339 fixed offset。"`
	LiuNianCount     *SeriesCount `json:"liu_nian_count,omitempty" jsonschema:"连续查询年数。"`
	Subject          NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

// ── bazi.liunian output ──

type BaziLiuNianResolved struct {
	LiuNianStartTime StandardTime `json:"liu_nian_start_time" jsonschema:"流年选择器时区标准时间（无 DST）。"`
	LiuNianStartDate DateString   `json:"liu_nian_start_date" jsonschema:"按当地标准日取的流年选择日期。"`
	LichunYear       int          `json:"lichun_year" jsonschema:"该日期所在的立春年。"`
	LichunStart      DateString   `json:"lichun_start" jsonschema:"立春年起始公历日（含）。"`
	LichunEnd        DateString   `json:"lichun_end" jsonschema:"下一立春日（不含）。"`
	BoundaryDay      bool         `json:"boundary_day" jsonschema:"输入日期是否为立春边界日（归入新年）。"`
}

type BaziLiuNianOutput struct {
	Year              int                                 `json:"year" jsonschema:"年。"`
	ShiShen           TenGodName                          `json:"shi_shen" jsonschema:"十神。"`
	NianGan           *string                             `json:"nian_gan,omitempty" jsonschema:"流年天干"`
	NianZhi           *string                             `json:"nian_zhi,omitempty" jsonschema:"流年地支"`
	NianGanWuxing     *string                             `json:"nian_gan_wuxing,omitempty" jsonschema:"流年天干五行。"`
	LiuNianNaYin      *string                             `json:"liu_nian_na_yin,omitempty" jsonschema:"流年干支纳音。"`
	ShengRi           *int                                `json:"sheng_ri,omitempty" jsonschema:"流年是否生扶日主：0/1。"`
	KeRi              *int                                `json:"ke_ri,omitempty" jsonschema:"流年是否克制日主：0/1。"`
	ShenSha           *[]BaziSeriesShenshaItem            `json:"shen_sha,omitempty" jsonschema:"流年神煞（动态：桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天乙贵人/羊刃——年支+日支双查；值年：病符/丧门/吊客/大耗——命局四柱逢煞支即应）"`
	FuYinFanYin       *[]BaziLiuNianFuyinFanyinItem       `json:"fu_yin_fan_yin,omitempty" jsonschema:"伏吟反吟"`
	DaYunInteractions *[]BaziLiuNianDayunInteractionsItem `json:"da_yun_interactions,omitempty" jsonschema:"查询年份所在大运与流年交互（未起运或大运过完时为空数组）"`
	NatalInteractions *[]BaziLiuNianNatalInteractionsItem `json:"natal_interactions,omitempty" jsonschema:"流年与命局交互"`
	NianJie           BaziLiuNianResolved                 `json:"nian_jie" jsonschema:"流年立春年界。"`
}

type BaziSeriesShenshaItem struct {
	Name        ShenshaName         `json:"name" jsonschema:"名称。"`
	Category    ShenShaCategoryName `json:"category" jsonschema:"类别。"`
	Description string              `json:"description" jsonschema:"说明。"`
}

type BaziLiuNianFuyinFanyinItem struct {
	Explanation *string `json:"relation_explanation" jsonschema:"关系解释。"`
	NatalIndex  *int    `json:"natal_index" jsonschema:"natal_index。"`
	Kind        *string `json:"relation_kind" jsonschema:"类型。"`
}

type BaziLiuNianDayunInteractionsItem struct {
	GanRels     []BaziLiuNianDayunInteractionsItemGanRelsItem `json:"gan_rels" jsonschema:"gan_rels列表。"`
	PillarLabel *string                                       `json:"pillar_label" jsonschema:"pillar_label。"`
	ZhiRels     []BaziLiuNianDayunInteractionsItemZhiRelsItem `json:"zhi_rels" jsonschema:"zhi_rels列表。"`
}

type BaziLiuNianDayunInteractionsItemGanRelsItem struct {
	GanA     *GanNameNullable `json:"gan_a" jsonschema:"甲干。"`
	GanB     *GanNameNullable `json:"gan_b" jsonschema:"乙干。"`
	Relation *string          `json:"relation" jsonschema:"关系。"`
	Kind     *string          `json:"relation_kind" jsonschema:"类型。"`
}

type BaziLiuNianDayunInteractionsItemZhiRelsItem struct {
	Explanation *string          `json:"relation_explanation" jsonschema:"关系解释。"`
	Kind        *string          `json:"relation_kind" jsonschema:"类型。"`
	ZhiA        *ZhiNameNullable `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB        *ZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziLiuNianNatalInteractionsItem struct {
	GanRels     []BaziSeriesGanRelationItem `json:"gan_rels" jsonschema:"gan_rels列表。"`
	PillarLabel *string                     `json:"pillar_label" jsonschema:"pillar_label。"`
	ZhiRels     []BaziSeriesZhiRelationItem `json:"zhi_rels" jsonschema:"zhi_rels列表。"`
}

type BaziSeriesGanRelationItem struct {
	GanA     *string          `json:"gan_a" jsonschema:"甲干。"`
	GanB     *GanNameNullable `json:"gan_b" jsonschema:"乙干。"`
	Relation *string          `json:"relation" jsonschema:"关系。"`
	Kind     *string          `json:"relation_kind" jsonschema:"类型。"`
}

type BaziSeriesZhiRelationItem struct {
	Explanation *string          `json:"relation_explanation" jsonschema:"关系解释。"`
	Kind        *string          `json:"relation_kind" jsonschema:"类型。"`
	ZhiA        *string          `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB        *ZhiNameNullable `json:"zhi_b" jsonschema:"乙支。"`
}

type BaziLiuYueInput struct {
	LiuYueStartTime StandardTime `json:"liu_yue_start_time" jsonschema:"流月选择起点：local standard time，RFC3339 fixed offset。"`
	LiuYueCount     *SeriesCount `json:"liu_yue_count,omitempty" jsonschema:"连续查询月数。"`
	Subject         NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

// ── bazi.liuyue output ──

type BaziLiuYueJieQi struct {
	LiuYueStartTime StandardTime `json:"liu_yue_start_time" jsonschema:"流月选择器时区标准时间（无 DST）。"`
	LiuYueStartDate DateString   `json:"liu_yue_start_date" jsonschema:"按当地标准日取的流月选择日期。"`
	JieqiStart      DateString   `json:"jieqi_start" jsonschema:"节气月起日（节，含）。"`
	JieqiEnd        DateString   `json:"jieqi_end" jsonschema:"下一节（不含）。"`
	BoundaryDay     bool         `json:"boundary_day" jsonschema:"输入日期是否为节边界日（归入新月）。"`
}

type BaziLiuYueOutput struct {
	ShiShen      TenGodName                   `json:"shi_shen" jsonschema:"十神。"`
	YueGan       *string                      `json:"yue_gan,omitempty" jsonschema:"流月天干"`
	YueZhi       *string                      `json:"yue_zhi,omitempty" jsonschema:"流月地支"`
	YueGanWuxing *string                      `json:"yue_gan_wuxing,omitempty" jsonschema:"流月天干五行。"`
	ShengRi      *int                         `json:"sheng_ri,omitempty" jsonschema:"流月是否生扶日主：0/1。"`
	KeRi         *int                         `json:"ke_ri,omitempty" jsonschema:"流月是否克制日主：0/1。"`
	ShenSha      *[]BaziSeriesShenshaItem     `json:"shen_sha,omitempty" jsonschema:"流月神煞（动态：桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天乙贵人/羊刃——年支+日支双查）"`
	GanRels      *[]BaziSeriesGanRelationItem `json:"gan_rels,omitempty" jsonschema:"天干关系"`
	ZhiRels      *[]BaziSeriesZhiRelationItem `json:"zhi_rels,omitempty" jsonschema:"地支关系"`
	JieQi        BaziLiuYueJieQi              `json:"jie_qi" jsonschema:"流月节气月界。"`
}

// ── bazi.liuri input ──

type BaziLiuRiInput struct {
	LiuRiStartTime StandardTime `json:"liu_ri_start_time" jsonschema:"流日选择起点：local standard time，RFC3339 fixed offset。"`
	LiuRiLongitude float64      `json:"liu_ri_longitude" jsonschema:"流日地点经度（-180..180）；真太阳时校正用于日柱。"`
	LiuRiCount     *SeriesCount `json:"liu_ri_count,omitempty" jsonschema:"连续查询日数。"`
	Subject        NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

type BaziLiuRiJie struct {
	LiuRiStartTime StandardTime `json:"liu_ri_start_time" jsonschema:"流日选择器时区标准时间（无 DST）。"`
	LiuRiLocalDate DateString   `json:"liu_ri_local_date" jsonschema:"流日当地标准日。"`
	LiuRiSolarDate DateString   `json:"liu_ri_solar_date" jsonschema:"真太阳时日（日柱所用日期）。"`
	BeijingDate    DateString   `json:"beijing_calendar_date" jsonschema:"北京时间民用日。"`
}

// ── bazi.liuri output ──

type BaziLiuRiOutput struct {
	Date        string                       `json:"liu_ri_date" jsonschema:"流日日期 YYYY-MM-DD。"`
	ShiShen     TenGodName                   `json:"shi_shen" jsonschema:"十神。"`
	RiGan       *string                      `json:"ri_gan,omitempty" jsonschema:"流日天干"`
	RiZhi       *string                      `json:"ri_zhi,omitempty" jsonschema:"流日地支"`
	DayNayin    *string                      `json:"liu_ri_na_yin,omitempty" jsonschema:"liu_ri_na_yin。"`
	Shensha     *[]BaziSeriesShenshaItem     `json:"shen_sha,omitempty" jsonschema:"流日神煞（动态：天乙贵人/文昌/驿马/桃花/华盖——按流日干支查表）"`
	GanRels     *[]BaziSeriesGanRelationItem `json:"gan_rels,omitempty" jsonschema:"天干关系"`
	ZhiRels     *[]BaziSeriesZhiRelationItem `json:"zhi_rels,omitempty" jsonschema:"地支关系"`
	DaYunRels   *[]BaziSeriesZhiRelationItem `json:"da_yun_rels,omitempty" jsonschema:"与大运关系"`
	LiuNianRels *[]BaziLiuRiLiunianRelsItem  `json:"liu_nian_rels,omitempty" jsonschema:"与流年关系"`
	RiJie       BaziLiuRiJie                 `json:"ri_jie" jsonschema:"流日日界。"`
}

type BaziLiuRiLiunianRelsItem struct {
	Explanation *string `json:"relation_explanation" jsonschema:"关系解释。"`
	Kind        *string `json:"relation_kind" jsonschema:"类型。"`
	ZhiA        *string `json:"zhi_a" jsonschema:"甲支。"`
	ZhiB        *string `json:"zhi_b" jsonschema:"乙支。"`
}

// ── bazi.liushi input ──

type BaziLiuShiInput struct {
	LiuShiStartTime StandardTime `json:"liu_shi_start_time" jsonschema:"流时选择起点：local standard time，RFC3339 fixed offset。"`
	LiuShiLongitude float64      `json:"liu_shi_longitude" jsonschema:"流时地点经度（-180..180）；真太阳时校正用于日柱/时柱。"`
	LiuShiCount     *SeriesCount `json:"liu_shi_count,omitempty" jsonschema:"连续查询时辰数。"`
	Subject         NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

// ── bazi.liushi output ──

type BaziLiuShiChen struct {
	LiuShiStartTime StandardTime `json:"liu_shi_start_time" jsonschema:"流时选择器时区标准时间（无 DST）。"`
	LiuShiLocalDate DateString   `json:"liu_shi_local_date" jsonschema:"流时当地标准日。"`
	Shichen         string       `json:"shichen" jsonschema:"真太阳时时辰。"`
}

type BaziLiuShiOutput struct {
	LiuShiTime string                       `json:"liu_shi_time" jsonschema:"流时时点。"`
	ShiShen    TenGodName                   `json:"shi_shen" jsonschema:"十神。"`
	ShiGan     *string                      `json:"shi_gan,omitempty" jsonschema:"流时天干"`
	ShiZhi     *string                      `json:"shi_zhi,omitempty" jsonschema:"流时地支"`
	GanRels    *[]BaziSeriesGanRelationItem `json:"gan_rels,omitempty" jsonschema:"天干关系"`
	ZhiRels    *[]BaziSeriesZhiRelationItem `json:"zhi_rels,omitempty" jsonschema:"地支关系"`
	ShiChen    BaziLiuShiChen               `json:"shi_chen" jsonschema:"流时时辰定位。"`
}

// ── bazi.xiaoyun input ──

type BaziXiaoYunInput struct {
	Subject      NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
	XiaoYunCount *SeriesCount `json:"xiao_yun_count,omitempty" jsonschema:"连续查询小运岁数。"`
}

// ── bazi.xiaoyun output ──

type BaziXiaoYunOutput struct {
	XiaoYun []BaziXiaoYunItem `json:"xiao_yun" jsonschema:"各学派小运列表；长度 = count。"`
}

type BaziXiaoYunItem struct {
	Method string                    `json:"method" jsonschema:"流派标识"`
	Basis  string                    `json:"basis" jsonschema:"起法依据"`
	Zhus   []BaziXiaoYunItemZhusItem `json:"zhus" jsonschema:"zhus列表。"`
}

type BaziXiaoYunItemZhusItem struct {
	Age     int         `json:"age" jsonschema:"age。"`
	Gan     GanName     `json:"gan" jsonschema:"天干。"`
	Zhi     ZhiName     `json:"zhi" jsonschema:"地支。"`
	Name    string      `json:"name" jsonschema:"名称。"`
	ShiShen *TenGodName `json:"shi_shen,omitempty" jsonschema:"十神。"`
}

// ── named natal series roots ──

type BaziLiuNianSeries struct {
	LiuNian []BaziLiuNianOutput `json:"liu_nian" jsonschema:"流年数组，长度 = count。"`
}

type BaziLiuYueSeries struct {
	LiuYue []BaziLiuYueOutput `json:"liu_yue" jsonschema:"流月数组，长度 = count。"`
}

type BaziLiuRiSeries struct {
	LiuRi []BaziLiuRiOutput `json:"liu_ri" jsonschema:"流日数组，长度 = count。"`
}

type BaziLiuShiSeries struct {
	LiuShi []BaziLiuShiOutput `json:"liu_shi" jsonschema:"流时数组，长度 = count。"`
}
