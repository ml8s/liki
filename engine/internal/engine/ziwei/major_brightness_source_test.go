package ziwei

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// iztro v2.6.1 corrected the brightness of three major stars in You.
// This test locks the upstream fix into the engine's data contract.
func TestMajorBrightnessIztroV261YouFixes(t *testing.T) {
	tests := []struct {
		star starIndex
		want brightness
	}{
		{star: TaiYang, want: Ping},
		{star: TaiYin, want: Wang},
		{star: QiSha, want: Wang},
	}

	for _, tt := range tests {
		if got := miaoWang(tt.star, ganzhi.ZhiYou); got != tt.want {
			t.Errorf("miaoWang(%s, 酉) = %s, want %s", starName(tt.star), got, tt.want)
		}
	}
}
