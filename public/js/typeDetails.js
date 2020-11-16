viewData = (function(window, document, undefined) {
  let typeRE = new RegExp("/types/(.*)");
  let match = window.location.pathname.match(typeRE);

  var currentUser = window.retrieve('/api/v1/user/current', 'error fetching current user');
  var types = retrieve('/data/types.json', 'error fetching sde types');
  var marketGroups = retrieve('/data/marketGroups.json', 'error fetching sde market groups'); 
  var stations = retrieve('/data/stations.json', 'error fetching sde stations');
  var systems = retrieve('/data/systems.json', 'error fetching sde systems');

  return {
    group: {},
    type: undefined,
    typeID: match[1],
    info: undefined,
    favorite: false,
    station: undefined,
    system: { name: "" },

    toggleFavorite() {
      setFavorite(this.typeID, !this.favorite)
      .then(obj => {
        this.favorite = obj.favorite;
      });
    },

    get parentGroups() {
      const arr = [];
      var g = this.group
      while (g.parent_id) {
        g = this.marketGroups.groups[g.parent_id];
        arr.unshift(g);
      }
      return arr
    },

    initialize() {
      const observer = new MutationObserver(() => {
        let div = document.getElementById("chart");
        if (div) {
          observer.disconnect();
          renderChart(this.info.history, 400, div.clientWidth);
        }
      });

      observer.observe(document.querySelector('main'), { childList: true, subtree: true });

      currentUser.then(user => {
        this.user = user;
        return stations;
      })
      .then(stations => {
        this.station = stations[''+this.user.station_id];
        return systems;
      })
      .then(systems => {
        this.system = systems[''+this.station.system_id]
      })
      .then(() => {
        const params = new URLSearchParams();
        params.set("location_id", this.station.id);
        params.set("region_id", this.system.region_id);
        const url = '/api/v1/types/' + this.typeID + '?' + params.toString();
        return retrieve(url);
      })
      .then(obj => {
        obj.history = obj.history.map(d => {
          return {
            date: Date.parse(d.date),
            average: +d.average,
          }
        });
        this.info = obj;
        this.favorite = obj.favorite;
      });

      types.then(types => {
        this.type = types[''+this.typeID];
        document.title += ' - ' + this.type.name;
        return marketGroups
      })
      .then(marketGroups => {
        this.marketGroups = marketGroups
        this.group = marketGroups.groups[''+this.type.market_group_id];
      });
    }
  }
})(window, document);

