package ziwei

// brightness level: 0=庙 1=旺 2=利 3=平 4=陷.
type brightness int

const (
	Miao brightness = iota
	Wang
	De
	Li
	Ping
	Xian
	Bu
)

var brightnessNames = [7]string{"庙", "旺", "得", "利", "平", "陷", "不"}

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
	// Convert Liki Zhi (子=1) to iztro display index (寅=0)
	iztroIdx := (zhi + 9) % 12
	if int(iztroIdx) < 0 || int(iztroIdx) >= 12 {
		return Ping
	}
	return miaoWangTable[star][iztroIdx]
}
