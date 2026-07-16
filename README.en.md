<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-灵机-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>liki — AI Chinese Metaphysics Skill Suite</strong>
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

**liki** is a collection of Chinese metaphysics Skills — install it in one command with `npx skills add ml8s/liki`. It includes four sub-skills:

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

- **Agent workflow** — parameter collection → API charting → LLM interpretation → structured report
- **Astronomical backend** — True Solar Time, daylight saving time correction, lat/lon timezone conversion
- **Auto-update** — version check on startup
- **Privacy** — birth info sent to API for stateless computation only; not persisted on the server
- **Ready to use** — one install command, callable immediately

## Quick Start

```bash
npx skills add ml8s/liki
```

Then talk to your AI assistant:

> Read my Bazi chart. Born May 20, 1990, 12:00 noon, Beijing, male.

The agent collects parameters, calls `tianwen.time` for True Solar Time correction, casts the chart, determines Yongshen, and outputs a structured report.

## Installation

```bash
# From GitHub
npx skills add ml8s/liki

# Or from the website
npx skills add https://liki.hk/
```

## Skills Detail

### Mingli (命理)

Combined Bazi and Ziwei analysis — chart casting, Yongshen, He-Hui-Chong-Xing, compatibility (Bond), decade cycles and annual trends.

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

Current stable version: **1.17.0**. See [CHANGELOG.md](./CHANGELOG.md).

## Links

- Website: [liki.hk](https://liki.hk)
- GitHub: [ml8s/liki](https://github.com/ml8s/liki)

## Disclaimer

liki.hk provides Chinese metaphysics analysis from a traditional cultural perspective for research and reference only. Its conclusions **do not constitute** medical diagnosis, legal advice, financial forecasts, or major life decisions. Please maintain a rational and proactive outlook.

---

Questions or suggestions? [Open an Issue](https://github.com/ml8s/liki/issues)
