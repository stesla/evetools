viewData = (function(window, document, undefined) {
  const urlParams = new URLSearchParams(window.location.search);
  return {
    favorites: [],
    filter: urlParams.get('q'),
    marketTypes: [],

    fetchData() {
      evetools.sdeTypes().then(types => {
        ids = Object.values(types).filter(t => {
          let filter = this.filter.toLowerCase();
          return t.name.toLowerCase().includes(filter);
        }).map(t => t.id);
        this.marketTypes = ids.map(id => types[''+id]);
      });
    },

    handleSearch(e) {
      e.preventDefault();
      window.handleSearch(this.filter);
    },

    initialize() {
      evetools.currentUser.then(user => {
        this.favorites = user.favorites;
      });
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
