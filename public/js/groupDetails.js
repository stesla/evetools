viewData = (function(window, document, undefined) {
  let typeRE = new RegExp("/groups/(.*)");
  let match = window.location.pathname.match(typeRE);

  var currentUser = window.retrieve('/api/v1/user/current', 'error fetching current user');
  var marketGroups = retrieve('/data/marketGroups.json', 'error fetching sde market groups'); 
  var types = retrieve('/data/types.json', 'error fetching sde types');


  return {
    group: { name: "", groups: [] },
    groupID: match[1],
    marketGroups: { root: [] },
    types: {},
    favorites: [],
    filter: "",
    parent: { name: "" },

    get children() {
      if (!this.group || Object.keys(this.types).length == 0)
        return [];

      if (this.group.groups) {
        return this.group.groups.map(id => {
          let g = this.marketGroups.groups[''+id];
          g.isGroup = true;
          return g
        }).sort(byName);
      } else if (this.group.types) {
        return this.group.types.map(id => {
          let t = this.types[''+id];
          t.isType = true;
          return t;
        }).sort(byName);
      }
    },

    initialize() {
      currentUser.then(user => {
        this.favorites = user.favorites;
      });

      marketGroups.then(data => {
        this.marketGroups = data;
        this.group = data.groups[''+this.groupID];
        this.parent = data.groups[''+this.group.parent_id];
        document.title += " - " + this.group.name;
      });

      types.then(types => {
        this.types = types;
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
