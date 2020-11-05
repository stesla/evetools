var evetools = {}

evetools.globalState = function() {
  return {
    avatarMenuOpen: false,
    loggedIn: false,
    user: null,

    get currentView() {
      if (!this.loggedIn) {
        return 'login'
      } else {
        return 'home'
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
      fetch('/api/v1/marketTypes/' + this.filter).
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

    initMarketTypes() {
      if (this.filter)
        this.fetchMarketTypes();
    },

    imgURL(type) {
      return 'https://images.evetech.net/types/' + type.id + '/icon?tenant=tranquility&size=128'
    }
  }
}
