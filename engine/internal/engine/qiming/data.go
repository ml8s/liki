package qiming

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"log"
	"strconv"
	"strings"

	"liki-engine/internal/engine/ganzhi"

	"gopkg.in/yaml.v3"
)

//go:embed data/gsc_pinyin_with_tone.csv
var gscCSV []byte

//go:embed data/sancai_numbers.yaml
var sanCaiNumbersYAML []byte

//go:embed data/sancai_configs.yaml
var sanCaiConfigsYAML []byte

//go:embed data/negative_chars.txt
var negativeCharsTxt []byte

//go:embed data/radicals.yaml
var radicalsYAML []byte

var charByElement = make(map[Wuxing][]Character)
var charByRune = make(map[rune]Character)
var sanCaiNums map[int]sanCaiNum
var sanCaiCfg map[string]sanCaiCfgEntry
var negativeChars = make(map[string]bool)
var radicalToElement = make(map[string]Wuxing)

type sanCaiNum struct {
	Element string
	Fortune string
	Desc    string
}

type sanCaiCfgEntry struct {
	Fortune string
	Desc    string
}

func init() {
	if err := loadNaming(); err != nil {
		log.Fatalf("qiming: load qiming data: %v", err)
	}
	if err := loadRadicals(); err != nil {
		log.Fatalf("qiming: load radicals: %v", err)
	}
}

// radicalToElement maps Kangxi radicals to five elements per Kangxi dictionary.

// inferElementFromRadical returns the element implied by a Kangxi radical.
func inferElementFromRadical(radical string) (Wuxing, bool) {
	e, ok := radicalToElement[radical]
	return e, ok
}

func loadNaming() error {
	{
		r := csv.NewReader(bytes.NewReader(gscCSV))
		records, err := r.ReadAll()
		if err != nil {
			return err
		}
		for i, rec := range records {
			if i == 0 {
				continue // skip header
			}
			if len(rec) < 11 {
				continue
			}
			word := rec[1]
			if word == "" {
				continue
			}
			elem := wuxingFromChinese(rec[5])
			if elem == 0 {
				var ok bool
				elem, ok = inferElementFromRadical(rec[3])
				if !ok {
					continue
				}
			}
			stroke, err := strconv.Atoi(rec[4])
			if err != nil {
				continue
			}
			tone, err := strconv.Atoi(rec[10])
			if err != nil {
				tone = 0
			}

			// Take the first reading when multiple pinyin are present ("zé,shì" → "ze").
			pinyin := rec[2]
			if idx := strings.IndexByte(pinyin, ','); idx >= 0 {
				pinyin = pinyin[:idx]
			}
			// Strip tone numbers and neutral-tone markers.
			pinyin = strings.TrimRight(pinyin, "0123456789·")

			ce := Character{
				Char:        word,
				Element:     elem,
				Stroke:      stroke,
				Radical:     rec[3],
				Pinyin:      pinyin,
				Tone:        tone,
				Traditional: rec[6],
			}
			charByElement[elem] = append(charByElement[elem], ce)
			for _, r := range word {
				if _, exists := charByRune[r]; !exists {
					charByRune[r] = ce
				}
			}
		}
	}

	{
		// Load negative-meaning characters for name filtering.
		for _, line := range strings.Split(strings.TrimSpace(string(negativeCharsTxt)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				negativeChars[line] = true
			}
		}
	}

	{
		var raw struct {
			Numbers map[int]struct {
				Element string `yaml:"element"`
				Fortune string `yaml:"fortune"`
				Desc    string `yaml:"desc"`
			} `yaml:"numbers"`
		}
		if err := yaml.Unmarshal(sanCaiNumbersYAML, &raw); err != nil {
			return err
		}
		sanCaiNums = make(map[int]sanCaiNum)
		for n, v := range raw.Numbers {
			sanCaiNums[n] = sanCaiNum{
				Element: elementYAMLToChinese(v.Element),
				Fortune: fortuneYAMLToChinese(v.Fortune),
				Desc:    v.Desc,
			}
		}
	}

	{
		var raw struct {
			Configs map[string]struct {
				Fortune string `yaml:"fortune"`
				Desc    string `yaml:"desc"`
			} `yaml:"configs"`
		}
		if err := yaml.Unmarshal(sanCaiConfigsYAML, &raw); err != nil {
			return err
		}
		sanCaiCfg = make(map[string]sanCaiCfgEntry)
		for k, v := range raw.Configs {
			sanCaiCfg[k] = sanCaiCfgEntry{
				Fortune: fortuneYAMLToChinese(v.Fortune),
				Desc:    v.Desc,
			}
		}
	}

	return nil
}

func loadRadicals() error {
	var data struct {
		Radicals map[string][]string `yaml:",inline"`
	}
	if err := yaml.Unmarshal(radicalsYAML, &data); err != nil {
		return err
	}
	elemMap := map[string]Wuxing{
		"木": ganzhi.WxMu, "火": ganzhi.WxHuo, "土": ganzhi.WxTu, "金": ganzhi.WxJin, "水": ganzhi.WxShui,
	}
	for elemName, radicals := range data.Radicals {
		elem, ok := elemMap[elemName]
		if !ok {
			log.Fatalf("qiming: unknown element %q in radicals", elemName)
		}
		for _, r := range radicals {
			radicalToElement[r] = elem
		}
	}
	return nil
}
