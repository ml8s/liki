package ganzhi

import "testing"

func FuzzParseGender(f *testing.F) {
	f.Add("male")
	f.Add("female")
	f.Add("男")
	f.Add("女")
	f.Add("unknown")
	f.Add("")
	f.Add("MALE")
	f.Add("Male")

	f.Fuzz(func(t *testing.T, s string) {
		g, err := ParseGender(s)
		switch {
		case s == "male" || s == "男":
			if err != nil {
				t.Errorf("ParseGender(%q) unexpected error: %v", s, err)
			}
			if g != Male {
				t.Errorf("ParseGender(%q) = %v, want Male", s, g)
			}
		case s == "female" || s == "女":
			if err != nil {
				t.Errorf("ParseGender(%q) unexpected error: %v", s, err)
			}
			if g != Female {
				t.Errorf("ParseGender(%q) = %v, want Female", s, g)
			}
		default:
			if err == nil {
				t.Errorf("ParseGender(%q) = %v, want error", s, g)
			}
		}
	})
}

func FuzzParseGan(f *testing.F) {
	f.Add("甲")
	f.Add("乙")
	f.Add("癸")
	f.Add("")
	f.Add("wood")
	f.Add("A")

	f.Fuzz(func(t *testing.T, s string) {
		g, err := ParseGan(s)
		if err == nil {
			if g < 1 || g > 10 {
				t.Errorf("ParseGan(%q) = %v, nil; want value in [1,10]", s, g)
			}
		} else {
			if g != 0 {
				t.Errorf("ParseGan(%q) = (%v, %v); want zero value on error", s, g, err)
			}
		}
	})
}

func FuzzParseZhi(f *testing.F) {
	f.Add("子")
	f.Add("亥")
	f.Add("")
	f.Add("earth")
	f.Add("猫")

	f.Fuzz(func(t *testing.T, s string) {
		z, err := ParseZhi(s)
		if err == nil {
			if z < 1 || z > 12 {
				t.Errorf("ParseZhi(%q) = %v, nil; want value in [1,12]", s, z)
			}
		} else {
			if z != 0 {
				t.Errorf("ParseZhi(%q) = (%v, %v); want zero value on error", s, z, err)
			}
		}
	})
}

func FuzzParseWuxing(f *testing.F) {
	f.Add("木")
	f.Add("水")
	f.Add("wood")
	f.Add("Fire")
	f.Add("")
	f.Add("earth")

	f.Fuzz(func(t *testing.T, s string) {
		w, err := ParseWuxing(s)
		if err == nil {
			if w < WxMu || w > WxShui {
				t.Errorf("ParseWuxing(%q) = %v, nil; want value in [Mu,Huo,Tu,Jin,Shui]", s, w)
			}
		} else {
			if w != 0 {
				t.Errorf("ParseWuxing(%q) = (%v, %v); want zero value on error", s, w, err)
			}
		}
	})
}

func FuzzNayinWuxing(f *testing.F) {
	f.Add("海中金")
	f.Add("炉中火")
	f.Add("大林木")
	f.Add("路旁土")
	f.Add("")
	f.Add("金")

	f.Fuzz(func(t *testing.T, s string) {
		wx := NayinWuxing(s)
		if wx != 0 && (wx < WxMu || wx > WxShui) {
			t.Errorf("NayinWuxing(%q) = %v; want 0 or valid Wuxing [%d,%d]", s, wx, WxMu, WxShui)
		}
	})
}

func FuzzParseShiShen(f *testing.F) {
	f.Add("比肩")
	f.Add("正官")
	f.Add("正印")
	f.Add("")
	f.Add("魁罡")

	f.Fuzz(func(t *testing.T, s string) {
		ss, err := ParseShiShen(s)
		if err == nil {
			if ss < 0 || ss > ShiShenZhengYin {
				t.Errorf("ParseShiShen(%q) = %v, nil; want 0..9", s, ss)
			}
		} else {
			if ss != -1 {
				t.Errorf("ParseShiShen(%q) = (%v, %v); want -1 on error", s, ss, err)
			}
		}
	})
}

func FuzzShiShenUnmarshalJSON(f *testing.F) {
	f.Add([]byte(`"比肩"`))
	f.Add([]byte(`"invalid"`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var ss ShiShen
		err := ss.UnmarshalJSON(data)
		if err == nil {
			if ss < 0 || ss > ShiShenZhengYin {
				t.Errorf("UnmarshalJSON(%s) = %v, nil; want value in [0,9] or error", data, ss)
			}
		}
	})
}

func FuzzParseYinYang(f *testing.F) {
	f.Add("阳")
	f.Add("阴")
	f.Add("")
	f.Add("yin")
	f.Add("Yang")

	f.Fuzz(func(t *testing.T, s string) {
		yy, err := ParseYinYang(s)
		if err == nil {
			if yy != Yang && yy != Yin {
				t.Errorf("ParseYinYang(%q) = %v, nil; want Yin or Yang", s, yy)
			}
		} else {
			if yy != false {
				t.Errorf("ParseYinYang(%q) = (%v, %v); want false on error", s, yy, err)
			}
		}
	})
}

func FuzzWuxingUnmarshalJSON(f *testing.F) {
	f.Add([]byte(`"木"`))
	f.Add([]byte(`"water"`))
	f.Add([]byte(`1`))
	f.Add([]byte(`"invalid"`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`0`))
	f.Add([]byte(`7`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var w Wuxing
		err := w.UnmarshalJSON(data)
		if err == nil {
			if w < WxMu || w > WxShui {
				t.Errorf("UnmarshalJSON(%s) = %v, nil; want value in [%d,%d] or error", data, w, WxMu, WxShui)
			}
		}
	})
}

func FuzzGanUnmarshalJSON(f *testing.F) {
	f.Add([]byte(`"甲"`))
	f.Add([]byte(`1`))
	f.Add([]byte(`11`))
	f.Add([]byte(`"invalid"`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var g Gan
		err := g.UnmarshalJSON(data)
		if err == nil {
			if g < 0 || g > 10 {
				t.Errorf("UnmarshalJSON(%s) = %v, nil; want value in [0,10] or error", data, g)
			}
		}
	})
}

func FuzzZhiUnmarshalJSON(f *testing.F) {
	f.Add([]byte(`"子"`))
	f.Add([]byte(`1`))
	f.Add([]byte(`13`))
	f.Add([]byte(`"invalid"`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var z Zhi
		err := z.UnmarshalJSON(data)
		if err == nil {
			if z < 0 || z > 12 {
				t.Errorf("UnmarshalJSON(%s) = %v, nil; want value in [0,12] or error", data, z)
			}
		}
	})
}
