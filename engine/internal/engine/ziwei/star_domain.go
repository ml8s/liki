package ziwei

// Star categories are stable domain classifications. They replace the former
// factor-only palace_facts projection.
const (
	StarLeiZhuXing = "zhu_xing"
	StarLeiFuXing  = "fu_xing"
	StarLeiLiuJi   = "liu_ji"
	StarLeiWenXing = "wen_xing"
	StarLeiShaXing = "sha_xing"
)

var (
	mainStars = map[starIndex]bool{
		ZiWei: true, TianJi: true, TaiYang: true, WuQu: true,
		TianTong: true, LianZhen: true, TianFu: true, TaiYin: true,
		TanLang: true, JuMen: true, TianXiang: true, TianLiang: true,
		QiSha: true, PoJun: true,
	}
	auspiciousStars = map[starIndex]bool{
		TianKui: true, TianYue: true, ZuoFu: true,
		YouBi: true, WenChang: true, WenQu: true,
	}
	literaryStars      = map[starIndex]bool{WenChang: true, WenQu: true}
	palaceMaleficStars = map[starIndex]bool{
		QingYang: true, TuoLuo: true, HuoXing: true,
		LingXing: true, DiKong: true, DiJie: true,
	}
)

func starLei(star starIndex, isMajor bool) []string {
	values := make([]string, 0, 4)
	if isMajor {
		values = append(values, StarLeiZhuXing)
	}
	if auspiciousStars[star] {
		values = append(values, StarLeiLiuJi)
	}
	if literaryStars[star] {
		values = append(values, StarLeiWenXing)
	}
	if palaceMaleficStars[star] {
		values = append(values, StarLeiShaXing)
	}
	if len(values) == 0 {
		values = append(values, StarLeiFuXing)
	}
	return values
}

func newStarInfo(star starIndex, isMajor bool) starInfo {
	return starInfo{Star: star, Name: starName(star), IsMajor: isMajor, Lei: starLei(star, isMajor)}
}
