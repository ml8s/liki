package huangli

import "liki-engine/internal/engine/ganzhi"

// dayMansion describes one of the 28 mansions (二十八宿).
type dayMansion struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	DongWu   string `json:"dong_wu"`
	Element  string `json:"wuxing"`
	Group    string `json:"group"`
	GroupIdx int    `json:"group_idx"`
}

// mansionsTable holds all 28 mansions in order.
var mansionsTable = [28]dayMansion{
	{0, "角木蛟", "蛟", "木", "东方青龙", 1},
	{1, "亢金龙", "龙", "金", "东方青龙", 2},
	{2, "氐土貉", "貉", "土", "东方青龙", 3},
	{3, "房日兔", "兔", "日", "东方青龙", 4},
	{4, "心月狐", "狐", "月", "东方青龙", 5},
	{5, "尾火虎", "虎", "火", "东方青龙", 6},
	{6, "箕水豹", "豹", "水", "东方青龙", 7},
	{7, "斗木獬", "獬", "木", "北方玄武", 1},
	{8, "牛金牛", "牛", "金", "北方玄武", 2},
	{9, "女土蝠", "蝠", "土", "北方玄武", 3},
	{10, "虚日鼠", "鼠", "日", "北方玄武", 4},
	{11, "危月燕", "燕", "月", "北方玄武", 5},
	{12, "室火猪", "猪", "火", "北方玄武", 6},
	{13, "壁水貐", "貐", "水", "北方玄武", 7},
	{14, "奎木狼", "狼", "木", "西方白虎", 1},
	{15, "娄金狗", "狗", "金", "西方白虎", 2},
	{16, "胃土雉", "雉", "土", "西方白虎", 3},
	{17, "昴日鸡", "鸡", "日", "西方白虎", 4},
	{18, "毕月乌", "乌", "月", "西方白虎", 5},
	{19, "觜火猴", "猴", "火", "西方白虎", 6},
	{20, "参水猿", "猿", "水", "西方白虎", 7},
	{21, "井木犴", "犴", "木", "南方朱雀", 1},
	{22, "鬼金羊", "羊", "金", "南方朱雀", 2},
	{23, "柳土獐", "獐", "土", "南方朱雀", 3},
	{24, "星日马", "马", "日", "南方朱雀", 4},
	{25, "张月鹿", "鹿", "月", "南方朱雀", 5},
	{26, "翼火蛇", "蛇", "火", "南方朱雀", 6},
	{27, "轸水蚓", "蚓", "水", "南方朱雀", 7},
}

// mansionForDay returns the 28-mansion entry for a given day pillar.
// The cycle: 甲子日 → 虚宿 (index 10), then advances one mansion per day.
func mansionForDay(riZhu ganzhi.Zhu) dayMansion {
	sbIdx := ganzhi.SixtyCycleIndex(riZhu.Gan, riZhu.Zhi)
	mi := (sbIdx + 10) % 28
	return mansionsTable[mi]
}
