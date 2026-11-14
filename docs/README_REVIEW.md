# README 评审清单

改 README 前后人工过一遍本清单；不进 CI，仅作为评审会话与大改版的检查工具。

## 文化差异原则（先读）

- EN 与 ZH 两份 README **不做互译**：各自从目标市场的文化特点独立撰写，品牌件（H1、logo、徽章、双语 slogan）保持一致。
- EN 面向全球英文用户：中国玄学兴趣人群（birth chart、astrology reading 语境）与 AI 工具用户（MCP、Claude、Codex、Cursor 语境）。
- ZH 面向中文客户：命理传统话语体系（八字、紫微、考时、真太阳时、节气）与中文互联网语境。
- 热点词分市场维护：英文社区关注 birth chart、manifestation、retrograde 等词的迁移；中文社区关注 AI 算命、电子烧香、赛博命理等词的演变。每季度评审时核对示例与 FAQ 用词是否仍贴近当前热词。
- 每次评审先检查：两份是否仍为独立叙事，有没有出现互译痕迹（句式一一对应、示例雷同、FAQ 雷同）。

## 六维度清单

### 品牌合规（红线项）

- [ ] H1、slogan、描述用语与 liki-web `BRAND.md` 逐字一致
- [ ] 禁用表述零出现：降格表述、slogan 变体、旧品类词 Metaphysics
- [ ] 语言纯净：EN 面无中文 slogan；ZH 面不直译英文文案

### 首屏转化（10 秒测试）

- [ ] 前 3 屏回答三个问题：这是什么 / 对我有什么用 / 怎么开始
- [ ] 安装命令在一屏内可见，复制即可执行
- [ ] 示例表能让目标用户对号入座（看到自己的问题）

### SEO（搜索引擎）

- [ ] 品类关键词出现在副标题、章节标题与正文首段
- [ ] repo description 与 GitHub topics 和 README 关键词呼应
- [ ] 标题携带语义关键词，拒绝无信息标题

### GEO（AI 引擎优化）

- [ ] 存在 "X is a ..." 开头的可摘录定义段
- [ ] 平台实体（Claude / Codex / Cursor / MCP）在正文有明确段落
- [ ] 独特数字事实在场（829 断语、192 案例、SHA-256）
- [ ] FAQ 句式贴近目标市场的真实提问方式

### 事实与合规

- [ ] 数字与代码一致（断语数由 `scripts/check_docs.py` 钉住，其余人工核对）
- [ ] 免责声明完整：传统文化视角、非专业建议、Mingli160 非准确率证明
- [ ] 无过度承诺（不出现绝对化准确率表述）

### 工程质量

- [ ] CI 基线绿：契约测试、markdownlint、check_docs
- [ ] 移动端渲染可读，表格不超宽
- [ ] 两份行数均在 180 行预算内

## GEO 引擎实测（每季度一次，人工）

在 Perplexity / ChatGPT / Claude 用目标问题实测引用情况，记录是否被引用、引用句来自 README 哪一段：

- EN：Which open-source skill can do BaZi reading in Claude?
- ZH：有没有能在 Claude 里算八字的开源 skill
