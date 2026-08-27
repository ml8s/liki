<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-Skill_for_Chinese_Metaphysics-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>Liki — Professional Skill for Chinese Metaphysics</strong><br>
  Built to professional standards: astronomical-engine charting, classically-sourced judgments, verifiable conclusions<br>
  BaZi · ZiWei · Liuyao · QiMen · Date Selection · Feng Shui · Naming
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://github.com/ml8s/liki"><img src="https://img.shields.io/badge/GitHub-ml8s/liki-4a9e6b?style=flat&logo=github&logoColor=white&labelColor=30305c"></a>
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-website-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="https://github.com/ml8s/liki/actions/workflows/ci.yml"><img src="https://img.shields.io/badge/CI-passing-4a9e6b?style=flat&logo=githubactions&logoColor=white&labelColor=30305c"></a>
  <a href="./README.md"><img src="https://img.shields.io/badge/中文-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
</p>

---

## In 30 Seconds

After installation, your AI assistant gains 4 metaphysics skills:

| Skill | What you can ask | Try this |
|-------|------------------|----------|
| **liki-bazi** destiny | Marriage, career, wealth, health, education, personality, family, compatibility, full life report | `Read my BaZi, born 1990-05-20 12:00 in Beijing, male` |
| **liki-naming** naming | Baby naming, renaming, company naming, Chinese names for English speakers, name evaluation | `Name my baby, born 2024-06-10 in Guangzhou, male, surname Chen` |
| **liki-divination** divination | Liuyao (outcome & timing), QiMen (direction & decision), auspicious date selection | `Will this work out? When will I see results?` |
| **liki-fengshui** feng shui | Bazhai chart & layout, Xuankong flying stars, annual feng shui | `How is the feng shui of my home?` |

**What professional standards mean here:**

- Charts are computed by an astronomical engine (true solar time, second-level solar terms) — the AI never invents numbers
- Judgments come from 597 truth-table rules, each citing classical sources (Ziping Zhenquan, Dih Tian Sui, etc.)
- Independently evaluated on 160 competition questions with answer isolation

## Installation

```bash
npx skills add ml8s/liki          # install all 4 skills
```

Install one:

```bash
npx skills add ml8s/liki --skill liki-bazi       # destiny (BaZi + ZiWei)
npx skills add ml8s/liki --skill liki-naming     # naming
npx skills add ml8s/liki --skill liki-divination # divination
npx skills add ml8s/liki --skill liki-fengshui   # feng shui
```

**After installing, start like this:**

```
Give me a full life reading, born 1990-05-20 12:00 in Beijing, male
Are we compatible? I was born 1992-03-15, she on 1994-08-20
How will my career and wealth go in 2026?
```

## User Guide

### liki-bazi (BaZi + ZiWei dual-chart)

**How to ask** — ask by life domain; provide birth info (date + time + place + gender):

