# Release model

Liki uses two coordinated version dimensions. They serve different purposes and must not be merged into one version string.

## 1. Runtime / engineering version: CalVer

The executable contract uses a date-stamped version:

```text
YYYY.MM.DD.SERIAL
```

Example:

```text
2026.09.15.7
```

Sources:

- `skills/liki/VERSION.txt`
- `engine/cmd/engine-mcp/VERSION`
- `counsel/app/VERSION.txt`
- `engine-mcp` / `counsel-mcp` `serverInfo.version`
- counsel tool/projection manifests
- skill version self-check
- compatibility gates

Rules:

- `make version` writes the same CalVer to the unified skill and engine version files.
- The serial resets to `0` each local date and increments for further releases on that date.
- CalVer is the engineering / compatibility version. It changes for normal development, fixes, docs and feature work.
- Tests and runtime compatibility require this value to remain dotted-integer CalVer. Do not replace it with SemVer.

## 2. Product release: SemVer tag

Formal product releases use a SemVer tag:

```text
vMAJOR.MINOR.PATCH
```

Examples:

```text
v5.0.0
v5.1.0
v5.0.1
```

SemVer is the public release identity. It is not stored in `VERSION.txt`; it is represented by an annotated Git tag and GitHub Release. For SkillHub distribution, the same release SemVer is mirrored to `skills/liki/SKILL.md` frontmatter `version`.

Version-surface rules:

| Surface | Version | Update cadence |
|---|---|---|
| Runtime / compatibility | CalVer in `VERSION.txt` and engine `VERSION` | Every engineering bump by `make version` |
| Product release | SemVer Git tag and GitHub Release | Only on an intentional release |
| SkillHub package | `skills/liki/SKILL.md` frontmatter `version` | Only when preparing a SkillHub release; do not bump it during normal CalVer changes |

Semantics:

| Change | Version |
|---|---|
| Breaking product, skill layout, RPC contract, output contract, or removal of public entry points | major |
| Backward-compatible new domain, tool, RPC, report, or product capability | minor |
| Bug fix, docs, small correction without breaking contracts | patch |

## 2.1 跨仓依赖 pin（独立发版）

各仓（liki / liki-agents / liki-deploy / liki-web）**独立发版、独立更新**，互不要求版本同名对齐。liki 通过显式 pin 锁定其跨仓依赖：

| 依赖 | pin 文件 | 锁定内容 | 更新时机 |
|---|---|---|---|
| AgentDeployment schema | `contracts/agent-definition.version` | schema version + sha256 | liki-agents 发布新 schema 后 |
| 装配运行时 base | `contracts/liki-agents.base` | liki-agents version + immutable digest | liki-agents 发布新版且本仓验证通过后 |

- 原则：**测试用哪个版本，部署就用哪个**——`make agents-validate`（本地 / CI）与 CI 装配 `FROM` 共用同一 pin（`python3 scripts/liki_agents_pin.py`）。
- liki-agents 稳定后很少更新；**未发新版则 pin 保持不变**，不随 liki 自己的版本自动对齐。
- CI 装配从 pin 取 `liki-agents:<version>` 并校验解析出的 immutable digest == pin.digest（防上游 tag 移动），不再依赖发布 tag 名。
- 校验已接入 `make check`（`python3 scripts/liki_agents_pin.py --check`）。

## 3. Historical SemVer baselines

Before `v5.0.0`, major architecture milestones were retro-tagged so public history has stable SemVer anchors:

| Tag | Commit | Date | Meaning |
|---|---|---|---|
| `v1.0.0` | `c2dd692` | 2026-07-16 | Historical prompt-based Liki Skills baseline. |
| `v2.0.0` | `2713559` | 2026-07-29 | Domain and application separation. |
| `v3.0.0` | `a0884cf` | 2026-08-16 | Four-skill split, flat domains, and unified versioning. |
| `v4.0.0` | `0743ce8` | 2026-08-25 | Liki Engine merged into the monorepo. |

### Final minor tag per historical major

For complete compatibility anchors, the highest minor of each historical major is also tagged:

