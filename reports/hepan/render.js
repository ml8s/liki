// hepan report renderer — data → HTML
// data: { title, person_a: { name }, person_b: { name }, sections: [{ title, analysis, rating }] }
window.renderHePan = function(data) {
  var report = data.data || data;
  var sections = report.sections || [];
  function esc(s) { return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }
  function ratingClass(r) {
    if (r === '上等' || r === '上') return 'report-rating report-rating-high';
    if (r === '中') return 'report-rating report-rating-mid';
    return 'report-rating report-rating-low';
  }

  var personsHtml = '';
  if (report.person_a) personsHtml += '<div class="report-person-card"><div class="label">甲方</div><div class="name">' + esc(report.person_a.name) + '</div></div>';
  if (report.person_b) personsHtml += '<div class="report-person-card"><div class="label">乙方</div><div class="name">' + esc(report.person_b.name) + '</div></div>';

  var sectionsHtml = '';
  for (var i = 0; i < sections.length; i++) {
    var s = sections[i];
    sectionsHtml += '<div class="report-section">' +
      '<div class="report-section-title">' + esc(s.title) + '</div>' +
      (s.analysis ? '<div class="report-analysis">' + esc(s.analysis) + '</div>' : '') +
      (s.rating ? '<span class="' + ratingClass(s.rating) + '">' + esc(s.rating) + '</span>' : '') +
      '</div>';
  }

  return '<div class="report">' +
    '<h1>' + esc(report.title || '') + '</h1>' +
    (personsHtml ? '<div class="report-persons">' + personsHtml + '</div>' : '') +
    sectionsHtml +
    '</div>';
};
