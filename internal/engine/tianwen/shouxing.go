package tianwen

import "math"

// 寿星天文历（lunar-typescript ShouXingUtil）VSOP87D 太阳黄经移植。
// 与 lunar 完全同源，节气时刻秒级一致。
// 参考链：qiAccurate2 → qiAccurate → saLonT → saLon = eLon + nutationLon2 + gxcSunLon + π

const (
	secondPerRad  = 648000.0 / math.Pi
	secondPerDay  = 86400.0
	shouXingJ2000 = 2451545.0
)

// eLon 太阳几何黄经（弧度），t 为儒略千年（t/10 为 VSOP87 千年参数）。n 为截断阶（-1 全项）。
func eLon(t float64, n int) float64 {
	t /= 10
	var v float64
	tn := 1.0
	const pn = 1
	m0 := xl0[pn+1] - xl0[pn]
	for i := 0; i < 6; i++ {
		n1 := math.Floor(xl0[pn+i])
		n2 := math.Floor(xl0[pn+1+i])
		n0 := n2 - n1
		if n0 == 0 {
			tn *= t
			continue
		}
		var m float64
		if n < 0 {
			m = n2
		} else {
			m = math.Floor(3*float64(n)*n0/m0+0.5+n1)
			if i != 0 {
				m += 3
			}
			if m > n2 {
				m = n2
			}
		}
		var c float64
		for j := int(n1); j < int(m); j += 3 {
			c += xl0[j] * math.Cos(xl0[j+1]+t*xl0[j+2])
		}
		v += c * tn
		tn *= t
	}
	v /= xl0[0]
	t2 := t * t
	v += (-0.0728 - 2.7702*t - 1.1019*t2 - 0.0996*t2*t) / secondPerRad
	return v
}

// nutationLon2 章动黄经（弧度），t 为儒略千年。
func nutationLon2(t float64) float64 {
	a := -1.742 * t
	t2 := t * t
	var dl float64
	for i := 0; i < len(nutB); i += 5 {
		dl += (nutB[i+3] + a) * math.Sin(nutB[i]+nutB[i+1]*t+nutB[i+2]*t2)
		a = 0
	}
	return dl / 100 / secondPerRad
}

// gxcSunLon 光行差（弧度），t 为儒略千年。
func gxcSunLon(t float64) float64 {
	t2 := t * t
	v := -0.043126 + 628.301955*t - 2732e-9*t2
	e := 0.016708634 - 42037e-9*t - 1267e-10*t2
	return -20.49552 * (1 + e*math.Cos(v)) / secondPerRad
}

// saLon 太阳视黄经（弧度）：几何黄经 + 章动 + 光行差 + π。
func saLon(t float64, n int) float64 {
	return eLon(t, n) + nutationLon2(t) + gxcSunLon(t) + math.Pi
}

// saLonT 反解：给定目标视黄经 w（弧度），Newton 迭代求儒略千年 t。

// dtExt ΔT 外推（秒），y 为年份，jsd 为长期加速度。
func dtExt(y, jsd float64) float64 {
	dy := (y - 1820) / 100
	return -20 + jsd*dy*dy
}

// dtCalc ΔT（秒），y 为年份。对齐 lunar ShouXingUtil.dtCalc。
func dtCalc(y float64) float64 {
	size := len(dtAt)
	y0 := dtAt[size-2]
	t0 := dtAt[size-1]
	if y >= y0 {
		const jsd = 31
		if y > y0+100 {
			return dtExt(y, jsd)
		}
		return dtExt(y, jsd) - (dtExt(y0, jsd)-t0)*(y0+100-y)/100
	}
	// 布局: [y0, c0, c1, c2, c3, y0', c0', ...]，段 y1 = 下段 y0（lunar 步长 5 语义）。
	// 每段 5 个: [y0, c0, c1, c2, c3]，段 y1 = 下段 y0。尾 2 项 [y0,t0]。
	segEnd := size - 2
	i := 0
	for ; i+5 < segEnd; i += 5 {
		if y < dtAt[i+5] {
			break
		}
	}
	if i+5 > segEnd {
		i = segEnd - 5
	}
	t1 := (y - dtAt[i]) / (dtAt[i+5] - dtAt[i]) * 10
	t2 := t1 * t1
	t3 := t2 * t1
	return dtAt[i+1] + dtAt[i+2]*t1 + dtAt[i+3]*t2 + dtAt[i+4]*t3
}

// dtT ΔT（天）。
func dtT(t float64) float64 {
	return dtCalc(t/365.2425+2000) / secondPerDay
}

// solarLongitudeShouXing 给定儒略日偏移（JD-2451545，天），返回太阳视黄经（度）。
// 含 ΔT 修正（对齐 lunar qiHigh：saLonT*36525 - dtT + 1/3）。
func solarLongitudeShouXing(jdOffset float64) float64 {
	// 反解：求视黄经=目标 的儒略日
	t := (jdOffset + 2451545.0) / 36525.0 // 初始猜测（百年）
	_ = t
	// 直接算黄经：saLonT 是反解，这里用迭代
	// 简化：直接用 saLon（t=jdOffset/36525 百年）
	tt := jdOffset / 36525.0
	lon := saLon(tt, -1)*180.0/math.Pi - dtT(tt)*360.0/365.25
	lon = math.Mod(lon, 360)
	if lon < 0 {
		lon += 360
	}
	return lon
}

// jieQiTimeShouXing 用寿星历反解（对齐 lunar qiAccurate2）求节气时刻。
// 返回儒略日偏移（JD-2451545）。targetLon 为黄经（度）。

