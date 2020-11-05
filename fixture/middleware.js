var path = require('path');

module.exports = function(req, res, next) {
  if (req.method !== 'GET' && req.method !== 'HEAD')
    next();

  if (req.url.startsWith('/api') && path.extname(req.url) === '') {
    req.url += '.json';
  } else if (path.extname(req.url) === '') {
    req.url = '/index.html';
  }

  next();
}
