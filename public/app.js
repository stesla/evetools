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
  console.log(user);
  if (user) {
    viewFn = evetools.profileView
  } else {
    viewFn = evetools.landingView
  }
  if (viewFn) {
    let viewContainer = document.querySelector('.view-container');
    while (viewContainer.firstChild) {
      viewContainer.firstChild.remove();
    }
    viewContainer.append(viewFn(user));
  }
}

evetools.template = function(name) {
  return document.querySelector('.templates .' + name).cloneNode(true);
}

evetools.landingView = function() {
  return evetools.template('landing-view');
}

evetools.profileView = function(user) {
  let view = evetools.template('profile-view');
  view.querySelector('.name').innerText = user.characterName;
  return view;
}
