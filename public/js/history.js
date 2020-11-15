viewData = (function(window, document, undefined) {
  return {
    initialize() {
      document.title += ' - Market Order History'
      evetools.sdeTypes().then(types => {
        this.types = types;
        return evetools.sdeStations();
      })
      .then(stations => {
        this.stations = stations;
        return retrieve('/api/v1/user/history?days=30', 'error fetching history');
      })
      .then(data => {
        this.orders = {
          buy: data.buy.map(o => setOrderFields(o, this.types, this.stations)),
          sell: data.sell.map(o => setOrderFields(o, this.types, this.stations)),
        };
      });
    },

    get sections() {
      return [
        {
          name: "Sell Orders",
          orders: this.orders && this.orders.sell.sort(byOrderID).reverse()
        },
        {
          name: "Buy Orders",
          orders: this.orders && this.orders.buy.sort(byOrderID).reverse()
        },
      ];
    }
  }
})(window, document, undefined);
