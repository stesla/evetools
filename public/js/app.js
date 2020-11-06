var evetools = {}

evetools.globalState = function() {
  return {
    avatarMenuOpen: false,
    loggedIn: false,
    user: null,

    get currentView() {
      if (!this.loggedIn) {
        return 'login';
      }

      let path = window.location.pathname;
      
      if (path.startsWith('/type/')) {
        return 'typeDetails';
      } else {
        return 'search';
      }
    },

    fetchCurrentUser() {
      fetch('/api/v1/currentUser').
        then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching current user');
          }
          return resp.json();
        }).
        then(user => {
          this.user = user
          this.user.avatarURL = 'https://images.evetech.net/characters/' + user.characterID + '/portrait?tenant=tranquility&size=128';
          this.loggedIn = true;
        }).catch(() => {});
    },

    handleEscape(e) {
      if (e.key === 'Esc' || e.key === 'Escape') {
        this.avatarMenuOpen = false;
      }
    }
  }
}

evetools.marketTypes = function() {
  const urlParams = new URLSearchParams(window.location.search);
  return {
    filter: urlParams.get('q'),
    marketTypes: [],

    fetchMarketTypes() {
      fetch('/api/v1/typeSearch/' + this.filter).
        then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching market types');
          }
          return resp.json();
        }).
        then(types => {
          this.marketTypes = types;
        });
    },

    href(id) {
      return '/type/' + id;
    },

    initMarketTypes() {
      if (this.filter)
        this.fetchMarketTypes();
    },

    handleSearch(e) {
      e.preventDefault();
      window.location = '?q=' + this.filter;
    },
  }
}

evetools.typeInfo = function() {
  let typeRE = new RegExp("/type/(.*)");
  let match = window.location.pathname.match(typeRE);
  let typeID = match[1];

  return {
    type: null,
    marketInfo: null,

    fetchTypeDetails() {
      fetch('/api/v1/types/' + typeID).
        then(resp => {
          if(!resp.ok) {
            throw new Error('error fetching type info');
          }
          return resp.json();
        }).
        then(type => {
          this.type = type;
        });

      fetch('/api/v1/types/' + typeID + '/marketInfo').
        then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching market info');
          }
          return resp.json();
        }).
        then(result => {
          this.marketInfo = result;
        });
    },
  }
}

window.imgURL = function(type) {
  if (type) {
    return 'https://images.evetech.net/types/' + type.id + '/icon?tenant=tranquility&size=128';
  } else {
    return undefined;
  }
}

