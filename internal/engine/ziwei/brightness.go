package ziwei

// brightness level: 0=庙 1=旺 2=利 3=平 4=陷.
type brightness int

const (
	Miao brightness = iota
	Wang
	Li
	Ping
	Xian
)

var brightnessNames = [5]string{"庙", "旺", "利", "平", "陷"}

func (b brightness) String() string { return brightnessNames[b] }

var miaoWangTable [14][12]brightness

func brightnessFrom(s string) brightness {
	for i, name := range brightnessNames {
		if name == s {
			return brightness(i)
		}
	}
	return Ping
}

func miaoWang(star starIndex, zhi Zhi) brightness {
	if star < 0 || int(star) >= 14 {
		return Ping
	}
	z := int(zhi) - 1
	if z < 0 || z >= 12 {
		return Ping
	}
	return miaoWangTable[star][z]
}
