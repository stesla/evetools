viewData = (function(window, document, undefined) {
  let typeRE = new RegExp("/groups/(.*)");
  let match = window.location.pathname.match(typeRE);
  let groupID = match[1];

  return {
    favorites: [],

    initialize() {
      retrieve('/api/v1/user/favorites')
      .then(data => { this.favorites = data });
    },

    isFavorite(typeID) {
      return this.favorites.find(id => id === typeID)
    },

    toggleFavorite(typeID) {
      let val = !this.isFavorite(typeID);
      setFavorite(typeID, val)
      .then(() => {
        if (val) {
          this.favorites.push(typeID);
        } else {
          this.favorites = this.favorites.filter(x => x !== typeID);
        }
      });
    },
  }
})(window, document, undefined);
