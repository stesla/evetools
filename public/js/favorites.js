favorites = (function(window, document, undefined) {
  window.deleteFavorites = function() {
    retrieve('/api/v1/user/favorites', 'error deleting favorites', {
      raw: true,
      method: 'DELETE',
    })
    .then(() => {
      location.reload(true);
    });
  };

  return function(faves){
    return {
      favorites: faves,

      isFavorite(typeID) {
        return this.favorites.find(id => id === typeID);
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
  }
})(window, document, undefined);
