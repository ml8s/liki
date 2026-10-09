package ganzhi

import "testing"

func BenchmarkWangShuaiOf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WangShuaiOf(WxMu, ZhiYin)
	}
}

func BenchmarkSixtyCycleIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SixtyCycleIndex(GanJia, ZhiZi)
	}
}

func BenchmarkSheng(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sheng(WxMu, WxHuo)
	}
}

func BenchmarkKe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Ke(WxMu, WxTu)
	}
}

func BenchmarkIsZhiHe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsZhiHe(ZhiZi, ZhiChou)
	}
}

func BenchmarkIsLiuChong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsLiuChong(ZhiZi, ZhiWu)
	}
}

func BenchmarkIsTripleHe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsTripleHe(ZhiZi, ZhiChou)
	}
}
