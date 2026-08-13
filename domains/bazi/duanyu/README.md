# bazi duanyu（已收敛）

本目录原为八字断语细则 md，已按三类收敛（迁移映射见 git 历史）：

- **断语（单一来源）**：`tools/bazi/*.csv`——真值表，机器执行，含 id+约束+结论+依据+经典原文
- **方法（断法细则）**：`domains/bazi/fangfa/*.md`——判断链/护栏/取象次序/边界澄清（csv 表达不了的"如何断"）

| 原 duanyu md | 收敛去向 |
|---|---|
| caiyun.md | 静态表删（财星类型/量级/正偏财/①②③）→ fangfa/caiyun.md（判断链/位置维度/破财护栏） |
| hehui.md | A 条删（流年推断 9 条 csv 已覆盖）→ fangfa/hehui.md（合冲刑害基础+6 条方法） |
| shishen.md | 整体迁 fangfa/shishen.md（场景表/强度换算/取清规则/制化优先） |
| shiye.md | 4 档位表删 → fangfa/shiye.md（升格护栏/取象护栏） |
| wuxing-jiankang.md | 脏腑/失衡表删 → fangfa/wuxing-jiankang.md（双向映射/特殊组合） |
| xueye.md | ②③④表删 → fangfa/xueye.md（前置/格局主导/护栏） |

断语以 csv 为准；判题方法看 fangfa；两者冲突以 csv 为准。
