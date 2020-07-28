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
    document.querySelector('body').append(div);
    return div
  });

beforeEach(function(done) {
  fixture.then(function(div) {
    let body = document.querySelector('body');
    let oldNode = body.querySelector('.fixture');
    let newNode = div.cloneNode(true);
    body.replaceChild(newNode, oldNode);
  }).then(done, fail);
});
