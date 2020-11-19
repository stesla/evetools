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
      retrieve('/api/v1/view/search?' + params.toString())
      .then(data => {
        this.favorites = data.favorites;
        this.marketTypes = data.types.sort(byName);
      });
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

    toggleFavorite(type) {
      let val = !this.isFavorite(type.id);
      setFavorite(type.id, val)
      .then(() => {
        if (val) {
          this.favorites.push(type.id);
        } else {
          this.favorites = this.favorites.filter(x => x !== type.id);
        }
      });
    },
  }
})(window, document, undefined);