- **Marriage**: `When will I marry?` `Will we divorce?` `What will my partner be like?`
- **Career**: `What industry suits me? Startup or employment?` `Which years bring career shifts?`
- **Wealth**: `What's my wealth source? Which years gain, which lose?`
- **Health**: `Which organ systems are weak? Which years need care?`
- **Education**: `How far will my education go? Exam luck?`
- **Personality / Appearance / Family**: `What's my personality?` `Parents/children affinity?`
- **Compatibility**: `Are we a match?` (both parties' birth info)
- **Full report**: `Give me a full life reading`

**What you get** — conclusion first, with basis and timing:

> Marriage palace stable; a real-relationship window opens in H2 2026 — Red Luan enters the spouse palace and the decade fortune reveals the wealth star. Spouse-star analysis: … palace checks: … (basis attached at each step)

### liki-naming

Yong-shen (favorable element) from BaZi → Sancai-Wuge numerology → candidate characters. Every recommended name carries its basis:

> Top pick: Guanlan — from Mencius "observe the waves", all-auspicious Sancai, supplements the Fire yong-shen.

### liki-divination

- **Liuyao**: outcomes and timing (`Will this succeed?` `When?`)
- **QiMen**: direction and decision (`Which direction?` `Should I act now?`)
- **Date selection**: auspicious dates (`Best day to move / sign / open business?`)

Output lists the hexagram basis first (yong-shen / shi-ying / moving lines), then a one-line verdict.

### liki-fengshui

- **Bazhai**: `What's my ming gua?` `How to arrange door/kitchen/bedroom?`
- **Xuankong**: `Is my home favorable this period?` `2026 annual cautions?`

### FAQ

**Does it need internet?**
Chart computation goes through the liki.hk JSON-RPC engine. If unreachable, the skill says so explicitly — it never falls back to "AI guesswork".

**Is my birth data stored?**
No. The skill explicitly commits: no birth-info storage outside your conversation, no real names requested; chart data lives only in your chat context.

**Don't know the exact birth hour?**
A built-in **calibration flow**: offer 2-3 candidate hours plus 3-5 life events with years; it cross-checks each chart and infers the most likely hour (with confidence). Babies/teens skip calibration and use a default hour.

**How should I interpret results?**
Conclusions are from a traditional-cultural perspective, for reference only — not medical, legal, or financial advice. Every conclusion carries its basis and classical citation so you can verify it yourself.

**How do I update?**
The skill self-checks its version (local fingerprint + remote) on startup; when prompted, re-run: `npx skills add ml8s/liki -y`.

## Why It's Trustworthy

- **Engine-computed, not AI-invented** — charts come from a Go astronomical engine: true solar time, DST, longitude-based timezone, VSOP87D second-level solar terms. The model interprets; it never computes charts.
- **Sourced judgments** — 46 truth tables with 597 rules, each with a classical-citation column; plus 7 annual-timing tables.
- **Dual-system cross-check** — BaZi leads, ZiWei verifies; conflicts resolved with explicit evidence.
- **Auditable process** — every step fills a checklist; conclusions trace back to specific steps.
- **Independent evaluation** — 160 competition questions (MingLi-Bench), answer isolation, public data (`tests/`).

---

## For Developers

### Architecture

```
skills/liki-bazi
├── SKILL.md    ← rules (process skeleton + hard constraints)
├── app/        ← process (9 cards: marriage/career/wealth/…)
├── domains/    ← knowledge (bazi 16 + ziwei 8 docs)
└── tools/      ← tools (5 tools + 46 judgment CSVs + factor tables)
repo root
├── engine/     ← Go JSON-RPC astronomical engine (8 domains)
├── tests/      ← evaluation (160 grouped cases + answer isolation)
└── scripts/    ← build / fingerprinting
```

Call chain: SKILL.md routes to an app card → the card calls tools (RPC charting + CSV matching) → interpreted via domains knowledge → rendered by the card's template. The tool layer is optional (liki-bazi has it; the other three use RPC + document translation directly).

### Quick Start

```bash
make hooks         # install git hooks (once)
make test-all      # full: skills unit + engine (lint/vet/race/integration/smoke) + e2e
make check         # table schema + doc contracts + version consistency
make build-archive # pack 4 skills + recompute content fingerprints
```

### Design Principles

- Single responsibility per layer: root=rules, app=process, domains=knowledge, tools=tools
- Single data source: parameters from rpc.discover; judgments from CSV truth tables
- Dual-system: BaZi leads, ZiWei reviews, conflicts explicit
- Semver + content fingerprints (anti stale-sync)
- Evaluation-driven: independent grading, answer isolation, public data

See [CONTRIBUTING.md](./CONTRIBUTING.md) and [CHANGELOG.md](./CHANGELOG.md). Design references include [mingli-skills](https://github.com/weizeW/mingli-skills), [bazi-skill](https://github.com/jinchenma94/bazi-skill), [iztro](https://github.com/SylarLong/iztro), and [MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench).

## License & Disclaimer

MIT. Conclusions are from a traditional cultural perspective, for research and reference only — they do **not** constitute medical diagnosis, legal advice, financial forecasts, or major life decisions. Please stay rational and proactive.
