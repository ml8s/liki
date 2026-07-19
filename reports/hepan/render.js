// hepan report renderer — pure JS string replacement
// data: { title, person_a: { name }, person_b: { name }, sections: [{ title, analysis, rating }] }
window.renderHePan = function(data) {
  var report = data.data || data;
  var sections = report.sections || [];

  function esc(s) { return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }

  var personsHtml = '';
  if (report.person_a) {
    personsHtml += '<div class="card card-paper p-3 text-sm"><span class="font-bold text-stone-700">甲方</span><br><span class="text-stone-500">' + esc(report.person_a.name) + '</span></div>';
  }
  if (report.person_b) {
    personsHtml += '<div class="card card-paper p-3 text-sm"><span class="font-bold text-stone-700">乙方</span><br><span class="text-stone-500">' + esc(report.person_b.name) + '</span></div>';
  }

  var sectionsHtml = '';
  for (var i = 0; i < sections.length; i++) {
    var s = sections[i];
    var ratingHtml = '';
    if (s.rating) {
      var cls;
      if (s.rating === '上等' || s.rating === '上') cls = 'bg-green-50 text-green-700';
      else if (s.rating === '中') cls = 'bg-amber-50 text-amber-700';
      else cls = 'bg-stone-50 text-stone-500';
      ratingHtml = '<span class="inline-block mt-3 text-xs px-2 py-1 rounded ' + cls + '">' + esc(s.rating) + '</span>';
    }
    sectionsHtml += '<section class="card card-paper p-4 mb-4">' +
      '<h2 class="font-bold text-stone-800 mb-2">' + esc(s.title) + '</h2>' +
      (s.analysis ? '<div class="text-stone-600 text-sm leading-relaxed whitespace-pre-line">' + esc(s.analysis) + '</div>' : '') +
      ratingHtml +
      '</section>';
  }

  return '<div class="report">' +
    '<div class="max-w-2xl mx-auto px-4 py-8">' +
    '<h1 class="text-2xl font-bold text-stone-800 mb-4">' + esc(report.title) + '</h1>' +
    (personsHtml ? '<div class="grid grid-cols-2 gap-4 mb-6">' + personsHtml + '</div>' : '') +
    sectionsHtml +
    '</div></div>';
};
