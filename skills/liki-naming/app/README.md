# app — 用户价值层（场景卡索引）

每张卡 = 一类用户问题 → 用户问法路由 + 📖 流程（逐步 + 每步「输出：□」填表）+ 输出模板。
frontmatter 声明 `依赖域`。流程由 app 卡自包含（排盘 → 定用神 → 取字 → 组名 → 评估），根 SKILL.md「流程约定」定全局骨架与强制填表规则。

## 起名全流程

| 卡 | 功能 | 依赖域 |
|----|------|--------|
| [naming.md](naming.md) | **通用起名/改名** | bazi,qiming |
| [foreign.md](foreign.md) | **外国人起中文名**（英文姓→音近中国姓→中文流程） | bazi,qiming |
| [selfcheck.md](selfcheck.md) | **自选名字评估**（跳过排盘取字，直接评估） | qiming |

## 卡间关系

- naming.md 和 foreign.md 共享「排盘→用神→参数→取字→组名→评估」骨架；foreign.md 前置「选姓」一步。selfcheck.md 只执行「评估」。
