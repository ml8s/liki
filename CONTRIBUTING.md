# 贡献指南

感谢你考虑为灵机（liki）贡献！

## 提交 Issue

- 命理计算结果不准确 → 请附上出生时间、地点、期望结果和实际结果
- SKILL 行为异常 → 请说明触发了什么流程、期望什么、实际发生了什么
- 建议新技能 → 请描述用户场景，越具体越好

## 提交 PR

1. Fork 本仓库
2. 创建一个功能分支：`git checkout -b feat/my-change`
3. 修改 SKILL.md 后更新对应 `version` 字段（小 bug 修 patch，流程变更升 minor）
4. 更新 `CHANGELOG.md`
5. 提交 PR，描述清楚改了什么、为什么

## 代码规范

- SKILL.md 使用中文，术语保持原文
- API method 名首字母小写
- 参数收集表头统一为 `产品 | 所需信息`
- 功能方法表头统一为 `功能 | Method`
