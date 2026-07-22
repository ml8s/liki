// lifebook report renderer — data → HTML
// data: { birth, sections: { personality, career, love, health, fortune, milestones } }
window.renderLifeBook = function(data) {
  var report = data.data || data;
  var sections = report.sections || {};
  function esc(s) { return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }

  var keys = ['personality', 'career', 'love', 'health', 'fortune', 'milestones'];
  var sectionsHtml = '';
  for (var i = 0; i < keys.length; i++) {
    var sec = sections[keys[i]];
    if (!sec) continue;
    var phasesHtml = '';
    if (keys[i] === 'fortune' && sec.phases && sec.phases.length) {
      var rows = '';
      for (var j = 0; j < sec.phases.length; j++) {
        var p = sec.phases[j];
        rows += '<tr><td class="age">' + esc(p.age) + '</td><td class="pillar">' + esc(p.pillar) + '</td><td>' + esc(p.analysis) + '</td></tr>';
      }
      phasesHtml = '<div class="report-phases"><table>' + rows + '</table></div>';
    }
    sectionsHtml += '<div class="report-section">' +
      '<div class="report-section-title">' + esc(sec.title) + '</div>' +
      (sec.analysis ? '<div class="report-analysis">' + esc(sec.analysis) + '</div>' : '') +
      (sec.advice ? '<div class="report-advice">💡 ' + esc(sec.advice) + '</div>' : '') +
      phasesHtml +
      '</div>';
  }

  return '<div class="report">' +
    '<h1>' + esc(report.birth || '') + '</h1>' +
    sectionsHtml +
    '</div>';
};
