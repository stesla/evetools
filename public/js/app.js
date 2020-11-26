evetools = (function(document, window, undefined) {
  var result = {};

  result.globalState = function() {
    return {
      avatarMenuOpen: false,
      marketMenuOpen: false,
      loggedIn: false,
      navOpen: false,
      character: { name: "", id: 0 },

      initialize() {
        retrieve('/api/v1/verify', 'error verifying auth')
        .then(user => {
          this.character = {
            id: user.character_id,
            name: user.character_name,
          }
          this.loggedIn = true;
          return user;
        })
        .catch(() => {});
      },
    }
  }

  return result;
})(document, window);
