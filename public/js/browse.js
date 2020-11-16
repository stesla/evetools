viewData = (function(window, document, undefined) {
  var marketGroups = retrieve('/data/marketGroups.json', 'error fetching sde market groups'); 

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
      marketGroups.then(data => {
        this.data = data
      });
      document.title += " - Find Items"
    }
  }
})(window, document, undefined);
