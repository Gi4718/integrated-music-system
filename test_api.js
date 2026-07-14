const http = require('http');
http.get('http://127.0.0.1:3000/song/url?id=405998843', function(res) {
  var d = '';
  res.on('data', function(c) { d += c; });
  res.on('end', function() { console.log(d); });
});
