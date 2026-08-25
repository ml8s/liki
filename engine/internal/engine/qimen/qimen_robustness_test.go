package qimen

import (
	"math/rand"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func TestRandomDates(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 500; i++ {
		year := 1900 + rng.Intn(200)
		month := 1 + rng.Intn(12)
		day := 1 + rng.Intn(28)
		hour := rng.Intn(24)
		bt := time.Date(year, time.Month(month), day, hour, rng.Intn(60), 0, 0, time.FixedZone("CST", 8*3600))
		st := tianwen.GregorianToSolar(bt, 116.4, 8)
		ch := ComputeChart(st, ShiQiMen)
		if ch.Pan.Jushu < 1 || ch.Pan.Jushu > 9 {
			t.Fatalf("随机 %d-%d-%d %d时: 局数越界 %d", year, month, day, hour, ch.Pan.Jushu)
		}
		if ch.Pan.RiGan == 0 {
			t.Fatalf("随机 %d-%d-%d: 日干缺失", year, month, day)
		}
		// 随机用神符号
		syms := []string{"生门", "开门", "天辅", "天芮", "六合", "值符", "戊", "庚", "乙", "天禽"}
		s, _ := ParseYongShen(syms[i%len(syms)])
		_ = ComputeYongShen(ch, []YongShenSymbol{s})
	}
}
