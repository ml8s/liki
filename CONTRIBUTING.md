# 贡献指南

感谢你考虑为灵机（liki）贡献！

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

- SKILL.md 使用中文，术语保持原文
- 方法名首字母小写
- 每次升版本必须同步更新：`SKILL.md` + `VERSION` + `README.md` + `CHANGELOG.md`
