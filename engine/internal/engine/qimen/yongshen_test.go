package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func chartForTest() Chart {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	return ComputeChart(st)
}

func TestParseYongShen(t *testing.T) {
	for _, name := range []string{"开门", "天禽", "太阴", "白虎", "太常", "戊"} {
		if _, err := ParseYongShen(name); err != nil {
			t.Errorf("ParseYongShen(%q): %v", name, err)
		}
	}
	for _, name := range []string{"火星", "甲"} {
		if _, err := ParseYongShen(name); err == nil {
			t.Errorf("ParseYongShen(%q) should reject", name)
		}
	}
}

func TestComputeYongShenSymbols(t *testing.T) {
	chart := chartForTest()
	symbols := make([]YongShenSymbol, 0, 2)
	for _, name := range []string{"生门", "戊"} {
		sym, err := ParseYongShen(name)
		if err != nil {
			t.Fatal(err)
		}
		symbols = append(symbols, sym)
	}
	result := computeYongShenSymbols(chart, symbols)
	if len(result.Symbols) != 2 {
		t.Fatalf("symbols = %d, want 2", len(result.Symbols))
	}
	for _, symbol := range result.Symbols {
		if symbol.Palace == 0 || len(symbol.TianGan) == 0 {
			t.Errorf("%+v has no palace or heaven gan", symbol)
		}
	}
}

func TestYongShenSpiritNameMatchesCurrentChart(t *testing.T) {
	yangChart := chartForTest()
	if yangChart.Pan.YinDun {
		t.Fatal("first fixture must be an Yang Dun chart")
	}
	yinChart := ComputeChart(tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	))
	if !yinChart.Pan.YinDun {
		t.Fatal("second fixture must be a Yin Dun chart")
	}

	for _, chart := range []Chart{yangChart, yinChart} {
		for _, spirit := range []SpiritIndex{SpiritGouChen, SpiritZhuQue} {
			inputName := spirit.YinName()
			displayName := spirit.YangName()
			if chart.Pan.YinDun {
				inputName, displayName = spirit.YangName(), spirit.YinName()
			}
			symbol, err := ParseYongShen(inputName)
			if err != nil {
				t.Fatal(err)
			}
			result := computeYongShenSymbols(chart, []YongShenSymbol{symbol})
			if len(result.Symbols) != 1 {
				t.Fatalf("symbols = %d, want 1", len(result.Symbols))
			}
			if got := result.Symbols[0].Symbol; got != displayName {
				t.Fatalf("spirit symbol = %q, want current-chart name %q", got, displayName)
			}
		}
	}
}

func TestYongShenBirthDateBoundary(t *testing.T) {
	chart := chartForTest()
	before := computeYongShenWithBirth(chart, nil, BirthDate{
		Time: time.Date(2024, 2, 4, 2, 0, 0, 0, time.UTC), Precision: BirthDateMoment, Has: true,
	})
	after := computeYongShenWithBirth(chart, nil, BirthDate{
		Time: time.Date(2024, 2, 4, 18, 0, 0, 0, time.UTC), Precision: BirthDateMoment, Has: true,
	})
	if before.NianGanPalace == nil || after.NianGanPalace == nil {
		t.Fatal("year-stem palace missing")
	}
	if before.Symbols == nil || after.Symbols == nil {
		t.Fatal("empty selected-symbol collection must serialize as an array")
	}
	if *before.NianGanPalace == *after.NianGanPalace {
		t.Error("precise Lichun boundary did not change the year stem")
	}
}
