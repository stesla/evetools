viewData = (function(window, document, undefined) {
  var types = retrieve('/data/types.json', 'error fetching sde types');
  var stations = retrieve('/data/stations.json', 'error fetching sde stations');

  function byOrderID(a, b) {
    return a.order_id < b.order_id ? -1 : 1;
  }

  function descendingByOrderID(a, b) {
    return -byOrderID(a, b)
  }

  let path = window.location.pathname;
  let isCurrent = path.startsWith('/orders');
 
  return {
    initialize() {
     document.title += isCurrent ? ' - Market Orders' : ' - Market Order History';

      types.then(types => {
        this.types = types;
        return stations;
      })
      .then(stations => {
        this.stations = stations;
        if (isCurrent) {
          return retrieve('/api/v1/user/orders', 'error fetching orders');
        } else {
          return retrieve('/api/v1/user/history?days=30', 'error fetching history');
        }
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
          orders: this.orders && this.orders.sell.sort(isCurrent ? byOrderID : descendingByOrderID),
        },
        {
         name: "Buy Orders",
          orders: this.orders && this.orders.buy.sort(isCurrent ? byOrderID : descendingByOrderID),
        },
      ]
    }
  }
})(window, document, undefined);
