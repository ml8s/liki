你收到了一份报告的审查意见。请根据以下反馈修正报告。

审查意见中提到的每个 section，请参考建议逐条修改。没提到的 section 保持原样。

修改后的报告必须是一个完整的 JSON 对象，与原始格式相同：
```json
{
  "summary": {
    "personality": {"title":"","bazi":"","ziwei":"","cross":"","advice":""},
    "career": {"title":"","bazi":"","ziwei":"","cross":"","window":"","advice":""},
    "love": {"title":"","bazi":"","ziwei":"","cross":"","advice":""},
    "health": {"title":"","bazi":"","ziwei":"","cross":"","advice":""},
    "fortune": {"title":"","phases":[],"daxian":[],"liunian":[],"advice":""},
    "milestones": {"title":"","items":[],"advice":""}
  },
  "bazi": {
    "pan": {"sizhu":"","rizhu":"","shenqiang":"","geju":"","yongshen":"","xishen":"","jishen":""},
    "sections": { "命盘总览":"","五行与十神分析":"","用神喜忌":"","格局判定":"","大运走势":"","当前流年":"","综合建议":"" }
  },
  "ziwei": {
    "pan": {"mingong":"","shengong":"","rating":"","patterns":[]},
    "sections": { "命盘总览":"","身宫定位":"","十二宫逐宫解读":"","四化分布":"","三方四正":"","特殊格局":"","大限":"","流年":"" }
  }
}
```

只输出完整的修正后 JSON，不要其他文字。不要问我任何问题。
