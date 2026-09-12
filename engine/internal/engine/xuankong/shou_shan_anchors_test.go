package xuankong

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// 收山出煞 理气锚点（《沈氏玄空学》正神正位装/拨水入零堂）。
// 注意：权威完整"出煞"需峦头砂水，纯排盘给出理气部分（收山=坐宫山星当令、
// 拨水入零堂=向宫向星零神），本测试锁定该理气判定与正零神表。
func TestShouShanChuSha_LiQi_Anchors(t *testing.T) {
	cases := []struct {
		name           string
		year           int
		sit, face      int
		wantZheng      int
		wantLing       int
		wantShouShanOK bool
		wantChuShaOK   bool
	}{
		// 七运酉山卯向（旺山旺向）：正神7/零神3；坐宫山星7=收山✓，向宫向星7≠3
		{"七运酉山卯向(旺山旺向)", 1995, 18, 6, 7, 3, true, false},
		// 八运子山午向（双星会向）：正神8/零神2；坐宫山星9、向宫向星8
		{"八运子山午向(双星会向)", 2010, 0, 12, 8, 2, false, false},
		// 五运前十年（寄坤）：正神2/零神8
		{"五运前十年(寄坤)", 1945, 0, 12, 2, 8, false, false},
		// 五运后十年（寄艮）：正神8/零神2
		{"五运后十年(寄艮)", 1960, 0, 12, 8, 2, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(
				time.Date(c.year, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
				116.4, 8)
			chart := ComputeChart(st, c.sit, c.face)
			ss := chart.ShouShanChuSha
			if ss.ZhengShen != c.wantZheng || ss.LingShen != c.wantLing {
				t.Errorf("正零神 = (%d,%d), want (%d,%d)（三元九运表）",
					ss.ZhengShen, ss.LingShen, c.wantZheng, c.wantLing)
			}
			if ss.ShouShanOK != c.wantShouShanOK || ss.ChuShaOK != c.wantChuShaOK {
				t.Errorf("收山=%v 拨水入零堂=%v, want (%v,%v)",
					ss.ShouShanOK, ss.ChuShaOK, c.wantShouShanOK, c.wantChuShaOK)
			}
			if ss.Assessment == "" {
				t.Error("assessment 为空")
			}
		})
	}
}

func TestXingJiaHuiUnknownCombinationIsUnlisted(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, 1, 13)
	encoded, err := json.Marshal(chart.XingJiaHui)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(encoded) == "" || json.Unmarshal(encoded, &[]map[string]any{}) != nil {
		t.Fatalf("invalid xing_jia_hui JSON: %s", encoded)
	}
	if strings.Contains(string(encoded), `"classification"`) || strings.Contains(string(encoded), `"auspicious"`) {
		t.Fatal("xing_jia_hui must not expose fixed auspiciousness")
	}
}

// 双星加会《玄空秘旨》权威星组内容锚点（防止表内容被误改）。
func TestXingJiaHui_Content_Anchors(t *testing.T) {
	type want struct {
		name string
	}
	cases := map[[2]int]want{
		{1, 4}: {"一四同宫"},
		{5, 7}: {"五七加会"},
		{7, 5}: {"七五加会"},
		{3, 9}: {"三九同宫"},
		{9, 3}: {"九三同宫"},
		{2, 5}: {"二五交加"},
		{5, 9}: {"五九交加"},
		{6, 9}: {"六九同宫"},
		{1, 6}: {"一六共宗"},
	}
	for key, w := range cases {
		got, ok := xingJiaHuiTable[key]
		if !ok {
			t.Errorf("xingJiaHuiTable 缺 [%d,%d]（%s）", key[0], key[1], w.name)
			continue
		}
		if got.Name != w.name {
			t.Errorf("[%d,%d] = %s, want %s", key[0], key[1], got.Name, w.name)
		}
	}
}
