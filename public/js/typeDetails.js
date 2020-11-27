viewData = (function(window, document, undefined) {
  return {
    initialize(typeID, isFavorite) {
      this.typeID = typeID
      this.favorite = isFavorite
    },

    toggleFavorite() {
      setFavorite(this.typeID, !this.favorite)
      .then(obj => {
        this.favorite = obj.favorite;
      });
    },
  }
})(window, document);

