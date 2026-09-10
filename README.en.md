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
| **liki-naming** naming | Baby naming, renaming, Chinese names for English speakers, name evaluation | `Name my baby, born 2024-06-10 in Guangzhou, male, surname Chen` |
| **liki-divination** divination | Liuyao (outcome & timing), QiMen (direction & decision), auspicious date selection | `Will this work out? When will I see results?` |
| **liki-fengshui** feng shui | Bazhai chart & layout, Xuankong flying stars, annual feng shui | `How is the feng shui of my home?` |

**What professional standards mean here:**

- Charts are computed by an astronomical engine (true solar time, second-level solar terms) — the AI never invents numbers
- Judgments come from 775 truth-table rules, each citing classical sources (Ziping Zhenquan, Dih Tian Sui, etc.)
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

### Getting Started

**Prepare**: birth date (Gregorian), birth time (to the minute if possible), birth city, gender.

**Just send it** — birth info and your question in one message:

> Help me look at marriage, female, born 1992-03-15 14:30 in Guangzhou

Missing details are fine: if you only know "morning" or don't know the time, the skill will follow up or start the calibration flow (see FAQ).

### How to Talk to the Skill

The skill works like a practitioner — **one topic at a time**, with structured analysis you can drill into:

| You want to… | Say |
|------|-------|
| Ask why | `Why?` `What's the basis?` |
| Ask about a year | `What about 2026?` `Next three years?` |
| Switch topic | `What about wealth?` `Health?` (same chart, no re-compute) |
| Compatibility | `Are we compatible?` (provide both birth infos) |
| Full report | `Give me a full life reading` |

### Detailed Guide by Skill

#### liki-bazi (BaZi + ZiWei dual-chart)

Ask by life domain — the skill automatically charts, queries judgment tables, and gives conclusion + basis + timing:

- **Marriage**: When to marry? Will we divorce? What's my partner like?
- **Career**: Which industry? Startup or employment? Which years shift?
- **Wealth**: Wealth source? Which years gain, which lose?
- **Health**: Which organ systems? Which years to watch?
- **Education**: How far? Exam luck?
- **Personality / Family**: What's my personality? Parents/children affinity?

**Output format**: conclusion first, basis attached. Every conclusion traces to specific steps and classical sources.

#### liki-naming

> Name my baby, born 2024-06-10 in Guangzhou, male, surname Chen

Flow: preference and taboo intake → BaZi yong-shen → five-element supplement → candidate filtering → composition and evaluation → candidate reports (strengths, trade-offs, rejection reasons, and source evidence).

Also supports: renaming, Chinese names for English speakers, name evaluation.

#### liki-divination

The skill chooses one method by user goal; it does **not** run dual divination by default:

| Goal | Default method | Try |
|---|---|---|
| Event outcome / timing | Liuyao | `Will this project succeed?` |
| Action, direction, strategy, timing | QiMen | `Should I sign now?` |
| Date selection | HuangLi | `Best day to move / sign / open?` |

If a request mixes outcome and strategy, the skill asks you to choose the primary question first. Explicit dual-method cross-check remains available on request.

Ordinary Qimen questions require **no chart-method selection**; the skill defaults to hour-scope Qimen with the rotating plate and chai-bu bureau. To choose an explicit method:

| Want | Try |
|---|---|
| Zhirun bureau | `Use the zhirun chart for now` |
| Luo Shu flying plate | `Use the Luo Shu flying plate for this matter` |
| Ten-minute Kejia | `Use the ten-minute Kejia chart` |
| Twelve-minute ten-division | `Use the twelve-minute ten-division chart` |
| Golden Mirror | `Use Golden Mirror for today` |

Output: method basis → one-line verdict → timing / direction → practical advice.
Liuyao and Qimen retain casting/charter receipts, snapshots, evidence references, conflict signals, audit results, and session integrity summaries. Follow-ups reuse the original chart.

Domain contract documents are listed under **For Developers → Domain contracts**.

#### liki-fengshui

- **Bazhai**: `What's my ming gua?` `How to arrange door/kitchen/bedroom?`
- **Xuankong**: `Is my home favorable this period?` `2026 annual cautions?`

### FAQ

**Don't know the exact birth hour?**
Offer 2-3 candidate hours + 3-5 life events with years; the skill cross-checks and infers the most likely hour (with confidence). Babies/teens skip calibration.

**Does it need internet?**
Chart computation goes through the liki.hk engine. If unreachable, the skill says so explicitly — never falls back to "AI guesswork".

**Is my birth data stored?**
No. The skill explicitly commits: no birth-info storage outside your conversation, no real names requested; chart data lives only in your chat context.

**How should I interpret results?**
Every conclusion carries its basis and classical citation — verify it yourself. Traditional cultural perspective, not medical/legal/financial advice.

