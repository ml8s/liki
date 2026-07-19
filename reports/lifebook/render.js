// lifebook report renderer — pure JS string replacement
// data: { birth, sections: { personality, career, love, health, fortune, milestones } }
window.renderLifeBook = function(data) {
  var report = data.data || data;
  var sections = report.sections || {};

  function esc(s) { return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }

  var sectionKeys = ['personality', 'career', 'love', 'health', 'fortune', 'milestones'];
  var sectionsHtml = '';

  for (var i = 0; i < sectionKeys.length; i++) {
    var key = sectionKeys[i];
    var sec = sections[key];
    if (!sec) continue;

    var phasesHtml = '';
    if (key === 'fortune' && sec.phases && sec.phases.length) {
      var rows = '';
      for (var j = 0; j < sec.phases.length; j++) {
        var p = sec.phases[j];
        rows += '<tr><td class="phase-age">' + esc(p.age) + '</td><td class="phase-pillar">' + esc(p.pillar) + '</td><td class="phase-analysis">' + esc(p.analysis) + '</td><td class="phase-advice">' + esc(p.advice) + '</td></tr>';
      }
      phasesHtml = '<table class="phases-table"><thead><tr><th>年龄</th><th>干支</th><th>分析</th><th>建议</th></tr></thead><tbody>' + rows + '</tbody></table>';
    }

    sectionsHtml += '<div class="section" data-section="' + esc(key) + '">' +
      '<h2 class="section-title">' + esc(sec.title) + '</h2>' +
      (sec.data ? '<div class="section-data">' + esc(sec.data) + '</div>' : '') +
      (sec.analysis ? '<div class="section-analysis">' + esc(sec.analysis) + '</div>' : '') +
      (sec.advice ? '<div class="section-advice">' + esc(sec.advice) + '</div>' : '') +
      phasesHtml +
      '</div>';
  }

  return '<div class="report-container">' +
    '<h1 class="report-title">' + esc(report.birth || '') + ' 的生命之书</h1>' +
    sectionsHtml +
    '</div>';
};
