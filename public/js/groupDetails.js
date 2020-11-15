viewData = (function(window, document, undefined) {
  let typeRE = new RegExp("/groups/(.*)");
  let match = window.location.pathname.match(typeRE);

  return {
    group: { name: "", groups: [] },
    groupID: match[1],
    marketGroups: { root: [] },
    types: {},
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
      evetools.sdeMarketGroups().then(data => {
        this.marketGroups = data;
        this.group = data.groups[''+this.groupID];
        this.parent = data.groups[''+this.group.parent_id];
        document.title += " - " + this.group.name;
      });

      evetools.sdeTypes().then(types => {
        this.types = types;
      });
    },
  }
})(window, document, undefined);
