viewData = (function(window, document, undefined) {
  return {
    data: { root: [] },
    filter: "",

    handleSearch(e) {
      e.preventDefault();
      window.handleSearch(this.filter);
    },

    get groups() {
      return this.data.root.map(id =>
        this.data.groups[''+id]
      ).sort(byName);
    },

    initialize() {
      evetools.sdeMarketGroups().then(data => {
        this.data = data
      });
      document.title += " - Find Items"
    }
  }
})(window, document, undefined);
