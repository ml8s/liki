<p align="center">
  <img alt="liki.hk" src="https://img.shields.io/badge/liki.hk-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>liki.hk — AI Skills for Chinese Metaphysics</strong>
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-website-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="./README.md"><img src="https://img.shields.io/badge/中文-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat"></a>
</p>

---

## Introduction

**liki.hk** provides AI Skills for Chinese Metaphysics — install in one command with `npx skills add ml8s/liki`. It covers four areas:

- **[Mingli (命理)](./mingli/)** — Bazi & Ziwei: chart casting, Yongshen analysis, Combination & Clash, decade/annual fortune
- **[Qiming (起名)](./qiming/)** — Naming based on Bazi Yongshen + Wuge-Sancai numerology
- **[Wengua (问卦)](./wengua/)** — Intelligent routing: Qimen Dunjia / Liuyao / Huangli almanac
- **[Fengshui (风水)](./fengshui/)** — Bazhai life-trigram + Xuankong Flying Star analysis

## What You Can Do

| You want to | Use |
|-------------|-----|
| Read a Bazi or Ziwei chart, check decade/annual fortune | **Mingli** |
| Name a baby or company | **Qiming** |
| Pick an auspicious date, ask about fortune, find a lost item | **Wengua** |
| Evaluate home or office Fengshui | **Fengshui** |

## Features

- **Precise API charting** — astronomical backend with True Solar Time, DST, lat/lon conversion
- **Full coverage** — Bazi, Ziwei, naming, divination, Fengshui in one suite
- **Smart routing** — one sentence auto-routes to the right domain, no skill selection needed
- **Privacy** — birth info sent to API for stateless computation only; not persisted

## Quick Start

```bash
npx skills add ml8s/liki
npx skills add https://liki.hk/   # or from the website
```

Then talk to your AI assistant:

## Skills Detail

### Mingli (命理)

Combined Bazi and Ziwei analysis — chart casting, Yongshen, Geju determination, Combination & Clash, compatibility (Bond), decade cycles, annual trends, comprehensive advice, and historical event calibration.

> Bazi chart, 1990-05-20 12:00 Beijing, male
> What's my fortune this year
> Check our compatibility: me (1984-06-15 Beijing, male) and her (1986-09-20 Shanghai, female)
> ZiWei chart for 1990-05-20 Beijing, male

### Qiming (起名)

Naming service that combines Bazi Yongshen with Wuge (五格, five-grid) and Sancai (三才, three-talent) numerology.

> Name my baby boy, born 2024-03-15 at 10:00 Guangzhou, surname Wang
> Help me rename myself, born 1995-08-10 Shenzhen, female, surname Li
> Review the name 'Zhang Wei'

### Wengua (问卦)

LLM-driven routing that selects the corresponding method based on your question:

- **Auspicious date** (when is good) → Huangli (黄历, Chinese almanac)
- **Fortune inquiry** (success/loss/finding things) → Liuyao (六爻, I-Ching)
- **Strategy & direction** (what to do / where to go) → Qimen Dunjia (奇门遁甲)

> Should we move house tomorrow?
> Can I find my lost keys?
> How's my career luck lately?
> Which direction should I go today?

### Fengshui (风水)

Covers both Bazhai (八宅, Eight Mansions) and Xuankong (玄空, Flying Star) schools — life-trigram matching, spatial orientation, and time-dimension analysis via Sanyuan Jiuyun (三元九运).

> My front door faces east. How's the Fengshui of this apartment?
> This house faces south. How's the Fengshui?
> My office is on the 8th floor, desk by the window — anything to watch out for?

## Version

Current stable version: **1.23.0**. See [CHANGELOG.md](./CHANGELOG.md).

## References

This project's design references the following open-source projects:

- [weizeW/mingli-skills](https://github.com/weizeW/mingli-skills) — Four-dimensional cross-validation framework
- [jinchenma94/bazi-skill](https://github.com/jinchenma94/bazi-skill) — Nine classical texts structured summary, historical event calibration
- [dzcmemory-web/bazi-ziwei-skill](https://github.com/dzcmemory-web/bazi-ziwei-skill) — Bazi+Ziwei cross-validation mode
- [yanouyuan-bit/bazi-roundtable](https://github.com/yanouyuan-bit/bazi-roundtable) — Multi-school review and conclusion strength labeling
- [shizhilya/yuan](https://github.com/shizhilya/yuan) — Conclusion-first output design
- [ai-freer/fortune-skill](https://github.com/ai-freer/fortune-skill) — Three-layer computation architecture
- [hhszzzz/taibu](https://github.com/hhszzzz/taibu) — Agent-friendly design
- [SylarLong/iztro](https://github.com/SylarLong/iztro) — Ziwei Doushu charting engine
- [2021291696/high-confidence-mingli-skill](https://github.com/2021291696/high-confidence-mingli-skill) — Confidence system, personality profile inference
- [DestinyLinker/MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) — Metaphysics benchmark reference

## Links

- Website: [liki.hk](https://liki.hk)
- GitHub: [ml8s/liki](https://github.com/ml8s/liki)

## Disclaimer

liki.hk provides Chinese metaphysics analysis from a traditional cultural perspective for research and reference only. Its conclusions **do not constitute** medical diagnosis, legal advice, financial forecasts, or major life decisions. Please maintain a rational and proactive outlook.

---

