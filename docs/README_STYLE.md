# README style guide

This guide defines the stable structure and formatting contract for `README.md`, `README.en.md`, and the user guides. The goal is a clean entry document: a reader should understand what Liki is, install it, and try one example in under two minutes.

## Canonical structure

`README.md` uses these H2 sections in this order:

1. 安装
2. 快速开始
3. 能力概览
4. 可信度与边界
5. 常见问题
6. 文档
7. 开发者
8. 贡献
9. 许可与声明

`README.en.md` uses these H2 sections in the same order:

1. Install
2. Quick start
3. What you get
4. Trust and boundaries
5. FAQ
6. Documentation
7. For developers
8. Contributing
9. License and disclaimer

Detailed usage belongs in `docs/USER_GUIDE.md` and `docs/USER_GUIDE.en.md`, not in the README.

## Heading rules

- Use exactly one visible H1 for the product name.
- HTML is allowed only for the centered README banner.
- Use H2 for canonical reader tasks.
- Use H3 for subtasks inside a canonical H2 section.
- Do not use H4 in a README.
- Do not skip heading levels.
- Do not end a declarative heading with punctuation; FAQ headings may be questions and may end with `?`.

### English casing

Use sentence case for headings:

- `Install`
- `Quick start`
- `What you get`
- `Trust and boundaries`
- `For developers`
- `Domain contracts`
- `Engine image`
- `Test commands`

Capitalize proper nouns, product names, and established technical acronyms only.

## Emphasis rules

Do not use a standalone bold line as a heading.

Bad:

```md
**Don't know the exact birth hour?**
```

Good:

```md
### Don't know the exact birth hour?
```

Bold is allowed inside a sentence for a short emphasis, but it must not replace structure.

## Examples and tables

Use a table when there are parallel examples or parallel domains. Put user-facing examples in inline code:

```md
`算八字，1990-05-20 12:00 北京出生，男`
```

Use fenced code blocks with a language for commands:

```bash
npx skills add ml8s/liki
```

Do not mix commands, examples, and explanations into one dense paragraph.

## Language and terminology

- Keep one space between CJK text and Latin words, for example `专业命理 Skill`.
- Use Chinese full-width punctuation in Chinese prose and ASCII punctuation in English prose.
- Capitalize product and technical terms consistently: `Liki`, `Skill`, `RPC`, `JSON-RPC`, `App 卡`, `CalVer`, `SemVer`.
- Use the canonical domain spellings from the [brand glossary](./brand.md): `Bazi`, `Ziwei`, `Liuyao`, `QiMen`, `Feng Shui`, `Date Selection`, `Naming`.
- Use lower-case for file and directory names exactly as they appear: `skills/liki`, `ENTRY.md`, `TOOLS.md`, `RPC.md`.
- Use `Go engine`, not `go Engine`.
- Use the canonical slogan in the banner and one closing callout only; do not repeat it throughout every body section.

## Links and paths

- Prefer relative links for repository files.
- Link to a specialized document instead of duplicating its details.
- Every linked local path must exist.
- Documentation tables should use document name, purpose, and relative link.
- Required README links are: the language-matched user guide, `docs/SKILL_PACKAGE.md`, `docs/README_STYLE.md`, `CONTRIBUTING.md`, and `CHANGELOG.md`.

## Length and density

- Keep READMEs below 180 lines and prefer editor soft wrapping over hard line breaks.
- Prefer a table over four sibling H4 sections.
- One section should answer one reader task.
- If a section needs more than a few short examples, move it to the user guide.

## Automated checks

Run:

```bash
make lint-readme
python3 -m pytest tests/test_readme_contract.py -q
```

CI also runs `markdownlint-cli2` for README and user-guide files.
