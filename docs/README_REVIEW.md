# README review checklist

Run this checklist manually before and after README changes. It is not wired
into CI — it exists for review sessions and major rewrites.

## Cultural differentiation principle (read first)

- The English and Chinese READMEs are **never translations of each other**:
  each is written independently for its target market. Brand elements (H1,
  logo, badges, per-language slogans) stay identical.
- The English face targets global English-speaking users: the Chinese
  metaphysics curiosity segment (birth chart / astrology reading context) and
  AI-tool users (MCP, Claude, Codex, Cursor context).
- The Chinese face targets Chinese customers: the traditional fortune-telling
  vocabulary (八字, 紫微, 考时, 真太阳时, 节气) and the Chinese internet context.
- Trending words are maintained per market: watch migration of English terms
  (birth chart, manifestation, retrograde) and evolution of Chinese terms
  (AI 算命, 赛博命理). Re-verify example phrasing and FAQ wording each quarter.
- At every review, first check that the two faces are still independent
  narratives — no mirrored sentence structures, no duplicated examples, no
  duplicated FAQ items.

## Six-dimension checklist

### Brand compliance (red lines)

- [ ] H1, slogans, and descriptor wording match the liki-web `BRAND.md` verbatim
- [ ] Zero forbidden phrasing: downgrading expressions, slogan variants, retired category word "Metaphysics"
- [ ] Language purity: no Chinese slogan on the English face; the Chinese face never translates English copy

### First-screen conversion (10-second test)

- [ ] The first three screens answer: what is this / what is in it for me / how do I start
- [ ] The install command is visible within one screen and copy-paste runnable
- [ ] Example tables let target users recognize their own question

### SEO (search engines)

- [ ] Category keywords appear in the subtitle, section headings, and first paragraph
- [ ] Repo description and GitHub topics echo the README keywords
- [ ] Headings carry semantic keywords — no information-free headings

### GEO (generative engine optimization)

- [ ] An extractable "X is a ..." definition paragraph exists
- [ ] Platform entities (Claude / Codex / Cursor / MCP) have a dedicated paragraph in the body
- [ ] Distinctive numeric facts are present (829 assertions, 192 cases, SHA-256)
- [ ] FAQ phrasing mirrors how real users ask in each market

### Facts and compliance

- [ ] Numbers match the code (assertion count is pinned by `scripts/check_docs.py`; verify the rest manually)
- [ ] Disclaimers are complete: traditional-cultural perspective, not professional advice, Mingli160 is not an accuracy benchmark
- [ ] No over-claiming (no absolute accuracy statements)

### Engineering quality

- [ ] CI baseline green: contract tests, markdownlint, check_docs
- [ ] Mobile rendering readable, tables not over-wide
- [ ] Both files stay within the 180-line budget

## GEO engine field test (quarterly, manual)

Query Perplexity / ChatGPT / Claude with target questions and record whether
we are cited, and which README paragraph the citation came from:

- EN: Which open-source skill can do BaZi reading in Claude?
- ZH: 有没有能在 Claude 里算八字的开源 skill
