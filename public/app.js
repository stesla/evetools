var evetools = {}

evetools.appOnReady = function() {
  document.querySelectorAll('.navbar-burger').forEach(el => {
    el.addEventListener('click', () => {
      let targetId = el.dataset.target;
      let target = document.getElementById(targetId);
      el.classList.toggle('is-active');
      target.classList.toggle('is-active');
    });
  });

  fetch('/api/v1/currentUser').then(function(resp) {
    return new Promise((resolve, reject) => {
      if (resp.ok) {
        resp.json().then(resolve);
      } else {
        reject();
      }
    });
  }).then(function(user) {
    let navEnd = evetools.template('navbar-end');
    let avatarUrl = 'https://imageserver.eveonline.com/Character/' + user.characterID + '_32.jpg';
    navEnd.querySelector('.avatar').src = avatarUrl;
    document.querySelector('.navbar-menu').appendChild(navEnd);
    return user
  }).then(function(user) {
    evetools.showView(window.location.hash, user);
  }).catch(function() {
    evetools.showView('');
  }).finally(() => {
    document.querySelector('body').classList.remove('hidden');
  });
}

evetools.showView = function(hash, user) {
  if (!user) return;

  let viewFn = evetools.homeView
  if (viewFn) {
    let viewContainer = document.querySelector('main');
    while (viewContainer.firstChild) {
      viewContainer.firstChild.remove();
    }
    viewContainer.append(viewFn(user));
  }
}

evetools.template = function(name) {
  return document.querySelector('.templates .' + name).cloneNode(true);
}

evetools.homeView = function(user) {
  let view = evetools.template('view-home');
  view.querySelector('.name').textContent = user.characterName;
  return view;
}
