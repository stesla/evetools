var fixture = fetch('/index.html')
  .then(response => response.text())
  .then(function(text) {
    let parser = new DOMParser();
    return parser.parseFromString(text, 'text/html')
  })
  .then(html => html.querySelector('section.main.container'))
  .then(function(main) {
    let div = document.createElement('div');
    div.classList.add('fixture');
    div.setAttribute('style', 'display: none;');
    div.appendChild(main);
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
