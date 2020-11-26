viewData = (function(window, document, undefined) {
  const urlParams = new URLSearchParams(window.location.search);
  const filter = urlParams.get('q');

  if (!filter || "" === filter)
    window.location = '/browse';

  return {
    favorites: [],
    filter: filter,
    marketTypes: [],

    fetchData() {
      const params = new URLSearchParams();
      params.set("q", this.filter);
      retrieve('/api/v1/user/favorites')
      .then(data => { this.favorites = data });
    },

    handleSearch(e) {
      e.preventDefault();
      window.handleSearch(this.filter);
    },

    initialize() {
      document.title += ' - Search for "' + this.filter + '"';
      if (this.filter) this.fetchData();
    },

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
})(window, document, undefined);
