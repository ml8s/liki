# 贡献指南

感谢你考虑为liki.hk贡献！

## 提交 Issue

- 命理计算结果不准确 → 请附上出生时间、地点、期望结果和实际结果
- SKILL 行为异常 → 请说明触发了什么流程、期望什么、实际发生了什么
- 建议新技能 → 请描述用户场景，越具体越好

## 提交 PR

1. Fork 本仓库
2. 创建一个功能分支：`git checkout -b feat/my-change`
3. 修改 SKILL.md 后更新根 `SKILL.md` 的 `version` 字段（小 bug 修 patch，流程变更升 minor）
4. 同步更新 `VERSION` 文件和 `CHANGELOG.md`
5. 提交 PR，描述清楚改了什么、为什么

## 代码规范

- SKILL.md 以中文为主，术语保持原文；不写方法名和参数（由引擎 schema 驱动）
- 每次升版本必须同步更新：`SKILL.md` + `VERSION` + `README.md` + `CHANGELOG.md`

## 推送前检查清单

**改方法名/函数名时**（全量搜索所有引用点）：
```bash
# 搜代码（含脚本）
grep -rn "旧方法名" --include="*.go" --include="*.sh" --include="*.py"
# 搜 skill 文档
grep -rn "旧方法名" skills/*/app/*.md skills/*/SKILL.md
```

**改 skill 文档后**（重算指纹）：
```bash
make build-archive
```

**推送前本地 CI**：
```bash
# skill 校验
for s in liki-bazi liki-divination liki-fengshui liki-naming; do
  python3 tests/check_docs.py "skills/$s"
done
```
