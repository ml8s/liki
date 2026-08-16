<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-Professional_Skill_for_Chinese_Metaphysics-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>Liki — Professional Skill for Chinese Metaphysics（v4.1.0）</strong>
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://github.com/ml8s/liki"><img src="https://img.shields.io/badge/GitHub-ml8s/liki-4a9e6b?style=flat&logo=github&logoColor=white&labelColor=30305c"></a>
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-website-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="https://github.com/ml8s/liki/actions/workflows/ci.yml"><img src="https://github.com/ml8s/liki/actions/workflows/ci.yml/badge.svg"></a>
  <a href="./README.md"><img src="https://img.shields.io/badge/中文-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat"></a>
</p>

---

**Liki** is a Skill for Chinese metaphysics: **BaZi, ZiWei, Liuyao, Qimen, Huangli date selection, Bazhai feng shui, Xuankong feng shui, and naming** — 8 domains in your AI assistant. It implements traditional metaphysics as an engineered tool with **executable process, verifiable conclusions, and traceable judgments**: charts are computed by an astronomical engine, judgments are driven by 46 truth tables (590 rules with classical references), each process step fills in a form, and conclusions are calibrated against known life events.

## Quick Start

Install all 4 skills at once:

```bash
npx skills add ml8s/liki
```

| skill | domain | install one |
|-------|--------|-------------|
| `liki-bazi` | metaphysics (BaZi + ZiWei dual-chart) | `--skill liki-bazi` |
| `liki-divination` | divination (Liuyao/Qimen/Huangli) | `--skill liki-divination` |
| `liki-fengshui` | feng shui (Bazhai/Xuankong) | `--skill liki-fengshui` |
| `liki-naming` | naming (BaZi yongshen + wuge-sancai) | `--skill liki-naming` |

Install one: `npx skills add ml8s/liki --skill liki-bazi`

Then just talk to your AI assistant; what each skill does and how to ask — see "Features" below.

## Features

### liki-bazi — metaphysics (BaZi + ZiWei dual-chart)

Covers the major life domains, each conclusion backed by classical references and timing years:

- **Marriage**: when you'll marry, relationship direction, divorce risk, what kind of partner
- **Career**: suitable industry, startup vs employment, career turning years
- **Wealth**: income type, windfall/loss years
- **Health**: vulnerable organs, likely ailments, risk years
- **Study**: education level, exam luck
- **Personality / appearance / family (parents & children) / compatibility (two people)**

Example: `BaZi chart, 1990-05-20 12:00 Beijing, male`; `Write me a full life report`.

### liki-naming — naming

BaZi yongshen fixes the five-element direction → wuge-sancai → candidate characters. Supports baby naming, renaming, company naming, English-to-Chinese naming, and self-picked name review.

Example: `Name my baby, 2024-06-10 Guangzhou, male, surname Chen`

### liki-divination — divination

- **Liuyao**: success/failure and timing ("will this work out", "when will I get the result")
- **Qimen**: direction and timing decisions ("which direction", "should I do it now")
- **Huangli date selection**: pick auspicious days ("which day to move/marry/open a shop")

Example: `Should we move house tomorrow?`

### liki-fengshui — feng shui

- **Bazhai**: ming gua, person-house match, door/master/stove layout
- **Xuankong**: yuan-yun flying stars, wang-shan-wang-xiang, annual feng shui

Example: `How's my home's feng shui?`

## Implementation

