package huangli

import (
	_ "embed"
	"encoding/json"
	"log"

	"liki-engine/internal/engine/ganzhi"

	"gopkg.in/yaml.v3"
)

var huangDaoStars [12]huangDaoStar
var qingLongStart map[ganzhi.Zhi]ganzhi.Zhi

//go:embed data/jianchu.yaml
var jianchuYAML []byte

//go:embed data/huangdao.json
var huangdaoJSON []byte

var jianChuCfg jianchuConfig

func init() {
	if err := loadJianchu(); err != nil {
		log.Fatalf("huangli: load jianchu: %v", err)
	}
	if err := loadHuangdao(); err != nil {
		log.Fatalf("huangli: load huangdao: %v", err)
	}
}

func loadJianchu() error {
	// Parse into anonymous struct — yaml.v3 skips map values with named types.
	var raw struct {
		Jianchu struct {
			Sequence   []string            `yaml:"sequence"`
			Suitable   map[string][]string `yaml:"suitable"`
			Forbidden  map[string][]string `yaml:"forbidden"`
			EventRules map[string]struct {
				Label     string   `yaml:"label"`
				Suitable  []string `yaml:"suitable"`
				Forbidden []string `yaml:"forbidden"`
			} `yaml:"event_rules"`
			ShenSha map[string]map[string][]string `yaml:"shensha"`
		} `yaml:"jianchu"`
	}
	if err := yaml.Unmarshal(jianchuYAML, &raw); err != nil {
		return err
	}

	cfg := raw.Jianchu
	c := &jianChuCfg
	c.Sequence = cfg.Sequence
	c.Suitable = cfg.Suitable
	c.Forbidden = cfg.Forbidden

	c.EventRules = make(map[string]eventRule, len(cfg.EventRules)*2)
	for k, v := range cfg.EventRules {
		r := eventRule{Label: v.Label, Suitable: v.Suitable, Forbidden: v.Forbidden}
		c.EventRules[k] = r
		c.EventRules[v.Label] = r
	}

	c.ShenSha = make(map[string]map[string][]string, len(cfg.ShenSha))
	for k, v := range cfg.ShenSha {
		c.ShenSha[k] = v
	}
	return nil
}

func loadHuangdao() error {
	var data struct {
		Stars []struct {
			Name string `json:"name"`
			Path string `json:"path"`
		} `json:"stars"`
		QingLongStart map[string]string `json:"qing_long_start"`
	}
	if err := json.Unmarshal(huangdaoJSON, &data); err != nil {
		return err
	}

	for i, s := range data.Stars {
		huangDaoStars[i] = huangDaoStar{
			Index:    i,
			Name:     s.Name,
			Path:     s.Path,
			Sequence: i,
		}
	}

	qingLongStart = make(map[ganzhi.Zhi]ganzhi.Zhi, len(data.QingLongStart))
	for branchStr, startStr := range data.QingLongStart {
		branch, err := ganzhi.ParseZhi(branchStr)
		if err != nil {
			return err
		}
		start, err := ganzhi.ParseZhi(startStr)
		if err != nil {
			return err
		}
		qingLongStart[branch] = start
	}

	return nil
}
