viewData = (function(window, document, undefined) {
  var characters = retrieve('/api/v1/user/characters', 'error fetching characters');

  return {
    data: undefined,
    characters: {},
    editingStation: false,
    favorites: [],
    station: { name: "" },
    stationName: "",
    stations: {},
    walletBalance: 0,
    buyTotal: 0,
    sellTotal: 0,

    initialize() {
      document.title += " - Dashboard"
      evetools.currentUser
      .then(user => {
        this.user = user;
        this.walletBalance = user.wallet_balance;
      })
      .then(() => {
        return retrieve('/api/v1/user/orders', 'error fetching market orders');
      })
      .then(orders => {
        this.buyTotal = orders.buy.reduce((a, o) => a + o.escrow, 0);
        this.sellTotal = orders.sell.reduce((a, x) => a + x.volume_remain * x.price, 0);
      });

      characters.then(chars => {
        this.characters = chars;
      });

      evetools.sdeStations().then(stations => {
        this.station = stations[''+this.user.station_id];
      });

      evetools.sdeTypes().then(types => {
        this.favorites = this.user.favorites.map(id => {
          let type = types[""+id];
          type.favorite = true;
          return type;
        }).sort(byName);
      });
    },

    fetchStations() {
      if (this.stationName.length < 3) {
        return;
      }
      evetools.sdeStations().then(stations => {
        this.stations = Object.values(stations).filter(s => {
          return s.name.toLowerCase().includes(this.stationName.toLowerCase());
        }).reduce((m, s) => {
          m[s.name] = s;
          return m;
        }, {});
      });
    },

    get characterList() {
      return Object.values(this.characters).sort(byName);
    },

    get stationList() {
      return Object.values(this.stations);
    },

    toggleFavorite(type) {
      let val = !type.favorite
      setFavorite(type.id, val)
      .then(() => {
        type.favorite = val
      });
    },

    saveStation() {
      if (this.stationName === "") {
        this.editingStation = false;
        return;
      }
      evetools.sdeStations().then(stations => {
        let station = Object.values(stations).find(s => s.name == this.stationName);
        this.station = station;
        return retrieve('/api/v1/user/station', 'error saving station', {
          raw: true,
          method: 'PUT',
          body: JSON.stringify(station),
        });
      })
      .then(() => {
        this.stationName = "";
        this.editingStation = false;
      });
    },
  }
})(window, document, undefined);
