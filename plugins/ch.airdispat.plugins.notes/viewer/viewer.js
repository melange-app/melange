document.addEventListener('DOMContentLoaded', function() {
  melange.viewer(function(d) {
    var template = document.getElementById('template').text;
    for(var i in d.components) {
      var s = i.split("/")
      var p = s[s.length - 1];
      d.components[p] = d.components[i];
      delete d.components[i];
    }
    var rendered = Mustache.render(template, d);
    document.getElementById('container').innerHTML = rendered;
  });
});
