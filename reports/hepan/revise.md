你收到了一份灵机合盘报告的审查意见。请根据以下反馈修正报告。

审查意见中提到的 section，请参考建议逐条修改。没提到的保持原样。

修改后的报告必须是一个完整的 JSON 对象，与原始格式相同：
```json
{
  "summary": {
    "fatal": {"title":"","year_chong":"","ri_gong_chong":"","pei_ou_kong":"","gan_ke":"","conclusion":""},
    "matching": {"title":"","rigan":"","fugi_gong":"","wuxing_complement":"","shishen_flow":"","cross_verify":""},
    "sustainability": {"title":"","dayun_sync":"","ziwei_verify":"","key_years":""},
    "rating": {"title":"","level":"","advantages":[],"risks":[],"advice":""}
  },
  "person_a": {
    "info": {"name":"","label":"","shengnian":"","sizhu":"","rizhu":"","yongshen":"","xiyong":"","mingong":"","shengong":"","rating":""},
    "bazi": {"命盘总览":"","五行与十神":"","用神喜忌":"","格局判定":"","夫妻宫":"","大运走势":"","综合建议":""},
    "ziwei": {"命盘总览":"","身宫定位":"","十二宫解读":"","夫妻宫":"","四化分布":"","特殊格局":"","大限":""}
  },
  "person_b": {
    "info": {"name":"","label":"","shengnian":"","sizhu":"","rizhu":"","yongshen":"","xiyong":"","mingong":"","shengong":"","rating":""},
    "bazi": {"命盘总览":"","五行与十神":"","用神喜忌":"","格局判定":"","夫妻宫":"","大运走势":"","综合建议":""},
    "ziwei": {"命盘总览":"","身宫定位":"","十二宫解读":"","夫妻宫":"","四化分布":"","特殊格局":"","大限":""}
  }
}
```

只输出完整的修正后 JSON，不要其他文字。
