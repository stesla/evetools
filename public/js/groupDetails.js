viewData = (function(window, document, undefined) {
  let typeRE = new RegExp("/groups/(.*)");
  let match = window.location.pathname.match(typeRE);
  let groupID = match[1];

  var data = retrieve('/api/v1/view/groupDetails/' + groupID, 'error fetching view data');

  return {
    group: { name: "" },
    groupID: groupID,
    marketGroups: { root: [] },
    types: {},
    favorites: [],
    filter: "",
    parent: { name: "" },

    get children() {
      if (!this.group.groups && !this.group.types)
        return [];

      if (this.group.groups.length > 0) {
        return this.group.groups.map(g => {
          g.isGroup = true;
          return g
        }).sort(byName);
      } else if (this.group.types.length > 0) {
        return this.group.types.map(t => {
          t.isType = true;
          return t;
        }).sort(byName);
      }
    },

    initialize() {
      data.then(data => {
        this.favorites = data.favorites;
        group = data.group;
        document.title += " - " + group.name;
        group.groups = data.groups;
        group.types = data.types;
        this.group = group;
        this.parent = data.parent;
      });
    },

    isFavorite(typeID) {
      return this.favorites.find(id => id === typeID)
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