**① Charting: astronomical engine** — BaZi/ZiWei charts are computed by the open-source [liki-engine](https://github.com/ml8s/liki-engine): true solar time, DST, lat-lon timezone. The model only interprets; it does not derive chart data itself.

**② Judgments: CSV truth tables** — 46 truth tables (BaZi 26 + ZiWei 20, **590 rules**), each with a **classical reference column** (《渊海子平》《子平真诠》《滴天髓》《三命通会》《紫微斗数全书》 etc.). Plus 7 yearly tables (marriage/kinship/wealth/career/health/study/children, incl. annual star spirits).

**③ Process: fill-in each step** — the root SKILL.md defines the global skeleton (chart → factors → judgments; annual questions follow the liunian chain) and mandatory rules; each domain follows its app card step by step, filling in an 「output: □」 form per step. An unfilled □ blocks the next step; conclusions must trace back to the filled □.

**④ Contextual factors eliminate rule conflicts** — contradicting rules on one chart are resolved with contextual factors:
- **Month-branch master element defines personality** (《子平真诠》) — one primary face; ten-god strength rules are secondary
- **Star-palace co-reference** (《三命通会》) — spouse star exposed + spouse palace clashed = marriage possible but turbulent
- **Annual star spirits** (《协纪辨方书》) — Bingfu/Sangmen/Diaoke/Dahao looked up by Tai-sui year
- **Zero-conflict verification**: scanning 19 domains for contradictory rules → add contextual factors → zero collisions

**⑤ Calibration** — before charting, calibrate the birth hour with known events; before concluding, verify against 3-5 known life periods, conclusion stands only if ≥2 periods match.

## For Developers

BaZi/ZiWei judgments are fully CSV truth-table based (readable, editable, reviewable); divination/feng shui (Liuyao/Qimen/Bazhai/Xuankong) judgments are unified "translation tables" (deterministic computation in the engine, LLM translates by table); the evaluation harness ships with the skill (160 questions, answer-isolated, auto-graded); engine and skill are separated (liki-engine is open source).

### Evaluation: reproducible, data in release posts

Liki maintains an evaluation harness on [MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) (160 questions from fortune-telling contests): answers separated from questions, Docker sandbox isolation (the agent cannot read answers), skill-up auto-grading.

Evaluation data (accuracy/iteration records) lives in release posts and CHANGELOG, not in this README.

Reproduce (eval config and grading scripts ship with the skill):

```bash
bash tests/run-qwen.sh --parallelism 16          # move answers → eval → restore → auto-grade
python3 -m pytest tests/ -q --ignore=tests/test_integration.py   # unit tests (factors/judgments/agent_cli)
```

> The eval configuration (160 questions, answer isolation, auto-grading) is reproducible from this repo; eval data is in release posts and CHANGELOG.

### Architecture

A skill = **documentation layer + tool layer**, plus an external **engine** (astronomical computation). The documentation layer is split into three layers by responsibility:

```
skill ── docs ── root SKILL.md      ← rules (global skeleton + mandatory fill-in + routing + RPC + output/interaction/behavior)
      │       ├─ app/               ← process (per-domain chart → lookup → output, each step 「output: □」+ output template)
      │       └─ domains/<domain>/  ← knowledge (methodology + judgment translation, flat per domain)
      └─ tools ── tools/            ← 5 tools (schema + impl) + judgment csv (46) + factors.csv
external ── engine ── liki-engine   ← open-source JSON-RPC astronomical computation (BaZi/ZiWei/Liuyao/Qimen/feng shui)
```

Call flow: **root SKILL.md reads the app card → the app card calls tools (inside tools/: RPC charting + csv matching) → tools return judgments → interpret via domains/ knowledge → render via the app card's output template**. The csv files are internal tool data (not read by the agent). The tool layer is **optional** — only skills with deterministic-computation needs have it (liki-bazi does; divination/fengshui/naming have no tool layer, using RPC + document translation directly).

### Project Structure

```
repo root (liki-skills engineering area — not installed by npx skills)
├── skills/                     ← 4 independent skills (`npx skills add ml8s/liki` installs all)
│   ├── liki-bazi/              ← metaphysics (BaZi+ZiWei dual-chart)
│   │   ├── SKILL.md            ← rules (process convention + RPC + output/interaction/behavior)
│   │   ├── app/                ← process (9 cards: report/marriage/career/wealth/health/study/personality/family/compatibility)
│   │   ├── domains/            ← knowledge (bazi 16 + ziwei 8 files, flat per domain)
│   │   ├── tools/              ← tools (skill-tools.json + 5 tools + judgment csv + factors)
│   │   │   ├── skill-tools.json ← tool schema (parameters + result_schema)
│   │   │   ├── bazi/           ← BaZi tables 26 (incl. yearly_* 7)
│   │   │   ├── ziwei/          ← ZiWei tables 20
│   │   │   ├── factors/        ← factor definitions (natal + liunian)
│   │   │   ├── paipan.py       ← charting (full_paipan/liunian)
│   │   │   └── duanyu.py       ← rule lookup (query/match)
│   │   └── VERSION / content.sha256  ← version + content fingerprint (self-check)
│   ├── liki-divination/        ← divination (Liuyao/Qimen/Huangli)
│   ├── liki-fengshui/          ← feng shui (Bazhai/Xuankong)
│   └── liki-naming/            ← naming (BaZi yongshen + wuge-sancai)
├── tests/      ← evaluation harness (160 grouped cases, answer separation, grading scripts)
├── scripts/    ← build scripts (build-archive.sh packs 4)
└── webapp/     ← web integration pipeline
```

### Design Principles

- **Single responsibility per layer** — root=rules, app=process, domains=knowledge, tools=tools; no cross-layer leakage.
- **Single data source** — parameters/return fields from rpc.discover (or skill-tools.json result_schema); judgment conclusions from query (csv truth tables); domain docs only write what rpc/csv lack (business mapping, decision chains, constraints, system isolation).
- **Dual-system cross-validation** — BaZi leads, ZiWei reviews; conflicts resolved with explicit evidence.
- **Semantic versioning** — VERSION + CHANGELOG.
- **Evaluation** — independent auto-grading, answer isolation, public data.

## References

This project's design references the following open-source projects:

- [weizeW/mingli-skills](https://github.com/weizeW/mingli-skills) — Four-dimensional cross-validation framework
- [jinchenma94/bazi-skill](https://github.com/jinchenma94/bazi-skill) — Classical summary design, historical event calibration
- [dzcmemory-web/bazi-ziwei-skill](https://github.com/dzcmemory-web/bazi-ziwei-skill) — Bazi+ZiWei cross-validation mode
- [shizhilya/yuan](https://github.com/shizhilya/yuan) — Conclusion-first output design
- [hhszzzz/taibu](https://github.com/hhszzzz/taibu) — Agent-friendly design
- [SylarLong/iztro](https://github.com/SylarLong/iztro) — Ziwei Doushu charting engine
- [ai-freer/fortune-skill](https://github.com/ai-freer/fortune-skill) — Three-layer computation architecture
- [yanouyuan-bit/bazi-roundtable](https://github.com/yanouyuan-bit/bazi-roundtable) — Multi-school review and conclusion strength labeling
- [DestinyLinker/MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) — Metaphysics benchmark reference
- [2021291696/high-confidence-mingli-skill](https://github.com/2021291696/high-confidence-mingli-skill) — Confidence system, personality profile inference

## Links

- Website: [liki.hk](https://liki.hk)
- GitHub: [ml8s/liki](https://github.com/ml8s/liki)

## License

MIT.

## Disclaimer

Liki provides Chinese Astrology analysis from a traditional cultural perspective for research and reference only. Its conclusions **do not constitute** medical diagnosis, legal advice, financial forecasts, or major life decisions. Please maintain a rational and proactive outlook.
