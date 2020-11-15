viewData = (function(window, document, undefined) {
  const urlParams = new URLSearchParams(window.location.search);
  return {
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
      document.title += ' - Search for "' + this.filter + '"';
      if (this.filter) this.fetchData();
    },
  }
})(window, document, undefined);
