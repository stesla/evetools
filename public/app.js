var evetools = {}

evetools.appOnReady = function() {
  fetch('/api/v1/currentUser').then(function(resp) {
    return new Promise((resolve, reject) => {
      if (resp.ok) {
        resp.json().then(resolve);
      } else {
        reject();
      }
    });
  }).then(function(user) {
    evetools.showView(window.location.hash, user);
  }).catch(function() {
    evetools.showView('');
  });
}

evetools.showView = function(hash, user) {
  var viewFn;
  if (user) {
    viewFn = evetools.homeView
  } else {
    viewFn = evetools.loginView
  }
  if (viewFn) {
    let viewContainer = document.querySelector('.view-container');
    while (viewContainer.firstChild) {
      viewContainer.firstChild.remove();
    }
    viewContainer.append(viewFn(user));
  }
  document.querySelector('section.main').classList.remove('hidden');
}

evetools.template = function(name) {
  return document.querySelector('.templates .' + name).cloneNode(true);
}

evetools.loginView = function() {
  return evetools.template('view-login');
}

evetools.homeView = function(user) {
  let view = evetools.template('view-home');
  view.querySelector('.name').textContent = user.characterName;
  view.querySelector('.portrait').src = 'https://imageserver.eveonline.com/Character/' + user.characterID + '_512.jpg';
  return view;
}
