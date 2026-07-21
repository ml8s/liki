# 本地特性

**本节仅适用于具备文件写入能力的 AI 客户端。无写入权限的环境跳过本节。**

---

## 记忆管理

启动时检测当前目录是否有 `liki-memory.json`。有则问"用上次的命盘存档？(y/n)"，Yes 跳过收集→排盘→用神步骤直接进入分析。

八字排盘+用神完成后，问"保存命盘以便后续使用？(y/n)"。Yes → 写入 `liki-memory.json`。

文件格式：`{"birth":"1984-02-15 08:00 | 上海 | 男","chart":{...},"yongshen":{...}}`
存储 `bazi.chart`、`bazi.yongshen`、`ziwei.chart`、`bazi.hehui`、`bazhai.chart`、`bazhai.minggua`、`bazhai.judgment`、`xuankong.chart` 的 data 全量。不存随时间变化的结果（流年/流月/六爻/奇门/黄历）。

首次保存时提醒："出生信息将保存在当前目录的 `liki-memory.json` 中。请勿分享或在公开仓库提交。"
帮他人排盘时不主动提议存档。

---

## 报告渲染

`reports/lifebook/lifebook.json` 生成后，用 Node.js 渲染为 HTML：

```js
const fs = require('fs');
global.window = {};
eval(fs.readFileSync('reports/lifebook/render.js', 'utf8'));
const data = JSON.parse(fs.readFileSync('reports/lifebook/lifebook.json','utf8'));
const html = '<!DOCTYPE html><html lang="zh-Hans"><head><meta charset="UTF-8"><link rel="stylesheet" href="https://liki.hk/css/report.css"></head><body>' +
  window.renderLifeBook(data) +
  '</body></html>';
fs.writeFileSync('reports/lifebook/output.html', html);
```

或直接拖 `lifebook.json` + `render.js` 到浏览器打开。
