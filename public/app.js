var evetools = {}

evetools.globalState = function() {
  return {
    burgerOpen: false,
    loggedIn: false,
    user: null,

    fetchCurrentUser() {
      fetch('/api/v1/currentUser').
        then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching current user');
          }
          return resp.json();
        }).
        then(user => {
          console.log(user);
          this.user = user
          this.user.avatarURL = 'https://imageserver.eveonline.com/Character/' + user.characterID + '_32.jpg';
          this.loggedIn = true;
        }).catch(() => {});
    },
  }
}

