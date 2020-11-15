viewData = (function(window, document, undefined) {
  let typeRE = new RegExp("/types/(.*)");
  let match = window.location.pathname.match(typeRE);

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

      evetools.currentUser.then(user => {
        this.user = user;
        return evetools.sdeStations();
      })
      .then(stations => {
        this.station = stations[''+this.user.station_id];
        return evetools.sdeSystems();
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

      evetools.sdeTypes().then(types => {
        this.type = types[''+this.typeID];
        document.title += ' - ' + this.type.name;
        return evetools.sdeMarketGroups()
      })
      .then(marketGroups => {
        this.marketGroups = marketGroups
        this.group = marketGroups.groups[''+this.type.market_group_id];
      });
    }
  }
})(window, document);

