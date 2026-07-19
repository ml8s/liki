你收到了一份报告的审查意见。请根据以下反馈修正报告。

审查意见中提到的每个 section，请参考建议逐条修改。没提到的 section 保持原样。

修改后的报告必须是一个完整的 JSON 对象，与原始格式相同：
```json
{
  "sections": {
    "personality": { "title": "性格画像", "data": "...", "analysis": "...", "advice": "..." },
    "career": { "title": "事业财运", "data": "...", "analysis": "...", "advice": "..." },
    "love": { "title": "情感婚姻", "data": "...", "analysis": "...", "advice": "..." },
    "health": { "title": "健康提示", "data": "...", "analysis": "...", "advice": "..." },
    "fortune": { "title": "十年大运", "data": "...", "analysis": "...", "advice": "...",
      "phases": [{"age": "", "pillar": "", "analysis": "", "advice": ""}]
    },
    "milestones": { "title": "关键提醒", "data": "...", "analysis": "...", "advice": "..." }
  }
}
```

只输出完整的修正后 JSON，不要其他文字。不要问我任何问题。
