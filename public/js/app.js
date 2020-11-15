evetools = (function(document, window, undefined) {
  var result = {};

  var currentUser = window.retrieve('/api/v1/user/current', 'error fetching current user');
  result.currentUser = currentUser;

  var _sdeTypes
  function sdeTypes() {
    if (!_sdeTypes)
      _sdeTypes = retrieve('/data/types.json', 'error fetching sde types');
    return _sdeTypes
  }
  result.sdeTypes = sdeTypes;

  var _sdeMarketGroups
  function sdeMarketGroups() {
    if (!_sdeMarketGroups)
      _sdeMarketGroups = retrieve('/data/marketGroups.json', 'error fetching sde market groups');
    return _sdeMarketGroups;
  }
  result.sdeMarketGroups = sdeMarketGroups;

  var _sdeStations
  function sdeStations() {
    if (!_sdeStations)
      _sdeStations = retrieve('/data/stations.json', 'error fetching sde stations');
    return _sdeStations;
  }
  result.sdeStations = sdeStations;

  var _sdeSystems
  function sdeSystems() {
    if (!_sdeSystems)
      _sdeSystems = retrieve('/data/systems.json', 'error fetching sde systems');
    return _sdeSystems;
  }
  result.sdeSystems = sdeSystems;

  result.globalState = function() {
    return {
      avatarMenuOpen: false,
      marketMenuOpen: false,
      loggedIn: false,
      navOpen: false,
      user: {
        character: { name: "", id: 0},
        sationName: "",
      },

      get currentView() {
        if (!this.loggedIn) {
          return 'login';
        }

        let path = window.location.pathname;

        if (path === '/')
          return 'index';

        if (path.startsWith('/browse'))
          return 'browse';

        if (path.startsWith('/groups/'))
          return 'groupDetails';

        if (path.startsWith('/history'))
          return 'history';

        if (path.startsWith('/orders'))
          return 'orders';

        if (path.startsWith('/search'))
          return 'search';

        if (path.startsWith('/settings'))
          return 'settings';

        if (path.startsWith('/transactions'))
          return 'transactions';

        if (path.startsWith('/types/'))
          return 'typeDetails';

        return 'notFound';
      },

      initialize() {
        currentUser.then(user => {
          this.user = user;
          this.loggedIn = true;
          return user;
        })
        .catch(() => {})
        .then(() => {
          const url = '/js/'+this.currentView+'.js';
          return retrieve(url, 'error fetching viewData', { raw: true });
        })
        .then(resp => resp.blob())
        .then(blob => {
          const script = document.createElement('script'),
                src = URL.createObjectURL(blob);
          script.src = src;
          document.body.appendChild(script);
        })
        .then(() => {
          const url = '/views/'+this.currentView+'.html';
          return retrieve(url, 'error fetching view', { raw: true });
        })
        .then(resp => resp.text())
        .then(html => {
          const parser = new DOMParser();
          const elt = parser.parseFromString(html, 'text/html').querySelector('main');
          const slot = document.querySelector('main');
          slot.parentNode.replaceChild(elt, slot);
        });
      },
    }
  }

  return result;
})(document, window);
