viewData = (function(window, document, undefined) {
  return {
    initialize() {
      document.title += ' - Market Orders'
      evetools.sdeTypes().then(types => {
        this.types = types;
        return evetools.sdeStations();
      })
      .then(stations => {
        this.stations = stations;
        return retrieve('/api/v1/user/orders', 'error fetching orders');
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
          orders: this.orders && this.orders.sell.sort(byTypeName),
        },
        {
         name: "Buy Orders",
          orders: this.orders && this.orders.buy.sort(byTypeName),
        },
      ]
    }
  }
})(window, document, undefined);