**How do I update?**
The skill self-checks its version on startup; when prompted, re-run: `npx skills add ml8s/liki -y`.

**Self-hosting an engine?**
Starting with `2026.09.10.5`, Skill and engine RPC contracts ship together. Skills fail closed when the engine is older; update and restart liki-engine before updating the Skill.

## Why It's Trustworthy

- **Engine-computed, not AI-invented** — charts come from a Go astronomical engine: true solar time, DST, longitude-based timezone, VSOP87D second-level solar terms. The model interprets; it never computes charts.
- **Sourced judgments** — 47 logical rule groups with 775 assertions, each with a classical-citation column.
- **Dual-system cross-check** — BaZi and ZiWei are evaluated separately, with an explicit synthesis layer; conflicts are resolved with explicit evidence.
- **Auditable process** — divination flows retain casting/charter receipts, snapshots, evidence references, report audits, and session integrity summaries; conclusions trace back to specific steps.
- **Independent evaluation** — 160 competition questions (MingLi-Bench) for accuracy, plus cross-domain skill-up smoke tests for behavior contracts.

---

## For Developers

### Architecture

```
skills/liki-bazi
├── SKILL.md    ← rules (process skeleton + hard constraints)
├── app/        ← process (10 cards: marriage/career/wealth/…)
├── domains/    ← knowledge (bazi 16 + ziwei 8 docs)
└── tools/      ← tools (6 Python tools + 2 assertion tables + 2 factor tables)
repo root
├── engine/     ← Go JSON-RPC astronomical engine (8 domains)
├── tests/      ← accuracy benchmark (160 grouped cases) + cross-domain behavior smoke
└── scripts/    ← build / distribution index
```

Call chain: SKILL.md routes to an app card → the card calls the six Python tools (`agent_cli.py` orchestrates RPC charting, factor evaluation, and CSV matching) → interpreted via domain knowledge → rendered by the card template. RPC methods are invisible to the liki-bazi LLM.

### Domain contracts

- [docs/DIVINATION_MODEL.md](./docs/DIVINATION_MODEL.md) — divination domain model and layers: casting, snapshot, evidence, answer, and audit boundaries.
- [docs/BAZI_MODEL.md](./docs/BAZI_MODEL.md) — BaZi domain model covering Four Pillars and Zi Wei: engine atomic facts, factor predicates, assertions, and query boundaries.
- [docs/FENGSHUI_MODEL.md](./docs/FENGSHUI_MODEL.md) — Feng Shui domain model and layers: Bazhai ming gua, door/master/stove, Xuankong flying stars, periods, and annual boundaries.
- [docs/NAMING_MODEL.md](./docs/NAMING_MODEL.md) — naming domain model and layers: BaZi yongshen strategy, character pools, candidate names, foreign surname candidates, evaluation, and source boundaries.

The complete factor inventory is sourced solely from `skills/liki-bazi/tools/factors/*.csv`.

### Engine Image

Images are auto-published on GitHub Releases (CI full tests → build + push + smoke test): `docker pull ghcr.io/ml8s/liki-engine:latest`. Build from source: `cd engine && docker compose -f deploy/docker-compose.yml up -d --build`.

### Test Commands

```bash
make benchmark-mingli160       # 160-question accuracy benchmark (model required)
make skillup-smoke-validate    # validate cross-domain behavior contracts (no model)
make skillup-smoke             # run cross-domain behavior smoke (model required)
```
### Quick Start

```bash
make hooks         # install git hooks (once)
make test-all      # full: skills unit + engine (lint/vet/race/integration/smoke) + e2e
make check         # table schema + doc contracts + version consistency
make build-archive # pack 4 skills + generate the distribution index/archive digest
```

### Design Principles

- Single responsibility per layer: root=rules, app=process, domains=knowledge, tools=tools
- Single data source: LLM tool contracts from `tools/skill-tools.json`; judgments from CSV truth tables
- Dual-system: BaZi leads, ZiWei reviews, conflicts explicit
- CalVer (date-stamped VERSION + CHANGELOG; startup version self-check; tags on milestones)
- Evaluation-driven: independent grading, answer isolation, public data

See [CONTRIBUTING.md](./CONTRIBUTING.md) and [CHANGELOG.md](./CHANGELOG.md). Design references include [mingli-skills](https://github.com/weizeW/mingli-skills), [bazi-skill](https://github.com/jinchenma94/bazi-skill), [iztro](https://github.com/SylarLong/iztro), and [MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench).


## License & Disclaimer

MIT. Conclusions are from a traditional cultural perspective, for research and reference only — they do **not** constitute medical diagnosis, legal advice, financial forecasts, or major life decisions. Please stay rational and proactive.
