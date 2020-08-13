var fixture = fetch('/index.html')
  .then(response => response.text())
  .then(function(text) {
    let parser = new DOMParser();
    return parser.parseFromString(text, 'text/html')
  })
  .then(function(html) {
    let div = document.createElement('div');
    div.classList.add('fixture');
    div.setAttribute('style', 'display: none;');
    div.appendChild(html.querySelector('main'));
    div.appendChild(html.querySelector('section.templates'));
    let body = document.querySelector('body');
    return function() {
      let oldNode = body.querySelector('.fixture');
      let newNode = div.cloneNode('true');
      if (oldNode) {
        body.replaceChild(newNode, oldNode);
      } else {
        body.append(newNode);
      }
    };
  });

beforeEach(function(done) {
  fixture.then(function(reload) {
    reload();
  }).then(done, fail);
});