| Tag | Commit | Date | Meaning |
|---|---|---|---|
| `v1.40.0` | `359fd54` | 2026-07-29 | Final 1.x prompt-skill release before the 2.x domain/app split. |
| `v2.4.0` | `90b8656` | 2026-08-07 | Final 2.x domain/app two-layer release. |
| `v3.10.2` | `fad44b0` | 2026-08-15 | Final 3.x skill-tools architecture release. |
| `v4.3.1` | `0d24031` | 2026-08-19 | Final four-skill split release before engine monorepo integration. |

These are historical anchors for the pre-unified architecture.

The last independent four-skill snapshot was `1039ec5`, where every old skill `VERSION` was synchronized to `2026.09.15.5`. Scoped tags:

| Tag | Skill |
|---|---|
| `bazi-v2026.09.15.5` | `liki-bazi` |
| `divination-v2026.09.15.5` | `liki-divination` |
| `fengshui-v2026.09.15.5` | `liki-fengshui` |
| `naming-v2026.09.15.5` | `liki-naming` |

They mark the final state immediately before consolidation into unified `liki`.

## 4. v5.0.0 baseline

`v5.0.0` is the first unified product release using the single-skill architecture.

Baseline decision:

- Product release: `v5.0.0`
- Runtime CalVer: `2026.09.16.0`
- Baseline commit: `6f407df`
- Release theme: unified `liki` skill, ready-to-use payload contracts, advisory safety, and cross-domain deterministic golden tests.

Breaking scope included:

- removal of the four old skill entry points;
- consolidation into one installable `liki` skill;
- unified `ENTRY.md` routing;
- `TOOLS.md` payload contracts;
- replacement of divination hard blocking with `safety_advisory`;
- explicit fail-closed discover scope and required-method checks.

## 5. Release identity

Use both identifiers in release material:

```text
Liki v5.0.0
Runtime CalVer: 2026.09.16.0
Commit: 6f407df
```

Recommended Git command:

```bash
git tag -a v5.0.0 6f407df -m "Liki 5.0.0 — Unified Chinese Metaphysics Skill Platform"
git push origin v5.0.0
```

The GitHub Release should use the same tag and include the CalVer and commit in its notes.

## 6. Release decision rules

A normal documentation or bug-fix push does not automatically create a SemVer release.

Create `vX.Y.Z` only when the main branch is green and the milestone is intentionally released.

- Breaking public behaviour, payload contract, skill layout, RPC removal, or old-entry removal → major.
- New domain, tool, RPC, report, or backward-compatible capability → minor.
- Fixes or corrections without breaking contracts → patch.

If breaking work lands on main after `v5.0.0`, the next release must be `v6.0.0`; do not hide breaking changes in `v5.1.0`.

## 7. Release process

1. Confirm main is green.
2. Confirm `skills/liki/VERSION.txt` and `engine/cmd/engine-mcp/VERSION` are the intended runtime CalVer.
3. If this commit is also a SkillHub release, confirm `skills/liki/SKILL.md` frontmatter `version` equals the SemVer tag. Do not change it for routine CalVer bumps.
4. Confirm `CHANGELOG.md` contains the milestone entries.
5. Create the annotated SemVer tag on the approved commit.
6. Push the tag.
7. Create the GitHub Release with both versions and release notes.
8. Verify release artifacts and CI.

## 8. Quality gates

Make targets are layered: lint → check → test → verify → gate.

```bash
# Layer 1: Lint（格式 & 风格，< 30s）
make lint                 # markdownlint + golangci-lint + go vet + ruff
make fmt-check            # Go format check

# Layer 2: Check（结构 / 契约 / 数据质量，< 30s）
make check                # format/vet + skill structure + schema + docs

# Layer 3: Test（逻辑正确性）
make test                 # contracts + engine + counsel + engine MCP smoke

# Layer 4: Verify（当前 MCP 协议 / 服务集成）
make verify               # engine MCP smoke + counsel integration

# Layer 5: Gate（发布门槛）
make gate                # engine lint + check + test + 160 题覆盖 + archive
```

Release checks add golden and package:

```bash
make golden
make build-archive
```

Model- or local-tool-backed checks are release evidence, not gate requirements:

```bash
# Legacy model-backed suites are being migrated to the current MCP skill surface.
```

TRACE is an external static Skill-quality review. It complements, but does not replace, deterministic golden tests, domain oracles, functional contracts, MCP integration tests, or archived model-backed evaluations.
