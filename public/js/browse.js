viewData = (function(window, document, undefined) {
  var data = retrieve('/api/v1/view/browse', 'error fetching sde market groups');

  return {
    filter: "",
    groups: [],

    handleSearch(e) {
      e.preventDefault();
      window.handleSearch(this.filter);
    },

    initialize() {
      document.title += " - Find Items"
      data.then(data => {
        this.groups = data.groups.sort(byName);
      });
    }
  }
})(window, document, undefined);
